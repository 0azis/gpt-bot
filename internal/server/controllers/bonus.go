package controllers

import (
	"gpt-bot/internal/db"
	"gpt-bot/internal/db/domain"
	"gpt-bot/tgbot"
	"gpt-bot/utils"
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
)

type bonusControllers interface {
	Create(c echo.Context) error
	GetAll(c echo.Context) error
	Delete(c echo.Context) error
	GetAward(c echo.Context) error
}

type bonus struct {
	store db.Store
	tg    tgbot.BotInterface
}

func (b bonus) Create(c echo.Context) error {
	var bonus domain.Bonus
	err := c.Bind(&bonus)
	if err != nil || !bonus.Valid() {
		return c.JSON(400, nil)
	}

	err = b.store.Bonus.Create(bonus)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	return c.JSON(201, nil)
}

func (b bonus) GetAll(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)
	value := c.QueryParam("completed")
	isCompleted, err := strconv.ParseBool(value)
	if err != nil {
		return c.JSON(400, nil)
	}

	switch isCompleted {
	case true:
		bonuses, err := b.store.Bonus.GetCompleted(jwtUserID)
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(500, nil)
		}

		for _, bonus := range bonuses {
			if b.tg.IsUserMember(bonus.Channel.Name, jwtUserID) {
				err := b.store.Bonus.MakeCompleted(bonus.ID, jwtUserID)
				if err != nil {
					slog.Error(err.Error())
					return c.JSON(500, nil)
				}
			}
		}

		return c.JSON(200, bonuses)

	case false:
		bonuses, err := b.store.Bonus.GetUncompleted(jwtUserID)
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(500, nil)
		}

		return c.JSON(200, bonuses)
	default:
		return c.JSON(400, nil)
	}
}

func (b bonus) Delete(c echo.Context) error {
	value := c.Param("id")
	bonusID, err := strconv.Atoi(value)
	if err != nil {
		return c.JSON(400, nil)
	}

	err = b.store.Bonus.Delete(bonusID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	return c.JSON(200, nil)
}

func (b bonus) GetAward(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)

	value := c.Param("id")
	bonusID, err := strconv.Atoi(value)
	if err != nil {
		return c.JSON(400, nil)
	}

	award, err := b.store.Bonus.GetAward(bonusID, jwtUserID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	err = b.store.User.RaiseBalance(jwtUserID, award)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	return c.JSON(200, nil)
}

func NewBonusControllers(store db.Store, tg tgbot.BotInterface) bonusControllers {
	return bonus{
		store, tg,
	}
}
