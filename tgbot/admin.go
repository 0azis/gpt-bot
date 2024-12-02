package tgbot

import (
	"context"
	"fmt"
	"gpt-bot/internal/db/domain"
	"log/slog"
	"strconv"
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
	createReferral     = "btn_8"
	deleteReferral     = "btn_9"
	getReferrals       = "btn_10"

	stateDefault fsm.StateID = "stateDefault"
	// state
	diamondStateAskUserID         fsm.StateID = "diamond_ask_user_id"
	diamondStateAskDiamondsAmount fsm.StateID = "diamond_ask_diamonds_amount"
	diamondStateFinish            fsm.StateID = "diamond_finish"
	// state 2
	subscriptionStateAskUserID fsm.StateID = "subscription_ask_user_id"
	subscriptionStateAskName   fsm.StateID = "subscription_ask_name"
	subscriptionStateFinish    fsm.StateID = "subscription_finish"
	// state 3
	bonusStateDelete     fsm.StateID = "bonus_delete"
	bonusStateAskChannel fsm.StateID = "bonus_ask_channel"
	bonusStateAskAward   fsm.StateID = "bonus_ask_award"
	bonusStateFinish     fsm.StateID = "bonus_finish"
	// state 4
	referralStateDelete fsm.StateID = "referral_delete"
)

const (
	menu = "btn_menu"

	statistics        = "btn_1"
	statisticsDaily   = "btn_1_1"
	statisticsWeekly  = "btn_1_2"
	statisticsMonthly = "btn_1_3"
	statisticsAll     = "btn_1_4"
	statisticsBack    = "btn_1_5"

	bonuses               = "btn_2"
	bonusesInfo           = "btn_2_1"
	bonusesCreate         = "btn_2_2"
	bonusesBack           = "btn_2_3"
	bonusesChaneChannel   = "btn_2_5"
	bonusesChangeName     = "btn_2_6"
	bonusesChangeMaxUsers = "btn_2_7"
	bonusesDelete         = "btn_2_8"
	bonuseID              = "id@"

	users     = "btn_3"
	referrals = "btn_4"
	requests  = "btn_5"
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

type referralData struct {
	refID int
}

var diamondsScheme diamondsData
var subscriptionScheme subscriptionData
var bonusScheme bonusData
var referralScheme referralData

func (tb tgBot) adminHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if !tb.store.Admin.CheckID(int(update.Message.From.ID)) {
		tb.informUser(ctx, update.Message.From.ID, "bad id")
		return
	}
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Статистика", CallbackData: statistics},
				{Text: "Бонусы", CallbackData: bonuses},
			},
			{
				{Text: "Пользователи", CallbackData: users},
				{Text: "ССылки", CallbackData: referrals},
			},
			{
				{Text: "Запросы", CallbackData: requests},
			},
		},
	}

	premiumUsers, err := tb.store.User.PremiumUsersCount()
	if err != nil {
	}
	statsDaily, err := tb.store.Stat.Daily()
	if err != nil {
	}
	statsMonthly, err := tb.store.Stat.Monthly()
	if err != nil {
	}
	statsAll, err := tb.store.Stat.All()
	if err != nil {
	}

	bonusesDaily, err := tb.store.Bonus.DailyBonusesCount()
	if err != nil {
	}
	bonusesMonthly, err := tb.store.Bonus.MonthlyBonusesCount()
	if err != nil {
	}
	bonusesAll, err := tb.store.Bonus.AllBonuses()
	if err != nil {
	}

	usersDaily, err := tb.store.User.DailyUsersCount()
	if err != nil {
	}
	usersMonthly, err := tb.store.User.MonthlyUsersCount()
	if err != nil {
	}
	usersAll, err := tb.store.User.AllUsersCount()
	if err != nil {
	}

	usersReferredDaily, err := tb.store.User.DailyUsersReferred()
	if err != nil {
	}
	usersReferredMonthly, err := tb.store.User.MonthlyUsersReferred()
	if err != nil {
	}
	usersReferredAll, err := tb.store.User.AllUsersReferred()
	if err != nil {
	}

	referralsUsersAll, err := tb.store.Referral.AllUsers()
	if err != nil {
	}
	referralUsersDaily, err := tb.store.Referral.DailyUsers()
	if err != nil {
	}
	referralUsersMonthly, err := tb.store.Referral.MonthlyUsers()
	if err != nil {
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.From.ID,
		Text:        fmt.Sprintf("👑 Премиум пользователей: %d\n\n🚀 Запусков: %d | %d | %d\n🎁 Выполнено бонусов: %d | %d | %d\n\n✅ Статистика пользователей\n|-Саморост: %d | %d | %d\n|-Приглашены: %d | %d | %d\n|-Реферальные ссылки: %d | %d | %d\n", premiumUsers, statsDaily, statsMonthly, statsAll, bonusesDaily, bonusesMonthly, bonusesAll, usersDaily, usersMonthly, usersAll, usersReferredDaily, usersReferredMonthly, usersReferredAll, referralUsersDaily, referralUsersMonthly, referralsUsersAll),
		ReplyMarkup: kb,
		ParseMode:   models.ParseModeHTML,
	})
}

