package tgbot

import (
	"context"
	"fmt"
	"gpt-bot/internal/db/domain"
	"gpt-bot/utils"
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

	stateDefault         fsm.StateID = "stateDefault"
	stateChannelNamePost fsm.StateID = "stateChannelNamePost"
	stateChannelNameHand fsm.StateID = "stateChannelNameHand"
	stateChannelLink     fsm.StateID = "stateChannelLink"
	stateBonusName       fsm.StateID = "stateBonusName"
	stateBonusMaxUsers   fsm.StateID = "stateBonusMaxUsers"
	stateBonusAward      fsm.StateID = "stateBonusAward"

	stateUserLimitsModel  fsm.StateID = "stateUserLimitsModel"
	stateUserLimitsAmount fsm.StateID = "stateUserLimitsAmount"
	stateUserPremium      fsm.StateID = "stateUserPremium"
	stateUserDiamonds     fsm.StateID = "stateUserDiamonds"

	stateReferralName fsm.StateID = "stateReferralName"
	stateReferralCode fsm.StateID = "stateReferralCode"
)

const (
	menu = "btn_menu"

	statistics        = "btn_1"
	statisticsDaily   = "btn_1_1"
	statisticsWeekly  = "btn_1_2"
	statisticsMonthly = "btn_1_3"
	statisticsAll     = "btn_1_4"
	statisticsBack    = "btn_1_5"

	bonuses                      = "btn_2"
	bonusesInfo                  = "btn_2_1"
	bonusesCreate                = "btn_2_2"
	bonusesBack                  = "btn_2_3"
	bonusesChangeChannelNamePost = "btn_2_4"
	bonusesChangeChannelNameHand = "btn_2_5"
	bonusesChangeName            = "btn_2_6"
	bonusesChangeMaxUsers        = "btn_2_7"
	bonusesChangeAward           = "btn_2_8"
	bonusesDelete                = "btn_2_9"
	bonusCheckTrue               = "btn_2_10"
	bonusCheckFalse              = "btn_2_11"
	bonuseID                     = "id@"

	usersPage        = "page@"
	usersSinge       = "user@"
	usersMakeAdmin   = "admin@"
	premiumUsersPage = "prempage@"
	usersLimits      = "btn_3_1"
	usersPremium     = "btn_3_2"
	usersDiamonds    = "btn_3_3"

	gpt4o      = "model@gpt_4o"
	gpt4o_mini = "model@gpt_4o_mini"
	o1_mini    = "model@o1_mini"
	o1_preview = "model@o1_preview"
	dall_e_3   = "model@dall_e_3"
	runware    = "model@runware"

	advancedMonth = "sub@advanced-month"
	advancedYear  = "sub@advanced-year"
	ultimateMonth = "sub@ultimate-month"
	ultimateYear  = "sub@ultimate-year"

	referralsPage       = "referral@"
	referralsSingle     = "ref@"
	referralsDel        = "del@"
	referralsChangeName = "btn_4_1"
	referralsChangeCode = "btn_4_2"
	referralsCreate     = "btn_4_3"

	requests        = "btn_5"
	requestsDaily   = "btn_5_1"
	requestsWeekly  = "btn_5_2"
	requestsMonthly = "btn_5_3"
	requestsAll     = "btn_5_4"
)

type bonusData struct {
	bonusID   int
	channelID int64
	link      string
	name      string
	maxUsers  int
}

type userLimits struct {
	userID int
	model  string
	amount int
}

type usersList struct {
	list map[int][]domain.User
}

type referralList struct {
	list map[int][]domain.Referral
}

