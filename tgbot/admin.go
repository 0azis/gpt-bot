package tgbot

import (
	"context"
	"fmt"
	"gpt-bot/internal/db/domain"
	"log/slog"
	"strings"

	"github.com/0azis/bot"
	"github.com/0azis/bot/models"
	"github.com/go-telegram/fsm"
)

const (
	userStatistics     = "btn_1"
	requestsStatistics = "btn_2"
	giveSubscription   = "btn_3"
	giveDiamonds       = "btn_4"
	createBonus        = "btn_5"
	deleteBonus        = "btn_6"
	bonusStatistics    = "btn_7"

	stateDefault fsm.StateID = "stateDefault"
	// state
	diamondStateAskUserID         fsm.StateID = "diamond_ask_user_id"
	diamondStateAskDiamondsAmount fsm.StateID = "diamond_ask_diamonds_amount"
	diamondStateFinish            fsm.StateID = "diamond_finish"
	// state 2
	subscriptionStateDefault   fsm.StateID = "subscription_default"
	subscriptionStateAskUserID fsm.StateID = "subscription_ask_user_id"
	subscriptionStateAskName   fsm.StateID = "subscription_ask_name"
	subscriptionStateFinish    fsm.StateID = "subscription_finish"
	// state 3
	bonusStateDefault    fsm.StateID = "bonus_default"
	bonusStateDelete     fsm.StateID = "bonus_delete"
	bonusStateAskChannel fsm.StateID = "bonus_ask_channel"
	bonusStateAskAward   fsm.StateID = "bonus_ask_award"
	bonusStateFinish     fsm.StateID = "bonus_finish"
)

type diamondsData struct {
	userID int
	amount int
}

type subscriptionData struct {
	userID int
	name   string
}

type bonusData struct {
	channel_name string
	award        int
}

var diamondsScheme diamondsData
var subscriptionScheme subscriptionData
var bonusScheme bonusData

func (tb tgBot) adminHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	msgSlice := strings.Split(update.Message.Text, " ")
	if len(msgSlice) == 1 {
		tb.informUser(ctx, update.Message.From.ID, adminPanelEnterPassword)
		return
	}
	password := msgSlice[1]
	if tb.telegram.GetAdminPassword() != password {
		tb.informUser(ctx, update.Message.From.ID, adminPanelWrondPassword)
		return
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Статистика запросов", CallbackData: requestsStatistics},
				{Text: "Статистика бонусов", CallbackData: bonusStatistics},
			},
			{{Text: "Статистика пользователей", CallbackData: userStatistics}},
			{
				{Text: "Выдать подписку", CallbackData: giveSubscription},
				{Text: "Выдать алмазы", CallbackData: giveDiamonds},
			},
			{
				{Text: "Создать бонус", CallbackData: createBonus},
				{Text: "Удалить бонус", CallbackData: deleteBonus},
			},
		},
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.From.ID,
		Text:        "<b>Добро пожаловать в админ панель!</b>",
		ReplyMarkup: kb,
		ParseMode:   models.ParseModeHTML,
	})
}

func (tb tgBot) callbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	switch update.CallbackQuery.Data {
	case "btn_1":
		tb.usersStatisticsCallback(ctx, b, update)
	case "btn_2":
		tb.requestsStatisticsCallback(ctx, b, update)
	case "btn_3":
		tb.f.Transition(update.CallbackQuery.From.ID, subscriptionStateAskUserID, update.CallbackQuery.From.ID)
	case "btn_4":
		tb.f.Transition(update.CallbackQuery.From.ID, diamondStateAskUserID, update.CallbackQuery.From.ID)
	case "btn_5":
		tb.f.Transition(update.CallbackQuery.From.ID, bonusStateAskChannel, update.CallbackQuery.From.ID)
	case "btn_6":
		tb.f.Transition(update.CallbackQuery.From.ID, bonusStateDelete, update.CallbackQuery.From.ID)
	case "btn_7":
		tb.bonusesStatisticsCallback(ctx, b, update)
	default:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.CallbackQuery.From.ID,
			Text:   "Ошибка при выборе функции",
		})
	}
}

func (tb tgBot) callbackUserID(f *fsm.FSM, args ...any) {
	chatID := args[0]

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Введите айди пользователя",
	})
}

func (tb tgBot) callbackDiamondsAmount(f *fsm.FSM, args ...any) {
	chatID := args[0]

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Введите количество алмазов",
	})
}

func (tb tgBot) callbackSubscriptionName(f *fsm.FSM, args ...any) {
	chatID := args[0]

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Введите название подписки",
	})
}

func (tb tgBot) callbackChannelName(f *fsm.FSM, args ...any) {
	chatID := args[0]

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Введите название канала (channelname)",
	})
}

func (tb tgBot) callbackBonusAward(f *fsm.FSM, args ...any) {
	chatID := args[0]

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Введите вознаграждение",
	})
}