func (tb tgBot) adminMenu(ctx context.Context, b *bot.Bot, update *models.Update) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Статистика", CallbackData: statistics},
				{Text: "Бонусы", CallbackData: bonuses},
			},
			{
				{Text: "Пользователи", CallbackData: users},
				{Text: "ССылки", CallbackData: referrals},
			},
			{
				{Text: "Запросы", CallbackData: requests},
			},
		},
	}
	premiumUsers, err := tb.store.User.PremiumUsersCount()
	if err != nil {
	}
	statsDaily, err := tb.store.Stat.Daily()
	if err != nil {
	}
	statsMonthly, err := tb.store.Stat.Monthly()
	if err != nil {
	}
	statsAll, err := tb.store.Stat.All()
	if err != nil {
	}

	bonusesDaily, err := tb.store.Bonus.DailyBonusesCount()
	if err != nil {
	}
	bonusesMonthly, err := tb.store.Bonus.MonthlyBonusesCount()
	if err != nil {
	}
	bonusesAll, err := tb.store.Bonus.AllBonuses()
	if err != nil {
	}

	usersDaily, err := tb.store.User.DailyUsersCount()
	if err != nil {
	}
	usersMonthly, err := tb.store.User.MonthlyUsersCount()
	if err != nil {
	}
	usersAll, err := tb.store.User.AllUsersCount()
	if err != nil {
	}

	usersReferredDaily, err := tb.store.User.DailyUsersReferred()
	if err != nil {
	}
	usersReferredMonthly, err := tb.store.User.MonthlyUsersReferred()
	if err != nil {
	}
	usersReferredAll, err := tb.store.User.AllUsersReferred()
	if err != nil {
	}

	referralsUsersAll, err := tb.store.Referral.AllUsers()
	if err != nil {
	}
	referralUsersDaily, err := tb.store.Referral.DailyUsers()
	if err != nil {
	}
	referralUsersMonthly, err := tb.store.Referral.MonthlyUsers()
	if err != nil {
	}
	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.From.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        fmt.Sprintf("👑 Премиум пользователей: %d\n\n🚀 Запусков: %d | %d | %d\n🎁 Выполнено бонусов: %d | %d | %d\n\n✅ Статистика пользователей\n|-Саморост: %d | %d | %d\n|-Приглашены: %d | %d | %d\n|-Реферальные ссылки: %d | %d | %d\n", premiumUsers, statsDaily, statsMonthly, statsAll, bonusesDaily, bonusesMonthly, bonusesAll, usersDaily, usersMonthly, usersAll, usersReferredDaily, usersReferredMonthly, usersReferredAll, referralUsersDaily, referralUsersMonthly, referralsUsersAll),
		ReplyMarkup: kb,
	})
}

