package cron

import (
	"gpt-bot/internal/db"
	"log/slog"

	"github.com/robfig/cron/v3"
)

type cronInterface interface {
	Run()

	dailyTask()
	updateBalance() error
	checkSubscription() error
}

type cronManager struct {
	*cron.Cron
	store db.Store
}

func Init(store db.Store) cronManager {
	c := cron.New()
	cronManager := cronManager{
		c, store,
	}
	cronManager.dailyTask()
	return cronManager
}

func (c cronManager) Run() {
	c.Start()
}

func (c cronManager) dailyTask() {
	c.AddFunc("@midnight", func() {
		err := c.checkSubscription()
		if err != nil {
			slog.Error(err.Error())
			return
		}

		err = c.updateBalance()
		if err != nil {
			slog.Error(err.Error())
			return
		}
	})
}

func (c cronManager) updateBalance() error {
	users, err := c.store.User.GetAll()
	if err != nil {
		return err
	}
	for _, user := range users {
		diamonds, err := c.store.Subscription.DailyDiamonds(user.Subscription.Name)
		if err != nil {
			return err
		}
		err = c.store.User.FillBalance(user.ID, diamonds)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c cronManager) checkSubscription() error {
	err := c.store.Subscription.EndTime()
	return err
}