var bonusScheme bonusData
var uLimits userLimits
var uList usersList
var rList referralList
var referralID int

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
				{Text: "Пользователи", CallbackData: usersPage + "1"},
				{Text: "Ссылки", CallbackData: referralsPage + "1"},
			},
			{
				{Text: "Запросы", CallbackData: requests},
			},
		},
	}

	premiumUsers, err := tb.store.User.PremiumUsersCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return

	}
	statsDaily, err := tb.store.Stat.Daily()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}
	statsMonthly, err := tb.store.Stat.Monthly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return

	}
	statsAll, err := tb.store.Stat.All()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}

	bonusesDaily, err := tb.store.Bonus.DailyBonusesCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}
	bonusesMonthly, err := tb.store.Bonus.MonthlyBonusesCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}
	bonusesAll, err := tb.store.Bonus.AllBonusesCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}

	usersDaily, err := tb.store.User.DailyUsersCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}
	usersMonthly, err := tb.store.User.MonthlyUsersCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}
	usersAll, err := tb.store.User.AllUsersCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}

	usersReferredDaily, err := tb.store.User.DailyUsersReferred()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}
	usersReferredMonthly, err := tb.store.User.MonthlyUsersReferred()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}
	usersReferredAll, err := tb.store.User.AllUsersReferred()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}

	referralsUsersAll, err := tb.store.Referral.AllUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}
	referralUsersDaily, err := tb.store.Referral.DailyUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}
	referralUsersMonthly, err := tb.store.Referral.MonthlyUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, internalError)
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.From.ID,
		Text:        fmt.Sprintf("👑 <b>Премиум пользователей:</b> %d\n\n🚀 <b>Запусков:</b> %d | %d | %d\n🎁 <b>Выполнено бонусов:</b> %d | %d | %d\n\n✅ <b>Статистика пользователей</b>\n|-Саморост: %d | %d | %d\n|-Приглашены: %d | %d | %d\n|-Реферальные ссылки: %d | %d | %d\n", premiumUsers, statsDaily, statsMonthly, statsAll, bonusesDaily, bonusesMonthly, bonusesAll, usersDaily, usersMonthly, usersAll, usersReferredDaily, usersReferredMonthly, usersReferredAll, referralUsersDaily, referralUsersMonthly, referralsUsersAll),
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
				{Text: "Пользователи", CallbackData: usersPage + "1"},
				{Text: "Ссылки", CallbackData: referralsPage + "1"},
			},
			{
				{Text: "Запросы", CallbackData: requests},
			},
		},
	}
	premiumUsers, err := tb.store.User.PremiumUsersCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	statsDaily, err := tb.store.Stat.Daily()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	statsMonthly, err := tb.store.Stat.Monthly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	statsAll, err := tb.store.Stat.All()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	bonusesDaily, err := tb.store.Bonus.DailyBonusesCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return

	}

	bonusesMonthly, err := tb.store.Bonus.MonthlyBonusesCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	bonusesAll, err := tb.store.Bonus.AllBonusesCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	usersDaily, err := tb.store.User.DailyUsersCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	usersMonthly, err := tb.store.User.MonthlyUsersCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	usersAll, err := tb.store.User.AllUsersCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	usersReferredDaily, err := tb.store.User.DailyUsersReferred()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	usersReferredMonthly, err := tb.store.User.MonthlyUsersReferred()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	usersReferredAll, err := tb.store.User.AllUsersReferred()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	referralsUsersAll, err := tb.store.Referral.AllUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	referralUsersDaily, err := tb.store.Referral.DailyUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	referralUsersMonthly, err := tb.store.Referral.MonthlyUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.From.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        fmt.Sprintf("👑 <b>Премиум пользователей:</b> %d\n\n🚀 <b>Запусков:</b> %d | %d | %d\n🎁 <b>Выполнено бонусов:</b> %d | %d | %d\n\n✅ <b>Статистика пользователей</b>\n|-Саморост: %d | %d | %d\n|-Приглашены: %d | %d | %d\n|-Реферальные ссылки: %d | %d | %d\n", premiumUsers, statsDaily, statsMonthly, statsAll, bonusesDaily, bonusesMonthly, bonusesAll, usersDaily, usersMonthly, usersAll, usersReferredDaily, usersReferredMonthly, usersReferredAll, referralUsersDaily, referralUsersMonthly, referralsUsersAll),
		ParseMode:   models.ParseModeHTML,
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

	dailyUsersCount, err := tb.store.User.DailyUsersCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return

	}
	messagesDaily, err := tb.store.Message.MessagesDaily()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	statsDaily, err := tb.store.Stat.Daily()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	dailyUsers, err := tb.store.User.DailyUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	newDailyUsers := dailyUsersCount
	newDailyUsersPercent := 100

	activeUsersDaily, err := tb.store.User.ActiveUsersDaily()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	var activeUsersDailyPercent float32
	if dailyUsersCount != 0 {
		activeUsersDailyPercent = (float32(activeUsersDaily) / float32(dailyUsersCount)) * 100
	}

	deadUsersCount := 0
	for _, user := range dailyUsers {
		if !tb.IsBotBanned(int64(user.ID)) {
			deadUsersCount += 1
		}
	}

	var deadUsersPercent float32
	if dailyUsersCount != 0 {
		deadUsersPercent = (float32(deadUsersCount) / float32(dailyUsersCount)) * 100
	}

	aliveUsers := dailyUsersCount - deadUsersCount
	var aliveUsersPercent float32
	if dailyUsersCount != 0 {
		aliveUsersPercent = (float32(aliveUsers) / float32(dailyUsersCount)) * 100
	}

	premiumUsersCount, err := tb.store.User.PremiumUsersCountDaily()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	var premiumUsersPercent float32
	if dailyUsersCount != 0 {
		premiumUsersPercent = (float32(premiumUsersCount) / float32(dailyUsersCount)) * 100
	}

	geoUsers, err := tb.store.User.GeoUsersDaily()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return

	}
	var geoUsersPercent float32
	if dailyUsersCount != 0 {
		geoUsersPercent = (float32(geoUsers) / float32(dailyUsersCount)) * 100
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("📊<b>Статистика бота:</b>\n|-Получено сообщений: <b>%d</b>\n|-Получено нажатий: <b>%d</b>\n\n👥<b>Статистика пользователей:</b>\n|-Всего: <b>%d</b>\n|-Новых: <b>%d</b> (%d %%)\n|-Активные: <b>%d</b> (%d %%)\n|-Живые: <b>%d</b> (%d %%)\n|-Мертвые: <b>%d</b> (%d %%)\n|-Премиум: <b>%d</b> (%d %%)\n\n🌎<b>Анализ аудитории:</b>\n|-🇷🇺RU: <b>%d</b> (%d %%)\n",
			messagesDaily, statsDaily, dailyUsersCount, newDailyUsers, newDailyUsersPercent, activeUsersDaily, int(activeUsersDailyPercent), aliveUsers, int(aliveUsersPercent), deadUsersCount, int(deadUsersPercent), premiumUsersCount, int(premiumUsersPercent), geoUsers, int(geoUsersPercent)),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}

func (tb tgBot) statisticsWeekly(ctx context.Context, b *bot.Bot, update *models.Update) {
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
	usersCount, err := tb.store.User.WeeklyUsersCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	messagesDaily, err := tb.store.Message.MessagesWeekly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	statsDaily, err := tb.store.Stat.All()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	users, err := tb.store.User.WeeklyUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	newDailyUsers := usersCount
	newDailyUsersPercent := 100

	activeUsersDaily, err := tb.store.User.ActiveUsersWeekly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	activeUsersDailyPercent := (float32(activeUsersDaily) / float32(usersCount)) * 100

	deadUsersCount := 0
	for _, user := range users {
		if !tb.IsBotBanned(int64(user.ID)) {
			deadUsersCount += 1
		}
	}
	deadUsersPercent := (float32(deadUsersCount) / float32(usersCount)) * 100

	aliveUsers := usersCount - deadUsersCount
	aliveUsersPercent := (float32(aliveUsers) / float32(usersCount)) * 100

	premiumUsersCount, err := tb.store.User.PremiumUsersCountWeekly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	premiumUsersPercent := (float32(premiumUsersCount) / float32(usersCount)) * 100

	geoUsers, err := tb.store.User.GeoUsersWeekly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return

	}

	var geoUsersPercent float32
	if usersCount != 0 {
		geoUsersPercent = (float32(geoUsers) / float32(usersCount)) * 100
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("📊<b>Статистика бота:</b>\n|-Получено сообщений: <b>%d</b>\n|-Получено нажатий: <b>%d</b>\n\n👥<b>Статистика пользователей:</b>\n|-Всего: <b>%d</b>\n|-Новых: <b>%d</b> (%d %%)\n|-Активные: <b>%d</b> (%d %%)\n|-Живые: <b>%d</b> (%d %%)\n|-Мертвые: <b>%d</b> (%d %%)\n|-Премиум: <b>%d</b> (%d %%)\n\n🌎<b>Анализ аудитории:</b>\n|-🇷🇺RU: <b>%d</b> (%d %%)\n",
			messagesDaily, statsDaily, usersCount, newDailyUsers, int(newDailyUsersPercent), activeUsersDaily, int(activeUsersDailyPercent), aliveUsers, int(aliveUsersPercent), deadUsersCount, int(deadUsersPercent), premiumUsersCount, int(premiumUsersPercent), geoUsers, int(geoUsersPercent)),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}

