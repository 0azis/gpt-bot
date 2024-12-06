package tgbot

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"gpt-bot/config"
	"gpt-bot/internal/db"
	"gpt-bot/internal/db/domain"
	"gpt-bot/utils"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/0azis/bot"
	"github.com/0azis/bot/models"
	"github.com/go-telegram/fsm"
)

type BotInterface interface {
	Start()
	Instance() *bot.Bot
	InitHandlers()

	// handlers
	startHandler(ctx context.Context, b *bot.Bot, update *models.Update)
	defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update)

	// helpers
	CreateInvoiceLink(payload []byte, paymentCredentials domain.Payment) (string, error)
	IsUserMember(channelID int, userID int) bool
	PaymentInfo(paymentCredentials domain.Payment, status bool)
	SendImage(imageLink string, userID int)
	GetChannelInfo(channelID int) (domain.Channel, error)
	getTelegramAvatar(ctx context.Context, userID int64) string
	informUser(ctx context.Context, userID int64, errMsg string)

	// admin
	adminHandler(ctx context.Context, b *bot.Bot, update *models.Update)
}

type tgBot struct {
	telegram config.Telegram
	store    db.Store
	b        *bot.Bot
	f        *fsm.FSM
}

func New(cfg config.Telegram, store db.Store) (BotInterface, error) {
	b, err := bot.New(cfg.GetToken())
	b.SetMyDescription(context.Background(), &bot.SetMyDescriptionParams{
		Description: "✨ Я - умная нейросеть, которая поможет тебе справиться с любой задачей!\n\n🔥Вы можете воспользоваться самыми современными инструментами:\n\n- DALLE + Runware + Midjourney: Генерация изображений на основе текста!\n\n- ChatGPT : получение ответов на вопросы, советы и помощь в разговоре на любую тему!\n\n➕ Редактор изображений, создание видео, увеличенная скорость ответа и генераций и многое другое.",
	})
	b.SetMyCommands(context.Background(), &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "/start", Description: "Главная страница"},
			{Command: "/app", Description: "Информация о приложении"},
			{Command: "/help", Description: "Техническая поддержка"},
			{Command: "/menu", Description: "Информация об аккаунте"},
		},
	})
	tgBot := tgBot{
		telegram: cfg,
		store:    store,
		b:        b,
	}

	f := fsm.New(stateDefault, map[fsm.StateID]fsm.Callback{
		stateChannelNamePost: tgBot.callbackChannelNamePost,
		stateChannelNameHand: tgBot.callbackChannelNameHand,
		stateChannelLink:     tgBot.callbackChannelLink,
		stateBonusName:       tgBot.callbackBonusName,
		stateBonusMaxUsers:   tgBot.callbackBonusMaxUsers,
		stateBonusAward:      tgBot.callbackBonusAward,

		stateUserLimitsModel:  tgBot.callbackUserLimitsModel,
		stateUserLimitsAmount: tgBot.callbackUserLimitsAmount,
		stateUserPremium:      tgBot.callbackUserPremium,
		stateUserDiamonds:     tgBot.callbackUserDiamonds,

		stateReferralName: tgBot.callbackReferralName,
		stateReferralCode: tgBot.callbackReferralCode,
	})
	tgBot.f = f
	return tgBot, err
}

func (tb tgBot) Start() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	tb.b.Start(ctx)
}

func (tb tgBot) Instance() *bot.Bot {
	return tb.b
}

