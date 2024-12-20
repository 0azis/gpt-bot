package server

import (
	"errors"
	"gpt-bot/internal/api"
	"gpt-bot/internal/db"
	"gpt-bot/internal/server/controllers"
	"gpt-bot/tgbot"
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
)

// InitRouter init all routes of all groups
func InitRoutes(e *echo.Echo, store db.Store, api api.Interface, b tgbot.BotInterface, savePath string) {
	apiRoute := e.Group("/api/v1") // basic api route
	apiRoute.Static("/uploads", savePath)

	userRoutes(apiRoute, store)
	chatRoutes(apiRoute, store)
	messageRoutes(apiRoute, store, api, savePath)
	paymentRoutes(apiRoute, store, b, api)
	imageRoutes(apiRoute, savePath, b)
	bonusRoutes(apiRoute, store, b)
}

func userRoutes(apiRoute *echo.Group, store db.Store) {
	user := apiRoute.Group("/users", AuthMiddleware) // basic user route
	controller := controllers.NewUserControllers(store)

	user.GET("", controller.GetUser)
}

func chatRoutes(apiRoute *echo.Group, store db.Store) {
	chat := apiRoute.Group("/chats", AuthMiddleware)
	controller := controllers.NewChatControllers(store)

	chat.GET("", controller.GetChats)
	chat.DELETE("/:id", controller.Delete)
}

func messageRoutes(apiRoute *echo.Group, store db.Store, api api.Interface, savePath string) {
	message := apiRoute.Group("/messages", AuthMiddleware)
	controller := controllers.NewMessageControllers(api, store, savePath)

	message.GET("/chat/:id", controller.GetMessages)
	message.POST("", controller.NewMessage)
	message.PUT("/:id", controller.ResendMessage)
}

func paymentRoutes(apiRoute *echo.Group, store db.Store, b tgbot.BotInterface, api api.Interface) {
	payment := apiRoute.Group("/payment")
	yoomoney := payment.Group("/yoomoney")
	controller := controllers.NewPaymentControllers(store, b, api)

	payment.POST("", controller.CreateInvoiceLink, AuthMiddleware)
	payment.POST("/webhook", controller.Webhook)

	yoomoney.POST("/auth", controller.AuthYooMoney)
	yoomoney.POST("/token", controller.TokenYooMoney)
	yoomoney.POST("/pay", controller.CreatePayment)
}

func imageRoutes(apiRoute *echo.Group, savePath string, b tgbot.BotInterface) {
	if _, err := os.Stat(savePath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(savePath, os.ModePerm)
		if err != nil {
			slog.Error(err.Error())
		}
	}
	image := apiRoute.Group("/image", AuthMiddleware)
	controller := controllers.NewImageControllers(savePath, b)

	image.POST("", controller.UploadImage)
	image.POST("/send", controller.SendImageToTelegram)
}

func bonusRoutes(apiRoute *echo.Group, store db.Store, b tgbot.BotInterface) {
	bonus := apiRoute.Group("/bonus", AuthMiddleware)
	controller := controllers.NewBonusControllers(store, b)

	bonus.POST("/:id", controller.GetAward)
	bonus.GET("", controller.GetAll)
}