func (tb tgBot) statisticsMenu(ctx context.Context, b *bot.Bot, update *models.Update) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "За сегодня", CallbackData: statisticsDaily},
				{Text: "За неделю", CallbackData: statisticsWeekly},
				{Text: "За месяц", CallbackData: statisticsMonthly},
			},
			{
				{Text: "За все время", CallbackData: statisticsAll},
			},
			{
				{Text: "Назад", CallbackData: statisticsBack},
				{Text: "В меню", CallbackData: menu},
			},
		},
	}

	b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:      update.CallbackQuery.From.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ReplyMarkup: kb,
	})
}

func (tb tgBot) statisticsDaily(ctx context.Context, b *bot.Bot, update *models.Update) {
	dailyUsersCount, err := tb.store.User.DailyUsersCount()
	if err != nil {
	}
	messagesDaily, err := tb.store.Message.MessagesDaily()
	if err != nil {
	}
	statsDaily, err := tb.store.Stat.Daily()
	if err != nil {
	}

	dailyUsers, err := tb.store.User.DailyUsers()
	if err != nil {
	}

	newDailyUsers := dailyUsersCount
	newDailyUsersPercent := 100

	activeUsersDaily, err := tb.store.User.ActiveUsersDaily()
	if err != nil {
		slog.Error(err.Error())
	}
	activeUsersDailyPercent := (activeUsersDaily / dailyUsersCount) * 100

	deadUsersCount := 0
	for _, user := range dailyUsers {
		if !tb.IsBotBanned(int64(user.ID)) {
			deadUsersCount += 1
		}
	}
	deadUsersPercent := (deadUsersCount / dailyUsersCount) * 100

	aliveUsers := dailyUsersCount - deadUsersCount
	aliveUsersPercent := (aliveUsers / dailyUsersCount) * 100

	premiumUsersCount, err := tb.store.User.PremiumUsersCount()
	if err != nil {
	}
	premiumUsersPercent := (premiumUsersCount / dailyUsersCount) * 100

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("Статистика бота:\n|-Получено сообщений: %d\n|-Получено нажатий: %d\n\nСтатистика пользователей:\n|-Всего: %d\n|-Новых: %d (%d %%)\n|-Активные: %d (%d %%)\n|-Живые: %d (%d %%)\n|-Мертвые: %d (%d %%)\n|-Премиум: %d (%d %%)",
			messagesDaily, statsDaily, dailyUsersCount, newDailyUsers, newDailyUsersPercent, activeUsersDaily, activeUsersDailyPercent, aliveUsers, aliveUsersPercent, deadUsersCount, deadUsersPercent, premiumUsersCount, premiumUsersPercent),
	})
}

func (tb tgBot) statisticsWeekly(ctx context.Context, b *bot.Bot, update *models.Update) {
	usersCount, err := tb.store.User.WeeklyUsersCount()
	if err != nil {
	}
	messagesDaily, err := tb.store.Message.MessagesWeekly()
	if err != nil {
	}
	statsDaily, err := tb.store.Stat.All()
	if err != nil {
	}

	users, err := tb.store.User.WeeklyUsers()
	if err != nil {
	}

	newDailyUsers := usersCount
	newDailyUsersPercent := 100

	activeUsersDaily, err := tb.store.User.ActiveUsersWeekly()
	if err != nil {
		slog.Error(err.Error())
	}
	activeUsersDailyPercent := (activeUsersDaily / usersCount) * 100

	deadUsersCount := 0
	for _, user := range users {
		if !tb.IsBotBanned(int64(user.ID)) {
			deadUsersCount += 1
		}
	}
	deadUsersPercent := (deadUsersCount / usersCount) * 100

	aliveUsers := usersCount - deadUsersCount
	aliveUsersPercent := (aliveUsers / usersCount) * 100

	premiumUsersCount, err := tb.store.User.PremiumUsersCount()
	if err != nil {
	}
	premiumUsersPercent := (premiumUsersCount / usersCount) * 100

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("Статистика бота:\n|-Получено сообщений: %d\n|-Получено нажатий: %d\n\nСтатистика пользователей:\n|-Всего: %d\n|-Новых: %d (%d %%)\n|-Активные: %d (%d %%)\n|-Живые: %d (%d %%)\n|-Мертвые: %d (%d %%)\n|-Премиум: %d (%d %%)",
			messagesDaily, statsDaily, usersCount, newDailyUsers, newDailyUsersPercent, activeUsersDaily, activeUsersDailyPercent, aliveUsers, aliveUsersPercent, deadUsersCount, deadUsersPercent, premiumUsersCount, premiumUsersPercent),
	})
}