func (tb tgBot) statisticsMonthly(ctx context.Context, b *bot.Bot, update *models.Update) {
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

	usersCount, err := tb.store.User.MonthlyUsersCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	messagesDaily, err := tb.store.Message.MessagesMonthly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	statsDaily, err := tb.store.Stat.Monthly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	users, err := tb.store.User.MonthlyUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	newDailyUsers := usersCount
	newDailyUsersPercent := 100

	activeUsersDaily, err := tb.store.User.ActiveUsersMonthly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	activeUsersDailyPercent := (float32(activeUsersDaily) / float32(usersCount)) * 100

	deadUsersCount := 0
	for _, user := range users {
		if !tb.IsBotBanned(int64(user.ID)) {
			deadUsersCount += 1
		}
	}
	deadUsersPercent := (float32(deadUsersCount) / float32(usersCount)) * 100

	aliveUsers := usersCount - deadUsersCount
	aliveUsersPercent := (float32(aliveUsers) / float32(usersCount)) * 100

	premiumUsersCount, err := tb.store.User.PremiumUsersCountMonthly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	premiumUsersPercent := (float32(premiumUsersCount) / float32(usersCount)) * 100

	geoUsers, err := tb.store.User.GeoUsersMonthly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return

	}

	var geoUsersPercent float32
	if usersCount != 0 {
		geoUsersPercent = (float32(geoUsers) / float32(usersCount)) * 100
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("📊<b>Статистика бота:</b>\n|-Получено сообщений: <b>%d</b>\n|-Получено нажатий: <b>%d</b>\n\n👥<b>Статистика пользователей:</b>\n|-Всего: <b>%d</b>\n|-Новых: <b>%d</b> (%d %%)\n|-Активные: <b>%d</b> (%d %%)\n|-Живые: <b>%d</b> (%d %%)\n|-Мертвые: <b>%d</b> (%d %%)\n|-Премиум: <b>%d</b> (%d %%)\n\n🌎<b>Анализ аудитории:</b>\n|-🇷🇺RU: <b>%d</b> (%d %%)\n",
			messagesDaily, statsDaily, usersCount, newDailyUsers, newDailyUsersPercent, activeUsersDaily, int(activeUsersDailyPercent), aliveUsers, int(aliveUsersPercent), deadUsersCount, int(deadUsersPercent), premiumUsersCount, int(premiumUsersPercent), geoUsers, int(geoUsersPercent)),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}

func (tb tgBot) statisticsAll(ctx context.Context, b *bot.Bot, update *models.Update) {
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
	usersCount, err := tb.store.User.AllUsersCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	messagesDaily, err := tb.store.Message.MessagesAll()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	statsDaily, err := tb.store.Stat.All()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	users, err := tb.store.User.AllUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	newDailyUsers := usersCount
	newDailyUsersPercent := 100

	activeUsersDaily, err := tb.store.User.ActiveUsersAll()
	if err != nil {
		slog.Error(err.Error())
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	activeUsersDailyPercent := (float32(activeUsersDaily) / float32(usersCount)) * 100

	deadUsersCount := 0
	for _, user := range users {
		if !tb.IsBotBanned(int64(user.ID)) {
			deadUsersCount += 1
		}
	}
	deadUsersPercent := (float32(deadUsersCount) / float32(usersCount)) * 100

	aliveUsers := usersCount - deadUsersCount
	aliveUsersPercent := (float32(aliveUsers) / float32(usersCount)) * 100

	premiumUsersCount, err := tb.store.User.PremiumUsersCount()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
	premiumUsersPercent := (float32(premiumUsersCount) / float32(usersCount)) * 100

	geoUsers, err := tb.store.User.GeoUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	var geoUsersPercent float32
	if usersCount != 0 {
		geoUsersPercent = (float32(geoUsers) / float32(usersCount)) * 100
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("📊<b>Статистика бота:</b>\n|-Получено сообщений: <b>%d</b>\n|-Получено нажатий: <b>%d</b>\n\n👥<b>Статистика пользователей:</b>\n|-Всего: <b>%d</b>\n|-Новых: <b>%d</b> (%d %%)\n|-Активные: <b>%d</b> (%d %%)\n|-Живые: <b>%d</b> (%d %%)\n|-Мертвые: <b>%d</b> (%d %%)\n|-Премиум: <b>%d</b> (%d %%)\n\n🌎<b>Анализ аудитории:</b>\n|-🇷🇺RU: <b>%d</b> (%d %%)\n",
			messagesDaily, statsDaily, usersCount, newDailyUsers, newDailyUsersPercent, activeUsersDaily, int(activeUsersDailyPercent), aliveUsers, int(aliveUsersPercent), deadUsersCount, int(deadUsersPercent), premiumUsersCount, int(premiumUsersPercent), geoUsers, int(geoUsersPercent)),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}

func (tb tgBot) bonuses(ctx context.Context, b *bot.Bot, update *models.Update) {
	bonuses, err := tb.store.Bonus.AllBonuses()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{},
	}

	for _, bonus := range bonuses {
		bonusCompleted, err := tb.store.Bonus.BonusesByID(bonus.ID)
		if err != nil {
			// slog.Error(err.Error())
			slog.Error(err.Error())
			tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
			return
		}

		statusText := "🟥"
		if bonus.Check {
			statusText = "🟩"
		}

		if bonus.Name == "" {
			bonus.Name = " "
		}

		kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
			{Text: fmt.Sprintf("#%d", bonus.ID), CallbackData: bonuseID + strconv.Itoa(bonus.ID)},
			{Text: bonus.Name, CallbackData: "sdf", URL: bonus.Link},
			{Text: fmt.Sprintf("%d/%d", bonusCompleted, bonus.MaxUsers), CallbackData: "sgf"},
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

	if update.CallbackQuery != nil {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      update.CallbackQuery.From.ID,
			MessageID:   update.CallbackQuery.Message.Message.ID,
			Text:        "Добавление и редактирование 🎁 бонусов",
			ReplyMarkup: kb,
		})
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.CallbackQuery.From.ID,
			Text:        "Добавление и редактирование 🎁 бонусов",
			ReplyMarkup: kb,
		})
	}
}

