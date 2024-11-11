package server

import (
	"gpt-bot/internal/api"
	"gpt-bot/internal/db"
	"gpt-bot/internal/server/controllers"

	"github.com/0azis/bot"
	"github.com/labstack/echo/v4"
)

// InitRouter init all routes of all groups
func InitRoutes(e *echo.Echo, store db.Store, api api.Interface, b *bot.Bot) {
	apiRoute := e.Group("/api/v1") // basic api route

	userRoutes(apiRoute, store)
	chatRoutes(apiRoute, store)
	messageRoutes(apiRoute, store, api)
	paymentRoutes(apiRoute, b)
}

func userRoutes(apiRoute *echo.Group, store db.Store) {
	user := apiRoute.Group("/users") // basic user route
	controller := controllers.NewUserControllers(store)

	user.GET("", controller.GetUser)
}

func chatRoutes(apiRoute *echo.Group, store db.Store) {
	chat := apiRoute.Group("/chats")
	controller := controllers.NewChatControllers(store)

	chat.GET("", controller.GetChats)
}

func messageRoutes(apiRoute *echo.Group, store db.Store, api api.Interface) {
	message := apiRoute.Group("/messages")
	controller := controllers.NewMessageControllers(api, store)

	message.POST("/chat", controller.NewMessage)
	message.POST("/chat/:id", controller.ChatMessage)
	message.GET("/chat/:id", controller.GetMessages)
}

func paymentRoutes(apiRoute *echo.Group, b *bot.Bot) {
	subscription := apiRoute.Group("/subscription")
	controller := controllers.NewSubscriptionControllers(b)

	subscription.POST("", controller.CreateInvoiceLink)
}