func (tb tgBot) statisticsMonthly(ctx context.Context, b *bot.Bot, update *models.Update) {
	usersCount, err := tb.store.User.MonthlyUsersCount()
	if err != nil {
	}
	messagesDaily, err := tb.store.Message.MessagesMonthly()
	if err != nil {
	}
	statsDaily, err := tb.store.Stat.Monthly()
	if err != nil {
	}

	users, err := tb.store.User.MonthlyUsers()
	if err != nil {
	}

	newDailyUsers := usersCount
	newDailyUsersPercent := 100

	activeUsersDaily, err := tb.store.User.ActiveUsersMonthly()
	if err != nil {
		slog.Error(err.Error())
	}
	activeUsersDailyPercent := (activeUsersDaily / usersCount) * 100

	deadUsersCount := 0
	for _, user := range users {
		if !tb.IsBotBanned(int64(user.ID)) {
			deadUsersCount += 1
		}
	}
	deadUsersPercent := (deadUsersCount / usersCount) * 100

	aliveUsers := usersCount - deadUsersCount
	aliveUsersPercent := (aliveUsers / usersCount) * 100

	premiumUsersCount, err := tb.store.User.PremiumUsersCount()
	if err != nil {
	}
	premiumUsersPercent := (premiumUsersCount / usersCount) * 100

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("Статистика бота:\n|-Получено сообщений: %d\n|-Получено нажатий: %d\n\nСтатистика пользователей:\n|-Всего: %d\n|-Новых: %d (%d %%)\n|-Активные: %d (%d %%)\n|-Живые: %d (%d %%)\n|-Мертвые: %d (%d %%)\n|-Премиум: %d (%d %%)",
			messagesDaily, statsDaily, usersCount, newDailyUsers, newDailyUsersPercent, activeUsersDaily, activeUsersDailyPercent, aliveUsers, aliveUsersPercent, deadUsersCount, deadUsersPercent, premiumUsersCount, premiumUsersPercent),
	})
}

func (tb tgBot) statisticsAll(ctx context.Context, b *bot.Bot, update *models.Update) {
	usersCount, err := tb.store.User.AllUsersCount()
	if err != nil {
	}
	messagesDaily, err := tb.store.Message.MessagesAll()
	if err != nil {
	}
	statsDaily, err := tb.store.Stat.All()
	if err != nil {
	}

	users, err := tb.store.User.AllUsers()
	if err != nil {
	}

	newDailyUsers := usersCount
	newDailyUsersPercent := 100

	activeUsersDaily, err := tb.store.User.ActiveUsersAll()
	if err != nil {
		slog.Error(err.Error())
	}
	activeUsersDailyPercent := (activeUsersDaily / usersCount) * 100

	deadUsersCount := 0
	for _, user := range users {
		if !tb.IsBotBanned(int64(user.ID)) {
			deadUsersCount += 1
		}
	}
	deadUsersPercent := (deadUsersCount / usersCount) * 100

	aliveUsers := usersCount - deadUsersCount
	aliveUsersPercent := (aliveUsers / usersCount) * 100

	premiumUsersCount, err := tb.store.User.PremiumUsersCount()
	if err != nil {
	}
	premiumUsersPercent := (premiumUsersCount / usersCount) * 100

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("Статистика бота:\n|-Получено сообщений: %d\n|-Получено нажатий: %d\n\nСтатистика пользователей:\n|-Всего: %d\n|-Новых: %d (%d %%)\n|-Активные: %d (%d %%)\n|-Живые: %d (%d %%)\n|-Мертвые: %d (%d %%)\n|-Премиум: %d (%d %%)",
			messagesDaily, statsDaily, usersCount, newDailyUsers, newDailyUsersPercent, activeUsersDaily, activeUsersDailyPercent, aliveUsers, aliveUsersPercent, deadUsersCount, deadUsersPercent, premiumUsersCount, premiumUsersPercent),
	})
}

