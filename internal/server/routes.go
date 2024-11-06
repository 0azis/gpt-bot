package server

import (
	"gpt-bot/internal/db"
	"gpt-bot/internal/server/controllers"

	"github.com/labstack/echo/v4"
)

// InitRouter init all routes of all groups
func InitRoutes(e *echo.Echo, store db.Store) {
	api := e.Group("/api/v1") // basic api route

	userRoutes(api, store)
}

func userRoutes(api *echo.Group, store db.Store) {
	user := api.Group("/users") // basic user route
	controller := controllers.NewUserControllers(store)

	user.GET("", controller.GetUser)
}