func (tb tgBot) bonusInfo(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	idString := strings.Split(update.CallbackQuery.Data, "@")
	if len(idString) > 1 {
		id, _ := strconv.Atoi(idString[1])
		bonusScheme.bonusID = id
		bonus, err := tb.store.Bonus.GetOne(id)
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
			return
		}
		bonusCompleted, err := tb.store.Bonus.BonusesByID(bonus.ID)

		isCheck := []string{}
		if bonus.Check {
			isCheck = append(isCheck, "* Проверять *")
			isCheck = append(isCheck, "Не проверять")

		} else {
			isCheck = append(isCheck, "Проверять")
			isCheck = append(isCheck, "* Не проверять *")
		}

		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "Изменить канал", CallbackData: bonusesChangeChannelNamePost},
				},
				{
					{Text: "Название", CallbackData: bonusesChangeName},
					{Text: "Количество", CallbackData: bonusesChangeMaxUsers},
					{Text: "Награда", CallbackData: bonusesChangeAward},
				},
				{
					{Text: isCheck[0], CallbackData: bonusCheckTrue},
					{Text: isCheck[1], CallbackData: bonusCheckFalse},
				},
				{
					{Text: "Удалить из списка", CallbackData: bonusesDelete},
				},
				{
					{Text: "Назад", CallbackData: bonuses},
					{Text: "В меню", CallbackData: menu},
				},
			},
		}

		tgChannel, err := tb.GetChannelInfo(bonus.Channel.ID)
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
			return
		}
		bonus.Channel = tgChannel

		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      update.CallbackQuery.From.ID,
			MessageID:   update.CallbackQuery.Message.Message.ID,
			Text:        fmt.Sprintf("|-Канал: <b>%d</b>\n|-Название кнопки: <b>%s</b>\n|-Награда: <b>%d</b>\n|-Количество подписок: <b>%d/%d</b>\n|-Создано: <b>%s</b>", bonus.Channel.ID, bonus.Name, bonus.Award, bonusCompleted, bonus.MaxUsers, bonus.CreatedAt),
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		})
	}
}

func (tb tgBot) bonusCreate(ctx context.Context, b *bot.Bot, update *models.Update) {
	err := tb.store.Bonus.Create()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.CallbackQuery.From.ID, internalError)
		return
	}
}

func (tb tgBot) bonusUpdate(bonusScheme bonusData, msg *models.Message) {
	err := tb.store.Bonus.UpdateChannel(bonusScheme.bonusID, int(bonusScheme.channelID), bonusScheme.link)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), msg.From.ID, internalError)
		return
	}
	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: msg.From.ID,
		Text:   "🥳 Спонсор успешно добавлен!",
	})
}

func (tb tgBot) bonusDelete(ctx context.Context, b *bot.Bot, update *models.Update) {
	err := tb.store.Bonus.Delete(bonusScheme.bonusID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}
}

func (tb tgBot) bonusCheck(status bool, update *models.Update) {
	if status {
		err := tb.store.Bonus.UpdateStatus(bonusScheme.bonusID, true)
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
			return
		}
	} else {
		err := tb.store.Bonus.UpdateStatus(bonusScheme.bonusID, false)
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
			return
		}
	}
}

func (tb tgBot) bonusName(bonusScheme bonusData, msg *models.Message) {
	err := tb.store.Bonus.UpdateName(bonusScheme.bonusID, bonusScheme.name)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), msg.From.ID, internalError)
		return
	}
}

func (tb tgBot) bonusMaxUsers(bonusScheme bonusData, msg *models.Message) {
	err := tb.store.Bonus.UpdateMaxUsers(bonusScheme.bonusID, bonusScheme.maxUsers)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), msg.From.ID, internalError)
		return
	}
}

func (tb tgBot) bonusAward(award int, msg *models.Message) {
	err := tb.store.Bonus.UpdateAward(bonusScheme.bonusID, award)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), msg.From.ID, internalError)
		return

	}
}

func (tb tgBot) users(ctx context.Context, b *bot.Bot, update *models.Update) {
	users, err := tb.store.User.AllUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	uList.list = map[int][]domain.User{}

	page := 1
	sum := 0
	for _, user := range users {
		if sum == 5 {
			sum = 0
			page++
		}

		chat, _ := tb.b.GetChat(ctx, &bot.GetChatParams{
			ChatID: user.ID,
		})

		if chat != nil {
			if chat.Username == "" {
				chat.Username = "null"
			}
			// kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
			// 	{Text: fmt.Sprintf("#%d - @%s", user.ID, chat.Username), CallbackData: "wer"},
			// })
			uList.list[page] = append(uList.list[page], domain.User{ID: user.ID, Username: chat.Username})
			sum++
		}
	}
}

func (tb tgBot) premiumUsers(ctx context.Context, b *bot.Bot, update *models.Update) {
	users, err := tb.store.User.PremiumUsers()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	uList.list = map[int][]domain.User{}

	page := 1
	sum := 0
	for _, user := range users {
		if sum == 5 {
			sum = 0
			page++
		}

		chat, _ := tb.b.GetChat(ctx, &bot.GetChatParams{
			ChatID: user.ID,
		})

		if chat != nil {
			if chat.Username == "" {
				chat.Username = "null"
			}
			// kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
			// 	{Text: fmt.Sprintf("#%d - @%s", user.ID, chat.Username), CallbackData: "wer"},
			// })
			uList.list[page] = append(uList.list[page], domain.User{ID: user.ID, Username: chat.Username})
			sum++
		}
	}
}

func (tb tgBot) premiumUsersPage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery != nil {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
	}

	tb.premiumUsers(ctx, b, update)

	pageString := strings.Split(update.CallbackQuery.Data, "@")
	var page = 1
	if len(pageString) > 1 {
		page, _ = strconv.Atoi(pageString[1])
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{},
	}

	for _, u := range uList.list[page] {
		kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
			{Text: fmt.Sprintf("👑 #%d - @%s", u.ID, u.Username), CallbackData: usersSinge + strconv.Itoa(u.ID)},
		})
	}

	pages := []models.InlineKeyboardButton{}
	for page := range uList.list {
		pages = append(pages, models.InlineKeyboardButton{
			Text: strconv.Itoa(page), CallbackData: usersPage + strconv.Itoa(page),
		})
	}
	kb.InlineKeyboard = append(kb.InlineKeyboard, pages)
	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
		{Text: "Выгрузить список", CallbackData: "string"},
	})
	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: usersPage + "1"},
		{Text: "В меню", CallbackData: menu},
	})

	if update.CallbackQuery != nil {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      update.CallbackQuery.From.ID,
			MessageID:   update.CallbackQuery.Message.Message.ID,
			Text:        "👤 <b>Пользователи</b>",
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		})
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.CallbackQuery.From.ID,
			Text:        "👤 <b>Пользователи</b>",
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		})
	}
}

func (tb tgBot) usersAdmin(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	idString := strings.Split(update.CallbackQuery.Data, "@")
	var id int
	if len(idString) > 1 {
		id, _ = strconv.Atoi(idString[1])

	}

	err := tb.store.Admin.MakeAdmin(id)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	tb.adminMenu(ctx, b, update)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    id,
		Text:      "🎉 <b>Вам дали доступ к админ-панеле WebAI</b>\n\nДля доступа к ней, напишите /admin в <b><a href='https://t.me/webai_robot'>бота</a></b>",
		ParseMode: models.ParseModeHTML,
	})

}