func (tb tgBot) bonuses(ctx context.Context, b *bot.Bot, update *models.Update) {
	bonuses, err := tb.store.Bonus.AllBonuses()
	if err != nil {
		slog.Error(err.Error())
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{},
	}

	for _, bonus := range bonuses {
		bonusCompleted, err := tb.store.Bonus.BonusesByID(bonus.ID)
		if err != nil {
			slog.Error(err.Error())
		}

		statusText := "🟥"
		if bonus.Status {
			statusText = "🟩"
		}

		if bonus.Name == "" {
			bonus.Name = " "
		}

		kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
			{Text: fmt.Sprintf("#%d", bonus.ID), CallbackData: bonuseID + strconv.Itoa(bonus.ID)},
			{Text: bonus.Name, CallbackData: "sdf"},
			{Text: strconv.Itoa(bonusCompleted), CallbackData: "sgf"},
			{Text: statusText, CallbackData: "sdfg"},
		})
	}

	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
		{Text: "+ Добавить канал", CallbackData: bonusesCreate},
	})
	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: bonusesBack},
		{Text: "В меню", CallbackData: menu},
	})

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      "- При нажатии на номер #номер - откроется меню управления и редактирование выбранного спонсора",
	})
	b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:      update.CallbackQuery.From.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ReplyMarkup: kb,
	})

}

// func (tb tgBot) bonusesCreate(ctx context.Context, b *bot.Bot, update *models.Update) {

// }
func (tb tgBot) bonusesInfo(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})
	fmt.Println(update.CallbackQuery.Data)

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Изменить канал", CallbackData: bonusesChaneChannel},
			},
			{
				{Text: "Название", CallbackData: bonusesChangeName},
				{Text: "Количество", CallbackData: bonusesChangeMaxUsers},
			},
			{
				{Text: "Проверять", CallbackData: "sdf"},
				{Text: "Не проверять", CallbackData: "sdf"},
			},
			{
				{Text: "Удалить из списка", CallbackData: bonusesDelete},
			},
			{
				{Text: "Назад", CallbackData: bonusesBack},
				{Text: "В меню", CallbackData: menu},
			},
		},
	}
	idString := strings.Split(update.CallbackQuery.Data, "@")[1]
	id, err := strconv.Atoi(idString)
	if err != nil {

	}
	bonus, err := tb.store.Bonus.GetOne(id)
	if err != nil {
	}

	tgChannel, err := tb.GetChannelInfo(bonus.Channel.Name)
	if err != nil {
	}
	bonus.Channel = tgChannel

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.From.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        fmt.Sprintf("|-Канал: @%s\n|-Название кнопки: %s\n|-Количество подписок: %d\n|-Создано: %s", bonus.Channel.Name, bonus.Name, bonus.MaxUsers, bonus.CreatedAt),
		ReplyMarkup: kb,
	})
	fmt.Println(err)
}