func (tb tgBot) InitHandlers() {
	// base
	tb.b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeContains, tb.startHandler)
	tb.b.SetDefaultHandler(tb.defaultHandler)

	// admin
	tb.b.RegisterHandler(bot.HandlerTypeMessageText, "/admin", bot.MatchTypeContains, tb.adminHandler)
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "btn_", bot.MatchTypePrefix, tb.callbackHandler)
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "id@", bot.MatchTypePrefix, tb.bonusInfo)
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "page@", bot.MatchTypePrefix, tb.usersPage)
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "prempage@", bot.MatchTypePrefix, tb.premiumUsersPage)
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "user@", bot.MatchTypePrefix, tb.userSingle)
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "admin@", bot.MatchTypePrefix, tb.usersAdmin)
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "model@", bot.MatchTypePrefix, tb.usersLimitsModel)
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "sub@", bot.MatchTypePrefix, tb.usersPremium)
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "referral@", bot.MatchTypePrefix, tb.referralsPage)
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "ref@", bot.MatchTypePrefix, tb.referralSingle)
	tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "del@", bot.MatchTypePrefix, tb.referralDelete)

	// tb.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "mini_app", bot.MatchTypePrefix, tb.cb)

	// other
	tb.b.RegisterHandler(bot.HandlerTypeMessageText, "/app", bot.MatchTypeExact, tb.appHandler)
	tb.b.RegisterHandler(bot.HandlerTypeMessageText, "/menu", bot.MatchTypeExact, tb.menuHandler)
	tb.b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, tb.helpHandler)

}

func (tb tgBot) startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var user domain.User
	user.ID = int(update.Message.From.ID)
	user.Avatar = tb.getTelegramAvatar(ctx, int64(user.ID))
	user.LanguageCode = update.Message.From.LanguageCode

	err := tb.store.User.Create(user)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(user.ID), userCreationError)
		return
	}
	err = tb.store.Subscription.InitStandard(user.ID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(user.ID), userCreationError)
		return
	}
	limits := domain.NewLimits(user.ID, domain.SubscriptionStandard)
	err = tb.store.Limits.Create(limits)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(user.ID), userCreationError)
		return
	}
	err = tb.store.Bonus.InitBonuses(user.ID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(user.ID), userCreationError)
		return
	}

	token := utils.NewToken()
	token.SetUserID(user.ID)
	err = token.SignJWT()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(user.ID), userCreationError)
		return
	}

	msgSlice := strings.Split(update.Message.Text, " ")
	if len(msgSlice) == 2 {
		refCode := &msgSlice[1]
		adRes, err := tb.store.Referral.GetOne(*refCode)
		if err != nil {
			slog.Error(err.Error())
		}

		userRes, err := tb.store.User.OwnerOfReferralCode(*refCode)
		if err != nil {
			slog.Error(err.Error())
		}

		if userRes != 0 {
			id, err := tb.store.User.IsUserReferred(user.ID)
			if err != nil {
				slog.Error(err.Error())
			}
			if id != 0 {
				ownerID, err := tb.store.User.OwnerOfReferralCode(*refCode)
				if errors.Is(err, sql.ErrNoRows) {
					tb.informUser(ctx, int64(user.ID), referralInvalid)
					return
				}
				if ownerID == user.ID {
					tb.informUser(ctx, int64(user.ID), referralSameUser)
					return
				}
				if err != nil {
					slog.Error(err.Error())
					tb.informUser(ctx, int64(user.ID), internalError)
					return
				}

				award := domain.ReferralAward
				err = tb.store.User.RaiseBalance(ownerID, award)
				if err != nil {
					slog.Error(err.Error())
					tb.informUser(ctx, int64(user.ID), internalError)
					return
				}
				err = tb.store.User.SetReferredBy(user.ID, *refCode)
				if err != nil {
					slog.Error(err.Error())
					tb.informUser(ctx, int64(user.ID), internalError)
					return
				}
			}
		}
		if adRes != 0 {
			refID, err := tb.store.Referral.GetOne(*refCode)
			if errors.Is(err, sql.ErrNoRows) {
				tb.informUser(ctx, int64(user.ID), referralInvalid)
				return
			}
			if err != nil {
				slog.Error(err.Error())
				tb.informUser(ctx, int64(user.ID), internalError)
				return
			}
			tb.store.Referral.AddUser(user.ID, refID)
		}
	}

	_, err = b.SetChatMenuButton(ctx, &bot.SetChatMenuButtonParams{
		ChatID: update.Message.Chat.ID,
		MenuButton: models.MenuButtonWebApp{
			Type: "web_app",
			Text: "Open Mini App",
			WebApp: models.WebAppInfo{
				URL: tb.telegram.GetWebAppUrl() + "?token=" + token.GetStrToken(),
			},
		},
	})
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, miniAppError)
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

	file, err := os.ReadFile("../assets/preview.png")
	if err != nil {
		slog.Error(err.Error())
	}

	b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID: update.Message.Chat.ID,
		Photo: &models.InputFileUpload{
			Filename: "preview.png",
			Data:     bytes.NewReader(file),
		},
		Caption:     "<b>👋 Добро пожаловать в оригинальное приложение для использования нейросетей!</b>\n\nБесплатные нейросети с ежедневным лимитом: <code>ChatGPT + Runware.</code>\n\n<b>Нейросети, с которыми работает WebAi:</b> <code>ChatGPT, Midjourney, DALLE, Runway, Runware, Suno, Gemini.</code>\n\n<b>Полезные команды:</b>\n/help - техническая поддержка.\n/menu - управление приложением.\n/app - открыть мини-приложение.\n/start - перезапустить приложение.\n\n<i>Для запуска приложения нажмите кнопку ниже...</i>",
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}

