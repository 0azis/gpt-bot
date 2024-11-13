package server

import (
	"gpt-bot/internal/api"
	"gpt-bot/internal/db"
	"gpt-bot/internal/server/controllers"
	"gpt-bot/tgbot"

	"github.com/labstack/echo/v4"
)

// InitRouter init all routes of all groups
func InitRoutes(e *echo.Echo, store db.Store, api api.Interface, b tgbot.BotInterface) {
	apiRoute := e.Group("/api/v1") // basic api route

	userRoutes(apiRoute, store)
	chatRoutes(apiRoute, store)
	messageRoutes(apiRoute, store, api)
	paymentRoutes(apiRoute, b, api)
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
}

func messageRoutes(apiRoute *echo.Group, store db.Store, api api.Interface) {
	message := apiRoute.Group("/messages", AuthMiddleware)
	controller := controllers.NewMessageControllers(api, store)

	message.GET("/chat/:id", controller.GetMessages)
	message.POST("/chat", controller.NewMessage)
	message.POST("/chat/:id", controller.NewMessageToChat)
}

func paymentRoutes(apiRoute *echo.Group, b tgbot.BotInterface, api api.Interface) {
	subscription := apiRoute.Group("/subscription")
	controller := controllers.NewSubscriptionControllers(b, api)

	subscription.POST("", controller.CreateInvoiceLink, AuthMiddleware)
	subscription.POST("/webhook", controller.Webhook)
}