func (tb tgBot) usersLimitsModel(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	model := strings.Split(update.CallbackQuery.Data, "@")
	if len(model) > 1 {
		uLimits.model = model[1]
	}

	tb.f.Transition(update.CallbackQuery.From.ID, stateUserLimitsAmount, update.CallbackQuery)
}

func (tb tgBot) usersPage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery != nil {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
	}

	tb.users(ctx, b, update)

	pageString := strings.Split(update.CallbackQuery.Data, "@")
	var page = 1
	if len(pageString) > 1 {
		page, _ = strconv.Atoi(pageString[1])
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{},
	}

	for _, u := range uList.list[page] {
		kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
			{Text: fmt.Sprintf("👤 #%d - @%s", u.ID, u.Username), CallbackData: usersSinge + strconv.Itoa(u.ID)},
		})
	}

	pages := []models.InlineKeyboardButton{}
	for page := range uList.list {
		pages = append(pages, models.InlineKeyboardButton{
			Text: strconv.Itoa(page), CallbackData: usersPage + strconv.Itoa(page),
		})
	}
	kb.InlineKeyboard = append(kb.InlineKeyboard, pages)
	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
		{Text: "Выгрузить список", CallbackData: "string"},
	})
	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
		{Text: "👑 Премиум-User", CallbackData: premiumUsersPage + "1"},
	})
	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: menu},
		{Text: "В меню", CallbackData: menu},
	})

	if update.CallbackQuery != nil {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      update.CallbackQuery.From.ID,
			MessageID:   update.CallbackQuery.Message.Message.ID,
			Text:        "👤 <b>Пользователи</b>",
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		})
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.CallbackQuery.From.ID,
			Text:        "👤 <b>Пользователи</b>",
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		})
	}
}

func (tb tgBot) userSingle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery != nil {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
	}

	idString := strings.Split(update.CallbackQuery.Data, "@")
	var id int
	if len(idString) > 1 {
		id, _ = strconv.Atoi(idString[1])
	}
	uLimits.userID = id

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Сделать администратором", CallbackData: usersMakeAdmin + strconv.Itoa(id)},
			},
			{
				{Text: "Установить количество запросов", CallbackData: usersLimits},
			},
			{
				{Text: "Выдать премиум", CallbackData: usersPremium},
			},
			{
				{Text: "Выдать алмазы", CallbackData: usersDiamonds},
			},

			{
				{Text: "Назад", CallbackData: usersPage + "1"},
				{Text: "В меню", CallbackData: menu},
			},
		},
	}

	user, err := tb.store.User.GetByID(id)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}
	chat, err := tb.b.GetChat(ctx, &bot.GetChatParams{
		ChatID: user.ID,
	})
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}
	if chat != nil {
		user.Username = chat.Username
	}

	messages, err := tb.store.Message.RequestsByUser(user.ID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	bonuses, err := tb.store.Bonus.BonusesByUser(user.ID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	lastMsg, err := tb.store.Message.LastMessageUser(user.ID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("👤 <b>Управление пользователем</b>\n|-Айди: <b><code>%d</code></b>\n|-Юзернейм: <b><code>@%s</code></b>\n|-Отправлено сообщений:\n<pre><code>ChatGPT 4o-mini: %d\nChatGPT o1-mini: %d\nChatGPT o1-preview: %d\nChatGPT 4o: %d\nDall-e-3: %d\nRunware: %d</code></pre>\n|-Выполнено бонусов: <b>%d</b>\n|-Баланс: <b>%d</b>\n|-Последний актив: <b>%s</b>\n|-Зарегестрирован: <b>%s</b>\n|-Подписка: <b>%s</b>\n|-Дневной лимит:\n<pre><code>ChatGPT 4o-mini: %d\nChatGPT o1-mini: %d\nChatGPT o1-preview: %d\nChatGPT 4o: %d\nDall-e-3: %d\nRunware: %d\n</code></pre>",
			user.ID, user.Username, messages["gpt-4o-mini"], messages["o1-mini"], messages["o1-preview"], messages["gpt-4o"], messages["dall-e-3"], messages["runware"], bonuses, user.Balance, lastMsg, user.CreatedAt, user.Subscription.Name, user.Limits.Gpt4oMini, user.Limits.O1Mini, user.Limits.O1Preview, user.Limits.Gpt4o, user.Limits.Dalle3, user.Limits.Runware),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}

func (tb tgBot) usersLimits(uLimits userLimits, msg *models.Message) {
	err := tb.store.Limits.AddLimits(uLimits.userID, uLimits.model, uLimits.amount)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), msg.From.ID, internalError)
		return
	}

	token := utils.NewToken()
	token.SetUserID(uLimits.userID)
	err = token.SignJWT()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), msg.From.ID, internalError)
		return
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Открыть приложение", WebApp: &models.WebAppInfo{
					URL: tb.telegram.GetWebAppUrl() + "?token=" + token.GetStrToken(),
				}},
			},
		},
	}

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID:      uLimits.userID,
		Text:        fmt.Sprintf("🎉 Администратор начислил вам <b>%d</b> лимитов для модели <b>%s</b>\n\n<i>Чтобы использовать их в приложении нажмите кнопку ниже...</i>", uLimits.amount, uLimits.model),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}

func (tb tgBot) usersPremium(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	subNameValue := strings.Split(update.CallbackQuery.Data, "@")
	var subName string
	if len(subNameValue) > 1 {
		subName = subNameValue[1]
	}

	var subscription domain.Payment
	subscription.SubscriptionName = subName
	subscription.UserID = uLimits.userID
	subscription.ToReadable()

	err := tb.store.Subscription.Update(subscription.UserID, subscription.SubscriptionName, subscription.End)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}
	diamonds, err := tb.store.Subscription.DailyDiamonds(subscription.SubscriptionName)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}
	err = tb.store.User.FillBalance(subscription.UserID, diamonds)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}
	limits := domain.NewLimits(subscription.UserID, subscription.SubscriptionName)
	err = tb.store.Limits.Update(limits)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	token := utils.NewToken()
	token.SetUserID(subscription.UserID)
	err = token.SignJWT()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Открыть приложение", WebApp: &models.WebAppInfo{
					URL: tb.telegram.GetWebAppUrl() + "?token=" + token.GetStrToken(),
				}},
			},
		},
	}

	b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID:      subscription.UserID,
		Text:        fmt.Sprintf("🎉 Администратор выдал вам <b>%s</b> подписку на срок до <b>%s</b>\n\n<i>Чтобы использовать ее в приложении нажмите кнопку ниже...</i>", subscription.SubscriptionName, subscription.End),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})

	tb.adminMenu(ctx, b, update)
}