func (tb tgBot) appHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	token := utils.NewToken()
	token.SetUserID(int(update.Message.From.ID))
	err := token.SignJWT()
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, update.Message.From.ID, userCreationError)
		return
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Открыть приложение", CallbackData: "mini_app", WebApp: &models.WebAppInfo{
					URL: tb.telegram.GetWebAppUrl() + "?token=" + token.GetStrToken(),
				}},
			},
		},
	}
	file, err := os.ReadFile("../assets/preview.png")
	if err != nil {
		slog.Error(err.Error())
	}

	b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID: update.Message.From.ID,
		Photo: &models.InputFileUpload{
			Filename: "preview.png",
			Data:     bytes.NewReader(file),
		},
		Caption:     "<b>👋 Добро пожаловать в оригинальное приложение для использования нейросетей!</b>\n\nБесплатные нейросети с ежедневным лимитом: <b>ChatGPT + Runware.</b>\n\n<i>Для запуска приложения нажмите кнопку ниже...</i>",
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}

// func (tb tgBot) cb(ctx context.Context, b *bot.Bot, update *models.Update) {
// 	fmt.Println("HELLO")

// 	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
// 		CallbackQueryID: update.CallbackQuery.ID,
// 	})

// 	// b.AnswerWebAppQuery(ctx, &bot.AnswerWebAppQueryParams{
// 	// 	Result: models.
// 	// })
// 	tb.store.Stat.Count()
// }

func (tb tgBot) menuHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := int(update.Message.From.ID)
	subscriptionName, err := tb.store.Subscription.GetSubscription(userID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(userID), internalError)
		return

	}
	limits, err := tb.store.Limits.GetByUser(userID)
	if err != nil {
		slog.Error(err.Error())
		tb.informUser(ctx, int64(userID), internalError)
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    userID,
		Text:      fmt.Sprintf("ID: %d\nПодписка: <i>%s</i>\n\n<pre><code><b>Лимиты запросов:</b>\nGPT o1: %d\nGPT o1-mini: %d\nGPT 4o: %d\nGPT 4o-mini: %d\nDALL-E 3: %d\nRunware: %d</code></pre>\n\n<i>Лимиты для пользователей обновляются каждый день</i>", userID, subscriptionName, limits.O1Preview, limits.O1Mini, limits.Gpt4o, limits.Gpt4oMini, limits.Dalle3, limits.Runware),
		ParseMode: models.ParseModeHTML,
	})
}

func (tb tgBot) helpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Support", URL: "https://t.me/WebAiSupport"},
			},
		},
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.From.ID,
		Text:        "<b>Возникла проблема? Есть предложения по улучшению проекта? Вопрос сотрудничества? Техническая поддержка ответит на все вопросы 😉.</b>\n\n🛠 Support - @WebAiSupport",
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}

