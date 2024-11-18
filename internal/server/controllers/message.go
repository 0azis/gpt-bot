package controllers

import (
	"errors"
	"gpt-bot/internal/api"
	"gpt-bot/internal/db"
	"gpt-bot/internal/db/domain"
	"gpt-bot/utils"
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
)

const uploadsRoute = "localhost:5000/api/v1/uploads/"

type messageControllers interface {
	NewMessage(c echo.Context) error
	// NewMessageToChat(c echo.Context) error
	GetMessages(c echo.Context) error

	modelAnswer(store db.Store, chat domain.Chat, msg []domain.Message) (string, error)
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
		// chatDb, err := m.store.Chat.GetByID(id)
		// if err != nil {
		// 	slog.Error(err.Error())
		// 	return c.JSON(500, nil)
		// }
		chat = newChat
		isNewChat = true
	} else {
		chatDb, err := m.store.Chat.GetByID(paramInt)
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(500, nil)
		}
		chat = chatDb
		isNewChat = false
	}

	var msgCredentials domain.Message
	err = c.Bind(&msgCredentials)
	if err != nil || !msgCredentials.Valid() {
		return c.JSON(400, nil)
	}

	balance, err := m.store.User.GetBalance(jwtUserID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}
	if balance == 0 {
		return c.JSON(403, nil)
	}

	file, _ := c.FormFile("file")
	if file != nil {
		name, err := utils.SaveImage(file, m.savePath)
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(500, nil)
		}

		imageMsg := domain.NewUserImageMessage(chat.ID, uploadsRoute+name)
		err = m.store.Message.Create(imageMsg)
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(500, nil)
		}
	}

	userMsg := domain.NewUserTextMessage(chat.ID, msgCredentials.Content)
	err = m.store.Message.Create(userMsg)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	go func() {
		title, _ := m.api.OpenAI.GenerateTopicForChat(userMsg)
		m.store.Chat.UpdateTitle(chat.ID, title)
	}()

	messages, err := m.store.Message.GetByChat(jwtUserID, chat.ID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	answer, err := m.modelAnswer(m.store, chat, messages)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	err = m.store.User.ReduceBalance(jwtUserID, domain.PriceOfMessage)
	if err != nil {
		slog.Error(err.Error())
		c.JSON(500, nil)
	}

	if isNewChat {
		return c.JSON(200, chat.ID)
	} else {
		return c.JSON(200, answer)
	}
}

// func (m message) NewMessageToChat(c echo.Context) error {
// 	jwtUserID := utils.ExtractUserID(c)
// 	value := c.Param("id")
// 	chatID, err := strconv.Atoi(value)
// 	if err != nil {
// 		return c.JSON(400, nil)
// 	}

// 	var msgCredentials domain.Message
// 	err = c.Bind(&msgCredentials)
// 	if err != nil || !msgCredentials.Valid() {
// 		return c.JSON(400, nil)
// 	}

// 	balance, err := m.store.User.GetBalance(jwtUserID)
// 	if err != nil {
// 		slog.Error(err.Error())
// 		return c.JSON(500, nil)
// 	}
// 	if balance == 0 {
// 		return c.JSON(403, nil)
// 	}

// 	chat, err := m.store.Chat.GetByID(chatID)
// 	if err != nil {
// 		slog.Error(err.Error())
// 		return c.JSON(500, nil)
// 	}

// 	userMsg := domain.NewUserTextMessage(chat.ID, msgCredentials.Content)
// 	err = m.store.Message.Create(userMsg)
// 	if err != nil {
// 		slog.Error(err.Error())
// 		return c.JSON(500, nil)
// 	}

// 	messages, err := m.store.Message.GetByChat(jwtUserID, chat.ID)
// 	if err != nil {
// 		slog.Error(err.Error())
// 		return c.JSON(500, nil)
// 	}

// 	err = m.store.User.ReduceBalance(jwtUserID, domain.PriceOfMessage)
// 	if err != nil {
// 		slog.Error(err.Error())
// 		return c.JSON(500, nil)
// 	}

// 	answer, err := m.modelAnswer(m.store, chat, messages)
// 	if err != nil {
// 		slog.Error(err.Error())
// 		return c.JSON(500, nil)
// 	}

// 	return c.JSON(200, answer)
// }

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

func (m message) modelAnswer(store db.Store, chat domain.Chat, msg []domain.Message) (string, error) {
	switch chat.Type {
	case domain.ChatImage:
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