func (tb tgBot) usersDiamonds(amount int, msg *models.Message) {
	err := tb.store.User.RaiseBalance(uLimits.userID, amount)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), msg.From.ID, internalError)
		return
	}

	token := utils.NewToken()
	token.SetUserID(uLimits.userID)
	err = token.SignJWT()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), msg.From.ID, internalError)
		return
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Открыть приложение", WebApp: &models.WebAppInfo{
					URL: tb.telegram.GetWebAppUrl() + "?token=" + token.GetStrToken(),
				}},
			},
		},
	}

	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID:      uLimits.userID,
		Text:        fmt.Sprintf("🎉 Администратор выдал вам <b>%d</b> алмазов\n\n<i>Чтобы использовать их в приложении нажмите кнопку ниже...</i>", amount),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}

func (tb tgBot) referrals(ctx context.Context, b *bot.Bot, update *models.Update) {
	links, err := tb.store.Referral.GetAll()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	rList.list = map[int][]domain.Referral{}

	page := 1
	sum := 0
	for _, link := range links {
		if sum == 5 {
			sum = 0
			page++
		}

		rList.list[page] = append(rList.list[page], link)
		sum++
	}
}

func (tb tgBot) referralsPage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery != nil {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})
	}

	tb.referrals(ctx, b, update)

	var page = 1
	if update.CallbackQuery != nil {
		pageString := strings.Split(update.CallbackQuery.Data, "@")
		if pageString[0] != "del" && pageString[0] != "btn_4_3" {
			p, err := strconv.Atoi(pageString[1])
			if err != nil {

			}
			page = p
		}
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{},
	}

	for _, link := range rList.list[page] {
		kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
			{Text: fmt.Sprintf("🔗 #%d - %s", link.ID, link.Name), CallbackData: referralsSingle + strconv.Itoa(link.ID)},
		})
	}

	pages := []models.InlineKeyboardButton{}
	for page := range rList.list {
		pages = append(pages, models.InlineKeyboardButton{
			Text: strconv.Itoa(page), CallbackData: referralsPage + strconv.Itoa(page),
		})
	}

	kb.InlineKeyboard = append(kb.InlineKeyboard, pages)
	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
		{Text: "Выгрузить список", CallbackData: "string"},
		{Text: "+ Создать новую ссылку", CallbackData: referralsCreate},
	})
	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: menu},
		{Text: "В меню", CallbackData: menu},
	})

	if update.CallbackQuery != nil {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      update.CallbackQuery.From.ID,
			MessageID:   update.CallbackQuery.Message.Message.ID,
			Text:        "🔗 <b>Реферальные ссылки</b>",
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		})

	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.CallbackQuery.From.ID,
			Text:        "🔗 <b>Реферальные ссылки</b>",
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		})

	}

}

func (tb tgBot) referralSingle(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	tb.referrals(ctx, b, update)

	idString := strings.Split(update.CallbackQuery.Data, "@")
	var id int
	if len(idString) > 1 {
		id, _ = strconv.Atoi(idString[1])
	}

	referralID = id

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Изменить ссылку", CallbackData: referralsChangeCode},
				{Text: "Изменить название", CallbackData: referralsChangeName},
			},
			{
				{Text: "Удалить из списка", CallbackData: referralsDel + strconv.Itoa(id)},
			},
			{
				{Text: "Назад", CallbackData: referralsPage + "1"},
				{Text: "В меню", CallbackData: menu},
			},
		},
	}

	ref, err := tb.store.Referral.GetOneByID(id)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}
	ref.SetLink()

	usersCount, err := tb.store.Referral.CountUsers(ref.ID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	var activeUsersPercent int
	activeUsersCount, err := tb.store.Referral.ActiveUsers(ref.Code)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}
	if usersCount != 0 {
		activeUsersPercent = (activeUsersCount / usersCount) * 100
	}

	premiumUsersCount := 0

	var deadUsersPercent int
	deadUsersCount := usersCount - activeUsersCount
	if deadUsersCount != 0 {
		deadUsersPercent = (deadUsersCount / usersCount) * 100
	}

	runMiniApp, err := tb.store.Referral.RunMiniApp(ref.Code)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return

	}
	notRunMiniApp := usersCount - runMiniApp

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("🔗<b>Управление ссылкой</b>\n|-Ссылка: <b>%s <a href='%s'>ссылка</a></b>\n|-Название: <b>%s</b>\n\n📊<b>Статистика ссылки</b>\n|-Всего переходов: <b>%d</b>\n|-Уникальных: <b>%d</b> (100%%)\n\n👥<b>Статистика пользователей</b>\n|-Всего: <b>%d</b>\n|-Активных: <b>%d</b> (%d %%)\n|-Мертвых: <b>%d</b> (%d %%)\n|-Премиум: <b>%d</b> (%d %%)\n|-RTL: <b>%d</b> (%d %%)\n\n🚪<b>Статистика проходимости</b>\n|-Запуски: <b>%d</b>\n|-Выполнили бонусов <b>%d</b>\n|-Ушли после /start: <b>%d</b>",
			ref.Code, ref.Link, ref.Name, usersCount, usersCount, usersCount, activeUsersCount, activeUsersPercent, deadUsersCount, deadUsersPercent, premiumUsersCount, premiumUsersCount, 0, 0, runMiniApp, 0, notRunMiniApp),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}

func (tb tgBot) referralCreate(ctx context.Context, b *bot.Bot, update *models.Update) {
	err := tb.store.Referral.Create()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}
}

func (tb tgBot) referralDelete(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	idString := strings.Split(update.CallbackQuery.Data, "@")
	var id int
	if len(idString) > 1 {
		id, _ = strconv.Atoi(idString[1])
	}

	err := tb.store.Referral.Delete(id)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	tb.referralsPage(ctx, b, update)
}

func (tb tgBot) referralChangeName(name string, msg *models.Message) {
	err := tb.store.Referral.UpdateName(referralID, name)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), msg.From.ID, internalError)
		return

	}

}
func (tb tgBot) referralChangeCode(code string, msg *models.Message) {
	err := tb.store.Referral.UpdateCode(referralID, code)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), msg.From.ID, internalError)
		return
	}
}