func (tb tgBot) getTelegramAvatar(ctx context.Context, userID int64) string {
	var url string

	// get list of photos
	photos, _ := tb.b.GetUserProfilePhotos(ctx, &bot.GetUserProfilePhotosParams{
		UserID: userID,
	})

	if photos.TotalCount == 0 {
		url = "" // if user don't have an avatar
	} else {
		file, _ := tb.b.GetFile(ctx, &bot.GetFileParams{
			FileID: photos.Photos[0][0].FileID, // take the first avatar
		})
		url = tb.b.FileDownloadLink(file) // generate link
	}

	return url
}

func (tb tgBot) CreateInvoiceLink(payload []byte, payment domain.Payment) (string, error) {
	link, err := tb.b.CreateInvoiceLink(context.Background(), &bot.CreateInvoiceLinkParams{
		Title:       fmt.Sprintf("%s subscription", payment.SubscriptionName),
		Description: "Buy subscription",
		Payload:     string(payload),
		Currency:    "XTR",
		Prices: []models.LabeledPrice{
			{Label: payment.SubscriptionName, Amount: payment.Amount},
		},
	})
	return link, err
}

func (tb tgBot) IsUserMember(channelID int, userID int) bool {
	member, _ := tb.b.GetChatMember(context.Background(), &bot.GetChatMemberParams{
		ChatID: channelID,
		UserID: int64(userID),
	})

	if member.Member == nil {
		return false
	}

	return true
}

func (tb tgBot) SendImage(imageLink string, userID int) {
	tb.b.SendPhoto(context.Background(), &bot.SendPhotoParams{
		ChatID: userID,
		Photo: &models.InputFileString{
			Data: imageLink,
		},
	})
}

func (tb tgBot) GetChannelInfo(channelID int) (domain.Channel, error) {
	var channel domain.Channel
	channel.ID = channelID

	tgChannel, err := tb.b.GetChat(context.Background(), &bot.GetChatParams{
		ChatID: channel.ID,
	})
	if err != nil {
		return channel, err
	}
	channel.Title = tgChannel.Title

	var url string
	if tgChannel.Photo != nil {
		file, err := tb.b.GetFile(context.Background(), &bot.GetFileParams{
			FileID: tgChannel.Photo.SmallFileID,
		})
		if err != nil {
			return channel, err
		}

		url = tb.b.FileDownloadLink(file)
	}

	channel.Avatar = url
	return channel, nil
}

func (tb tgBot) informUser(ctx context.Context, userID int64, errorMsg string) {
	tb.b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: userID,
		Text:   fmt.Sprintf("%s", errorMsg),
	})
}

func (tb tgBot) PaymentInfo(payment domain.Payment, status bool) {
	if status {
		tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID:    payment.UserID,
			Text:      fmt.Sprintf(`Вы успешно купили подписку <b>%s</b> за <i>%d</i> звезд. Ваша подписка доступна до: %s`, payment.SubscriptionName, payment.Amount, payment.End),
			ParseMode: models.ParseModeHTML,
		})
	} else {
		tb.b.SendMessage(context.Background(), &bot.SendMessageParams{
			ChatID:    payment.UserID,
			Text:      fmt.Sprintf(`Неудалось купить подписку <b>%s</b> за <i>%d</i> звезд. Попробуйте позже`, payment.SubscriptionName, payment.Amount),
			ParseMode: models.ParseModeHTML,
		})
	}
}

