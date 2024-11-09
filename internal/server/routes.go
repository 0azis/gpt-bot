package server

import (
	"gpt-bot/internal/api"
	"gpt-bot/internal/db"
	"gpt-bot/internal/server/controllers"

	"github.com/labstack/echo/v4"
)

// InitRouter init all routes of all groups
func InitRoutes(e *echo.Echo, store db.Store, api api.ApiInterface) {
	apiRoute := e.Group("/api/v1") // basic api route

	userRoutes(apiRoute, store)
	chatRoutes(apiRoute, store)
	messageRoutes(apiRoute, store, api)
}

func userRoutes(apiRoute *echo.Group, store db.Store) {
	user := apiRoute.Group("/users") // basic user route
	controller := controllers.NewUserControllers(store)

	user.GET("", controller.GetUser)
	// user.GET("/referral", controller.GetReferralCode)
}

func chatRoutes(apiRoute *echo.Group, store db.Store) {
	chat := apiRoute.Group("/chats")
	controller := controllers.NewChatControllers(store)

	chat.POST("", controller.Create)
	chat.GET("", controller.GetChats)
}

func messageRoutes(apiRoute *echo.Group, store db.Store, api api.ApiInterface) {
	message := apiRoute.Group("/messages")
	controller := controllers.NewMessageControllers(api, store)

	message.POST("", controller.NewMessage)
	message.GET("/chat/:id", controller.GetMessages)
}