func (tb tgBot) requests(ctx context.Context, b *bot.Bot, update *models.Update) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "За сегодня", CallbackData: requestsDaily},
				{Text: "За неделю", CallbackData: requestsWeekly},
				{Text: "За месяц", CallbackData: requestsMonthly},
			},
			{
				{Text: "За все время", CallbackData: requestsAll},
			},
			{
				{Text: "Назад", CallbackData: menu},
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

func (tb tgBot) requestsDaily(ctx context.Context, b *bot.Bot, update *models.Update) {
	msgs, err := tb.store.Message.RequestsDaily()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	kb := &models.InlineKeyboardMarkup{

		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "За сегодня", CallbackData: requestsDaily},
				{Text: "За неделю", CallbackData: requestsWeekly},
				{Text: "За месяц", CallbackData: requestsMonthly},
			},
			{
				{Text: "За все время", CallbackData: requestsAll},
			},
			{
				{Text: "Назад", CallbackData: menu},
				{Text: "В меню", CallbackData: menu},
			},
		},
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("📊<b>Cтатистика запросов к нейросетям</b>\n\n<pre><code>ChatGPT 4o-mini: %d\nChatGPT o1-mini: %d\nChatGPT o1-preview: %d\nChatGPT 4o: %d\nDall-e-3: %d\nRunware: %d</code></pre>",
			msgs["gpt-4o-mini"], msgs["o1-mini"], msgs["o1-preview"], msgs["gpt-4o"], msgs["dall-e-3"], msgs["runware"]),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}
func (tb tgBot) requestsWeekly(ctx context.Context, b *bot.Bot, update *models.Update) {
	msgs, err := tb.store.Message.RequestsWeekly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	kb := &models.InlineKeyboardMarkup{

		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "За сегодня", CallbackData: requestsDaily},
				{Text: "За неделю", CallbackData: requestsWeekly},
				{Text: "За месяц", CallbackData: requestsMonthly},
			},
			{
				{Text: "За все время", CallbackData: requestsAll},
			},
			{
				{Text: "Назад", CallbackData: menu},
				{Text: "В меню", CallbackData: menu},
			},
		},
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("📊<b>Cтатистика запросов к нейросетям</b>\n\n<pre><code>ChatGPT 4o-mini: %d\nChatGPT o1-mini: %d\nChatGPT o1-preview: %d\nChatGPT 4o: %d\nDall-e-3: %d\nRunware: %d</code></pre>",
			msgs["gpt-4o-mini"], msgs["o1-mini"], msgs["o1-preview"], msgs["gpt-4o"], msgs["dall-e-3"], msgs["runware"]),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})

}
func (tb tgBot) requestsMonthly(ctx context.Context, b *bot.Bot, update *models.Update) {
	msgs, err := tb.store.Message.RequestsMontly()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}
	kb := &models.InlineKeyboardMarkup{

		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "За сегодня", CallbackData: requestsDaily},
				{Text: "За неделю", CallbackData: requestsWeekly},
				{Text: "За месяц", CallbackData: requestsMonthly},
			},
			{
				{Text: "За все время", CallbackData: requestsAll},
			},
			{
				{Text: "Назад", CallbackData: menu},
				{Text: "В меню", CallbackData: menu},
			},
		},
	}
	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("📊<b>Cтатистика запросов к нейросетям</b>\n\n<pre><code>ChatGPT 4o-mini: %d\nChatGPT o1-mini: %d\nChatGPT o1-preview: %d\nChatGPT 4o: %d\nDall-e-3: %d\nRunware: %d</code></pre>",
			msgs["gpt-4o-mini"], msgs["o1-mini"], msgs["o1-preview"], msgs["gpt-4o"], msgs["dall-e-3"], msgs["runware"]),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})

}
func (tb tgBot) requestsAll(ctx context.Context, b *bot.Bot, update *models.Update) {
	msgs, err := tb.store.Message.RequestsAll()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(context.Background(), update.CallbackQuery.From.ID, internalError)
		return
	}

	kb := &models.InlineKeyboardMarkup{

		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "За сегодня", CallbackData: requestsDaily},
				{Text: "За неделю", CallbackData: requestsWeekly},
				{Text: "За месяц", CallbackData: requestsMonthly},
			},
			{
				{Text: "За все время", CallbackData: requestsAll},
			},
			{
				{Text: "Назад", CallbackData: menu},
				{Text: "В меню", CallbackData: menu},
			},
		},
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.From.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("📊<b>Cтатистика запросов к нейросетям</b>\n\n<pre><code>ChatGPT 4o-mini: %d\nChatGPT o1-mini: %d\nChatGPT o1-preview: %d\nChatGPT 4o: %d\nDall-e-3: %d\nRunware: %d</code></pre>",
			msgs["gpt-4o-mini"], msgs["o1-mini"], msgs["o1-preview"], msgs["gpt-4o"], msgs["dall-e-3"], msgs["runware"]),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
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
	case statisticsWeekly:
		tb.statisticsWeekly(ctx, b, update)
	case statisticsMonthly:
		tb.statisticsMonthly(ctx, b, update)
	case statisticsAll:
		tb.statisticsAll(ctx, b, update)
	case statisticsBack:
		tb.adminMenu(ctx, b, update)
	case menu:
		tb.adminMenu(ctx, b, update)

	// bonuses
	case bonuses:
		tb.bonuses(ctx, b, update)
	case bonusesCreate:
		tb.bonusCreate(ctx, b, update)
		tb.bonuses(ctx, b, update)
	case bonusesChangeChannelNamePost:
		tb.f.Transition(update.CallbackQuery.From.ID, stateChannelNamePost, update.CallbackQuery)
	case bonusesChangeChannelNameHand:
		tb.f.Transition(update.CallbackQuery.From.ID, stateChannelNameHand, update.CallbackQuery)
	case bonusesDelete:
		tb.bonusDelete(ctx, b, update)
		tb.bonuses(ctx, b, update)
	case bonusCheckTrue:
		tb.bonusCheck(true, update)
		tb.bonuses(ctx, b, update)
	case bonusCheckFalse:
		tb.bonusCheck(false, update)
		tb.bonuses(ctx, b, update)
	case bonusesChangeName:
		tb.f.Transition(update.CallbackQuery.From.ID, stateBonusName, update.CallbackQuery)
	case bonusesChangeMaxUsers:
		tb.f.Transition(update.CallbackQuery.From.ID, stateBonusMaxUsers, update.CallbackQuery)
	case bonusesChangeAward:
		tb.f.Transition(update.CallbackQuery.From.ID, stateBonusAward, update.CallbackQuery)
	// users
	case usersLimits:
		tb.f.Transition(update.CallbackQuery.From.ID, stateUserLimitsModel, update.CallbackQuery)
	case usersPremium:
		tb.f.Transition(update.CallbackQuery.From.ID, stateUserPremium, update.CallbackQuery)
	case usersDiamonds:
		tb.f.Transition(update.CallbackQuery.From.ID, stateUserDiamonds, update.CallbackQuery)

	case referralsCreate:
		tb.referralCreate(ctx, b, update)
		tb.referralsPage(ctx, b, update)

	case referralsChangeName:
		tb.f.Transition(update.CallbackQuery.From.ID, stateReferralName, update.CallbackQuery)
	case referralsChangeCode:
		tb.f.Transition(update.CallbackQuery.From.ID, stateReferralCode, update.CallbackQuery)

	case requests:
		tb.requests(ctx, b, update)
	case requestsDaily:
		tb.requestsDaily(ctx, b, update)
	case requestsWeekly:
		tb.requestsWeekly(ctx, b, update)
	case requestsMonthly:
		tb.requestsMonthly(ctx, b, update)
	case requestsAll:
		tb.requestsAll(ctx, b, update)

	}
}

