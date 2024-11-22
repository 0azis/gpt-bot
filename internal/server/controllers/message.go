package controllers

import (
	"database/sql"
	"errors"
	"gpt-bot/internal/api"
	"gpt-bot/internal/db"
	"gpt-bot/internal/db/domain"
	"gpt-bot/utils"
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
)

const uploadsRoute = "https://mywebai.top/api/v1/uploads/"

type messageControllers interface {
	NewMessage(c echo.Context) error
	GetMessages(c echo.Context) error

	modelAnswer(userID int, store db.Store, chat domain.Chat, msg []domain.Message) (string, error)
}

type message struct {
	api      api.Interface
	store    db.Store
	savePath string
}

func (m message) NewMessage(c echo.Context) error {
	paramStr := c.QueryParam("chat_id")
	if paramStr == "" {
		paramStr = "0"
	}
	paramInt, err := strconv.Atoi(paramStr)
	if err != nil {
		return c.JSON(400, nil)
	}

	jwtUserID := utils.ExtractUserID(c)
	model := c.QueryParam("model")
	if model == "" {
		return c.JSON(400, nil)
	}

	user, err := m.store.User.GetByID(jwtUserID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}
	if !user.IsModelValid(model) {
		return c.JSON(403, nil)
	}

	var message domain.Message
	multipart, err := c.MultipartForm()
	if err != nil {
		return c.JSON(400, nil)
	}
	if d, ok := multipart.Value["data"]; ok {
		if d[0] == "" {
			return c.JSON(400, nil)
		}
		message.Content = d[0]
		// err := utils.BindToJSON(&message, d[0])
		// if err != nil || !message.Valid() {
		// 	return c.JSON(400, nil)
		// }
	}

	var chat domain.Chat
	var isNewChat bool
	if paramInt == 0 {
		newChat := domain.Chat{
			UserID: jwtUserID,
			Model:  model,
		}
		if ok := newChat.SetType(); !ok {
			return c.JSON(400, nil)
		}

		id, err := m.store.Chat.Create(newChat)
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(500, nil)
		}
		newChat.ID = id
		chat = newChat

		go func() {
			title, err := m.api.OpenAI.GenerateTopicForChat(message)
			if err == nil {
				m.store.Chat.UpdateTitle(id, title)
			}
		}()

		isNewChat = true
	} else {
		chatDb, err := m.store.Chat.GetByID(paramInt)
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(404, nil)
		}
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(500, nil)
		}
		chat = chatDb
		isNewChat = false
	}

	if f, ok := multipart.File["file"]; ok {
		fileObj := f[0]
		uuid, err := utils.SaveImage(fileObj, m.savePath)
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(500, nil)
		}
		imageMsg := domain.NewUserImageMessage(chat.ID, uploadsRoute+uuid)
		err = m.store.Message.Create(imageMsg)
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(500, nil)
		}
	}

	balance, err := m.store.User.GetBalance(jwtUserID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}
	if balance == 0 {
		return c.JSON(403, nil)
	}

	modelLimits, err := m.store.Limits.GetLimitsByModel(jwtUserID, model)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}
	if modelLimits == 0 {
		return c.JSON(403, nil)
	}

	err = m.store.Limits.Reduce(jwtUserID, model)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	userMsg := domain.NewUserTextMessage(chat.ID, message.Content)
	err = m.store.Message.Create(userMsg)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	messages, err := m.store.Message.GetByChat(jwtUserID, chat.ID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	answer, err := m.modelAnswer(jwtUserID, m.store, chat, messages)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	if isNewChat {
		return c.JSON(200, chat.ID)
	} else {
		return c.JSON(200, answer)
	}
}

func (m message) GetMessages(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)
	value := c.Param("id")
	chatID, err := strconv.Atoi(value)
	if err != nil {
		return c.JSON(400, nil)
	}

	messages, err := m.store.Message.GetByChat(jwtUserID, chatID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	return c.JSON(200, messages)
}

func (m message) modelAnswer(userID int, store db.Store, chat domain.Chat, msg []domain.Message) (string, error) {
	switch chat.Type {
	case domain.ChatImage:
		err := m.store.User.ReduceBalance(userID, domain.PriceOfImageMessage)
		if err != nil {
			slog.Error(err.Error())
			return "", err
		}
		switch chat.Model {
		case "runware":
			answer, err := m.api.Runware.SendMessage(msg[len(msg)-1].Content)
			if err != nil {
				slog.Error(err.Error())
				return answer, err
			}
			assistantMsg := domain.NewAssistantImageMessage(chat.ID, answer)
			err = store.Message.Create(assistantMsg)
			return answer, err
		case "dall-e-3":
			answer, err := m.api.OpenAI.SendImageMessage(msg[len(msg)-1].Content)
			if err != nil {
				slog.Error(err.Error())
				return answer, err
			}
			assistantMsg := domain.NewAssistantImageMessage(chat.ID, answer)
			err = store.Message.Create(assistantMsg)
			return answer, err
		default:
			return "", errors.New("model error")
		}

	case domain.ChatText:
		err := m.store.User.ReduceBalance(userID, domain.PriceOfTextMessage)
		if err != nil {
			slog.Error(err.Error())
			return "", err
		}
		answer, err := m.api.OpenAI.SendMessage(chat.Model, msg)
		if err != nil {
			slog.Error(err.Error())
			return answer, err
		}
		assistantMsg := domain.NewAssistantTextMessage(chat.ID, answer)
		err = store.Message.Create(assistantMsg)
		return answer, err
	default:
		return "", errors.New("model error")
	}
}

func NewMessageControllers(api api.Interface, store db.Store, savePath string) messageControllers {
	return message{
		api, store, savePath,
	}
}