func (tb tgBot) defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var payment domain.Payment

	if update.PreCheckoutQuery != nil {
		err := json.Unmarshal([]byte(update.PreCheckoutQuery.InvoicePayload), &payment)
		if err != nil {
			slog.Error(err.Error())
			return
		}

		_, err = b.AnswerPreCheckoutQuery(ctx, &bot.AnswerPreCheckoutQueryParams{
			PreCheckoutQueryID: update.PreCheckoutQuery.ID,
			OK:                 true,
			ErrorMessage:       "",
		})
		if err != nil {
			slog.Error(err.Error())
			tb.informUser(ctx, int64(payment.UserID), paymentError)
			return
		}

		return
	}

	if update.Message != nil {
		if update.Message.SuccessfulPayment != nil {
			err := json.Unmarshal([]byte(update.Message.SuccessfulPayment.InvoicePayload), &payment)
			if err != nil {
				slog.Error(err.Error())
				tb.informUser(ctx, update.Message.From.ID, internalError)
				return
			}
			err = tb.store.Subscription.Update(payment.UserID, payment.SubscriptionName, payment.End)
			if err != nil {
				slog.Error(err.Error())
				tb.informUser(ctx, int64(payment.UserID), internalError)
				return
			}

			diamonds, err := tb.store.Subscription.DailyDiamonds(payment.SubscriptionName)
			if err != nil {
				slog.Error(err.Error())
				tb.informUser(ctx, int64(payment.UserID), internalError)
				return
			}
			err = tb.store.User.FillBalance(payment.UserID, diamonds)
			if err != nil {
				slog.Error(err.Error())
				tb.informUser(ctx, int64(payment.UserID), internalError)
				return
			}
			limits := domain.NewLimits(payment.UserID, payment.SubscriptionName)
			err = tb.store.Limits.Update(limits)
			if err != nil {
				slog.Error(err.Error())
				tb.informUser(ctx, int64(payment.UserID), internalError)
				return
			}
			tb.PaymentInfo(payment, true)
			return
		}
		if update.Message.Text != "" {
			switch tb.f.Current(update.Message.From.ID) {
			case stateChannelNameHand:
				// bonusScheme.channelName = update.Message.ForwardOrigin.MessageOriginChannel.Chat.ID
				tb.f.Transition(update.Message.From.ID, stateChannelLink, update.Message.From.ID)
			case stateChannelLink:
				bonusScheme.link = update.Message.Text
				tb.bonusUpdate(bonusScheme, update.Message)
				tb.bonuses(ctx, b, update)
				tb.f.Transition(update.Message.From.ID, stateDefault)
			case stateBonusName:
				bonusScheme.name = update.Message.Text
				tb.bonusName(bonusScheme, update.Message)
				tb.bonuses(ctx, b, update)
			case stateBonusMaxUsers:
				value, err := strconv.Atoi(update.Message.Text)
				if err != nil {
				}
				bonusScheme.maxUsers = value
				tb.bonusMaxUsers(bonusScheme, update.Message)
				tb.bonuses(ctx, b, update)
			case stateBonusAward:
				value, err := strconv.Atoi(update.Message.Text)
				if err != nil {
				}
				tb.bonusAward(value, update.Message)
				tb.bonuses(ctx, b, update)
			case stateUserLimitsAmount:
				value, err := strconv.Atoi(update.Message.Text)
				if err != nil {
				}
				uLimits.amount = value
				tb.usersLimits(uLimits, update.Message)
				tb.f.Transition(update.Message.From.ID, stateDefault)
			case stateUserDiamonds:
				diamonds, err := strconv.Atoi(update.Message.Text)
				if err != nil {
				}
				tb.usersDiamonds(diamonds, update.Message)
				tb.f.Transition(update.Message.From.ID, stateDefault)
				tb.adminHandler(ctx, b, update)

			case stateReferralName:
				tb.referralChangeName(update.Message.Text, update.Message)
				tb.f.Transition(update.Message.From.ID, stateDefault)
				tb.referralsPage(ctx, b, update)
			case stateReferralCode:
				tb.referralChangeCode(update.Message.Text, update.Message)
				tb.f.Transition(update.Message.From.ID, stateDefault)
				tb.referralsPage(ctx, b, update)

			}
		}
		if update.Message.ForwardOrigin != nil {
			switch tb.f.Current(update.Message.From.ID) {
			case stateChannelNamePost:
				bonusScheme.channelID = update.Message.ForwardOrigin.MessageOriginChannel.Chat.ID
				tb.f.Transition(update.Message.From.ID, stateChannelLink, update.Message.From.ID)
			}
		}
	}
}

func (tb tgBot) IsBotBanned(chatID int64) bool {
	_, err := tb.b.GetChat(context.Background(), &bot.GetChatParams{
		ChatID: chatID,
	})
	return err != nil
}
