package controllers

import (
	"database/sql"
	"errors"
	"gpt-bot/internal/db"
	"gpt-bot/tgbot"
	"gpt-bot/utils"
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
)

type bonusControllers interface {
	GetAll(c echo.Context) error
	GetAward(c echo.Context) error
}

type bonus struct {
	store db.Store
	tg    tgbot.BotInterface
}

func (b bonus) GetAll(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)

	bonuses, err := b.store.Bonus.GetAll(jwtUserID)
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}
	for _, bonus := range bonuses {
		channel, err := b.tg.GetChannelInfo(bonus.Channel.Name)
		if err != nil {
			slog.Error(err.Error())
			return c.JSON(500, nil)
		}
		bonus.Channel = channel
		if b.tg.IsUserMember(channel.Name, jwtUserID) {
			bonus.Completed = true
		}
	}

	return c.JSON(200, bonuses)
}

func (b bonus) GetAward(c echo.Context) error {
	jwtUserID := utils.ExtractUserID(c)

	value := c.Param("id")
	bonusID, err := strconv.Atoi(value)
	if err != nil {
		return c.JSON(400, nil)
	}

	bonus, err := b.store.Bonus.GetOne(bonusID)
	if errors.Is(err, sql.ErrNoRows) {
		return c.JSON(404, nil)
	}
	if err != nil {
		slog.Error(err.Error())
		return c.JSON(500, nil)
	}

	if !b.tg.IsUserMember(bonus.Channel.Name, jwtUserID) {
		return c.JSON(400, nil)
	}
	if bonus.Awarded {
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

	err = b.store.Bonus.MakeAwarded(bonusID, jwtUserID)
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