func (tb tgBot) callbackChannelNamePost(f *fsm.FSM, args ...any) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Ручное добавление", CallbackData: bonusesChangeChannelNameHand},
				{Text: "Отмена", CallbackData: bonuses},
			},
		},
	}

	callbackQuery := args[0].(*models.CallbackQuery)
	tb.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:      callbackQuery.From.ID,
		MessageID:   callbackQuery.Message.Message.ID,
		Text:        "Перешлите мне пост от спонсора, либо выберите ручное добавление",
		ReplyMarkup: kb,
	})
}

func (tb tgBot) callbackChannelNameHand(f *fsm.FSM, args ...any) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Отмена", CallbackData: bonuses},
			},
		},
	}

	callbackQuery := args[0].(*models.CallbackQuery)
	tb.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:      callbackQuery.From.ID,
		MessageID:   callbackQuery.Message.Message.ID,
		Text:        "Напиши ID спонсора",
		ReplyMarkup: kb,
	})
}

func (tb tgBot) callbackChannelLink(f *fsm.FSM, args ...any) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Отмена", CallbackData: bonuses},
			},
		},
	}

	chatID := args[0]
	tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "🔗 Отправьте пригласительную ссылку",
		ReplyMarkup: kb,
	})

}

func (tb tgBot) callbackBonusName(f *fsm.FSM, args ...any) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Отмена", CallbackData: bonuses},
			},
		},
	}

	callbackQuery := args[0].(*models.CallbackQuery)
	tb.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:      callbackQuery.From.ID,
		MessageID:   callbackQuery.Message.Message.ID,
		Text:        "Напишите название бонуса",
		ReplyMarkup: kb,
	})
}

func (tb tgBot) callbackBonusAward(f *fsm.FSM, args ...any) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Отмена", CallbackData: bonuses},
			},
		},
	}

	callbackQuery := args[0].(*models.CallbackQuery)
	tb.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:      callbackQuery.From.ID,
		MessageID:   callbackQuery.Message.Message.ID,
		Text:        "Введите награду",
		ReplyMarkup: kb,
	})
}

func (tb tgBot) callbackBonusMaxUsers(f *fsm.FSM, args ...any) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Отмена", CallbackData: bonuses},
			},
		},
	}

	callbackQuery := args[0].(*models.CallbackQuery)
	tb.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:      callbackQuery.From.ID,
		MessageID:   callbackQuery.Message.Message.ID,
		Text:        "Напишите максимальное число пользователей",
		ReplyMarkup: kb,
	})
}

func (tb tgBot) callbackUserLimitsModel(f *fsm.FSM, args ...any) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: gpt4o, CallbackData: gpt4o},
				{Text: gpt4o_mini, CallbackData: gpt4o_mini},
			},
			{
				{Text: o1_mini, CallbackData: o1_mini},
				{Text: o1_preview, CallbackData: o1_preview},
			},
			{
				{Text: dall_e_3, CallbackData: dall_e_3},
				{Text: runware, CallbackData: runware},
			},
		},
	}

	callbackQuery := args[0].(*models.CallbackQuery)
	tb.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:      callbackQuery.From.ID,
		MessageID:   callbackQuery.Message.Message.ID,
		Text:        "Choose",
		ReplyMarkup: kb,
	})
}

func (tb tgBot) callbackUserLimitsAmount(f *fsm.FSM, args ...any) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Отмена", CallbackData: usersPage + "1"},
			},
		},
	}

	callbackQuery := args[0].(*models.CallbackQuery)
	tb.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:      callbackQuery.From.ID,
		MessageID:   callbackQuery.Message.Message.ID,
		Text:        "Количество лимитов",
		ReplyMarkup: kb,
	})

}

func (tb tgBot) callbackUserPremium(f *fsm.FSM, args ...any) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Advanced (месяц)", CallbackData: advancedMonth},
				{Text: "Advanced (год)", CallbackData: advancedYear},
			},
			{
				{Text: "Ultimate (месяц)", CallbackData: ultimateMonth},
				{Text: "Ultimate (год)", CallbackData: ultimateYear},
			},
		},
	}

	callbackQuery := args[0].(*models.CallbackQuery)
	tb.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:      callbackQuery.From.ID,
		MessageID:   callbackQuery.Message.Message.ID,
		Text:        "Выберите подписку",
		ReplyMarkup: kb,
	})
}

func (tb tgBot) callbackUserDiamonds(f *fsm.FSM, args ...any) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Отмена", CallbackData: menu},
			},
		},
	}

	callbackQuery := args[0].(*models.CallbackQuery)
	tb.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:      callbackQuery.From.ID,
		MessageID:   callbackQuery.Message.Message.ID,
		Text:        "Напишите количество алмазов",
		ReplyMarkup: kb,
	})

}

func (tb tgBot) callbackReferralName(f *fsm.FSM, args ...any) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Отмена", CallbackData: menu},
			},
		},
	}

	callbackQuery := args[0].(*models.CallbackQuery)
	tb.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:      callbackQuery.From.ID,
		MessageID:   callbackQuery.Message.Message.ID,
		Text:        "Напишите название",
		ReplyMarkup: kb,
	})

}

func (tb tgBot) callbackReferralCode(f *fsm.FSM, args ...any) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Отмена", CallbackData: menu},
			},
		},
	}

	callbackQuery := args[0].(*models.CallbackQuery)
	tb.b.EditMessageText(context.Background(), &bot.EditMessageTextParams{
		ChatID:      callbackQuery.From.ID,
		MessageID:   callbackQuery.Message.Message.ID,
		Text:        "Напишите название",
		ReplyMarkup: kb,
	})

}
