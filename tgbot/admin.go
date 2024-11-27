package tgbot

import (
	"context"
	"fmt"
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

	// state
	stateDefault fsm.StateID = "default"
	// stateStart             fsm.StateID = "start"
	stateAskUserID         fsm.StateID = "ask_user_id"
	stateAskDiamondsAmount fsm.StateID = "ask_diamonds_amount"
	stateFinish            fsm.StateID = "finish"
)

type diamondsScheme struct {
	userID int
	amount int
}

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
				{Text: "Статистика пользователей", CallbackData: userStatistics},
				{Text: "Статистика запросов", CallbackData: requestsStatistics},
				{Text: "Выдать подписку пользователю", CallbackData: giveSubscription},
				{Text: "Выдать алмазы пользователю", CallbackData: giveDiamonds},
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
		tb.giveSubscriptionCallback(ctx, b, update)
	case "btn_4":
		tb.f.Transition(update.CallbackQuery.From.ID, stateAskUserID, update.CallbackQuery.From.ID)
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
	weeklyRequests := 0
	// weeklyRequests, err := tb.store.Message.RequestsWeekly()
	// if err != nil {
	// 	slog.Error(err.Error())
	// 	tb.informUser(ctx, update.CallbackQuery.From.ID, adminRequestsStatistics)
	// 	return
	// }
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

func (tb tgBot) giveSubscriptionCallback(ctx context.Context, b *bot.Bot, update *models.Update) {

}

func (tb tgBot) giveDiamonds(userID, amount int) {
	err := tb.store.User.RaiseBalance(userID, amount)
	if err != nil {

	}
	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: userID,
		Text:   "Алмазы успешно зачислены",
	})
}