func (tb tgBot) callbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	switch update.CallbackQuery.Data {
	// statistics
	case statistics:
		tb.statisticsMenu(ctx, b, update)
	case statisticsDaily:
		tb.statisticsDaily(ctx, b, update)
		tb.statisticsMenu(ctx, b, update)
	case statisticsWeekly:
		tb.statisticsWeekly(ctx, b, update)
		tb.statisticsMenu(ctx, b, update)
	case statisticsMonthly:
		tb.statisticsMonthly(ctx, b, update)
		tb.statisticsMenu(ctx, b, update)
	case statisticsAll:
		tb.statisticsAll(ctx, b, update)
		tb.statisticsMenu(ctx, b, update)
	case statisticsBack:
		tb.adminMenu(ctx, b, update)
	case menu:
		tb.adminMenu(ctx, b, update)

	case bonuses:
		tb.bonuses(ctx, b, update)

		// statistics
		// case userStatistics:
		// 	tb.usersStatisticsCallback(ctx, b, update)
		// case requestsStatistics:
		// 	tb.requestsStatisticsCallback(ctx, b, update)
		// case bonusStatistics:
		// 	tb.bonusesStatisticsCallback(ctx, b, update)
		// // subscription
		// case giveSubscription:
		// 	tb.f.Transition(update.CallbackQuery.From.ID, subscriptionStateAskUserID, update.CallbackQuery.From.ID)

		// // diamonds
		// case giveDiamonds:
		// 	tb.f.Transition(update.CallbackQuery.From.ID, diamondStateAskUserID, update.CallbackQuery.From.ID)

		// // bonuses
		// case createBonus:
		// 	tb.f.Transition(update.CallbackQuery.From.ID, bonusStateAskChannel, update.CallbackQuery.From.ID)
		// case deleteBonus:
		// 	tb.f.Transition(update.CallbackQuery.From.ID, bonusStateDelete, update.CallbackQuery.From.ID)

		// // referrals
		// case createReferral:
		// 	tb.createReferral(ctx, b, update)
		// case deleteReferral:
		// 	tb.f.Transition(update.CallbackQuery.From.ID, referralStateDelete, update.CallbackQuery.From.ID)
		// case getReferrals:
		// 	tb.getReferral(ctx, b, update.CallbackQuery.From.ID)

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

func (tb tgBot) callbackReferralId(f *fsm.FSM, args ...any) {
	chatID := args[0]

	tb.getReferral(context.Background(), tb.b, chatID.(int64))

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      "<b>Выберите реферальную ссылку из списка и напиши ее ID</b>",
		ParseMode: models.ParseModeHTML,
	})
}

func (tb tgBot) usersStatisticsCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	allUsers, err := tb.store.User.AllUsersCount()
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
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, adminRequestsStatistics)
		return
	}
	dailyBonuses, err := tb.store.Bonus.DailyBonusesCount()
	if err != nil {
		tb.informUser(ctx, update.CallbackQuery.From.ID, adminRequestsStatistics)
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
		slog.Error(err.Error())
		tb.informUser(context.Background(), msg.From.ID, internalError)
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
	// bonus := domain.Bonus{
	// 	Channel: domain.Channel{
	// 		Name: bonusScheme.channel_name,
	// 	},
	// 	Award: bonusScheme.award,
	// }
	err := tb.store.Bonus.Create(domain.Bonus{})
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), msg.From.ID, internalError)
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
		slog.Error(err.Error())
		tb.informUser(context.Background(), msg.From.ID, internalError)
		return
	}
	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: msg.From.ID,
		Text:   "Бонус удален",
	})
}

func (tb tgBot) createReferral(ctx context.Context, b *bot.Bot, update *models.Update) {
	err := tb.store.Referral.Create()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(update.CallbackQuery.From.ID), internalError)
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.CallbackQuery.From.ID,
		Text:   "Реферальная ссылка создана",
	})
}

func (tb tgBot) deteleReferral(refID int, msg *models.Message) {
	err := tb.store.Referral.Delete(refID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), int64(msg.From.ID), internalError)
		return
	}

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: msg.From.ID,
		Text:   "Реферальная ссылка удалена",
	})
}

func (tb tgBot) getReferral(ctx context.Context, b *bot.Bot, userID int64) {
	links, err := tb.store.Referral.GetAll()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, userID, internalError)
		return
	}
	var message string
	for _, link := range links {
		link.SetLink()
		countUsers, err := tb.store.Referral.CountUsers(link.ID)
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, userID, internalError)
			return
		}
		f := fmt.Sprintf("%d: %s <b>(%d переходов)</b>\n", link.ID, link.Link, countUsers)
		message = message + f
	}

	b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID:    userID,
		Text:      "<b>Реферальные ссылки</b>\n" + message,
		ParseMode: models.ParseModeHTML,
	})
}