func (tb tgBot) usersStatisticsCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	allUsers, err := tb.store.User.CountUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, adminUserStatistics)
		return
	}
	dailyUsers, err := tb.store.User.DailyUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, adminUserStatistics)
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.CallbackQuery.From.ID,
		Text:      fmt.Sprintf("<b>Все пользователи: </b>%d\n<b>Новые пользователи за сегодня: </b>%d", allUsers, dailyUsers),
		ParseMode: models.ParseModeHTML,
	})
}

func (tb tgBot) requestsStatisticsCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	dailyRequests, err := tb.store.Message.RequestsDaily()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, adminRequestsStatistics)
		return
	}
	weeklyRequests, err := tb.store.Message.RequestsWeekly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, adminRequestsStatistics)
		return
	}
	montlyRequests, err := tb.store.Message.RequestsMontly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, adminRequestsStatistics)
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.CallbackQuery.From.ID,
		Text:      fmt.Sprintf("<b>Запросы за сегодня: </b>%d\n<b>Запросы за неделю: </b>%d\n<b>Запросы за месяц </b>%d", dailyRequests, weeklyRequests, montlyRequests),
		ParseMode: models.ParseModeHTML,
	})

}

func (tb tgBot) bonusesStatisticsCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	allBonuses, err := tb.store.Bonus.AllBonuses()
	if err != nil {
	}
	dailyBonuses, err := tb.store.Bonus.DailyBonuses()
	if err != nil {
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.CallbackQuery.From.ID,
		Text:      fmt.Sprintf("<b>Выполненные бонусы за сегодня: </b>%d\n<b>Выполненные бонусы (всего): </b>%d\n", dailyBonuses, allBonuses),
		ParseMode: models.ParseModeHTML,
	})
}

func (tb tgBot) giveSubscription(subscriptionScheme subscriptionData, msg *models.Message) {
	ctx := context.Background()
	var subscription domain.Payment
	subscription.SubscriptionName = subscriptionScheme.name
	subscription.UserID = subscriptionScheme.userID
	subscription.ToReadable()

	err := tb.store.Subscription.Update(subscription.UserID, subscription.SubscriptionName, subscription.End)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(msg.From.ID), internalError)
		return
	}
	diamonds, err := tb.store.Subscription.DailyDiamonds(subscription.SubscriptionName)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(msg.From.ID), internalError)
		return
	}
	err = tb.store.User.FillBalance(subscription.UserID, diamonds)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(msg.From.ID), internalError)
		return
	}
	limits := domain.NewLimits(subscription.UserID, subscription.SubscriptionName)
	err = tb.store.Limits.Update(limits)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(msg.From.ID), internalError)
		return
	}

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID:    msg.From.ID,
		Text:      fmt.Sprintf("Подписка отправлена пользователю: %d", subscription.UserID),
		ParseMode: models.ParseModeHTML,
	})

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID:    subscription.UserID,
		Text:      fmt.Sprintf("Админ выдал вам подписку: <b>%s</b> до <i>%s</i>", subscription.SubscriptionName, subscription.End),
		ParseMode: models.ParseModeHTML,
	})
}

func (tb tgBot) giveDiamonds(diamondsScheme diamondsData, msg *models.Message) {
	err := tb.store.User.RaiseBalance(diamondsScheme.userID, diamondsScheme.amount)
	if err != nil {
		tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: msg.From.ID,
			Text:   "Ошибка при отправке алмазов. Проверьте данные еще раз или попробуйте позже",
		})
		return
	}

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID:    msg.From.ID,
		Text:      fmt.Sprintf("Алмазы отправлены пользователю: <b>%d</b>", diamondsScheme.userID),
		ParseMode: models.ParseModeHTML,
	})

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID:    diamondsScheme.userID,
		Text:      fmt.Sprintf("Админ выдал вам алмазы: <b>%d</b> штук", diamondsScheme.amount),
		ParseMode: models.ParseModeHTML,
	})
}

func (tb tgBot) createBonus(bonusScheme bonusData, msg *models.Message) {
	bonus := domain.Bonus{
		Channel: domain.Channel{
			Name: bonusScheme.channel_name,
		},
		Award: bonusScheme.award,
	}
	err := tb.store.Bonus.Create(bonus)
	if err != nil {
		tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: msg.From.ID,
			Text:   "Ошибка при создании бонуса. Проверьте данные еще раз или попробуйте позже",
		})
		return
	}

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: msg.From.ID,
		Text:   "Бонус создан",
	})

}

func (tb tgBot) deleteBonus(bonusScheme bonusData, msg *models.Message) {
	err := tb.store.Bonus.Delete(bonusScheme.channel_name)
	if err != nil {
		tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID: msg.From.ID,
			Text:   "Ошибка при удалении бонуса. Проверьте данные еще раз или попробуйте позже",
		})
		return
	}
	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: msg.From.ID,
		Text:   "Бонус удален",
	})
}
