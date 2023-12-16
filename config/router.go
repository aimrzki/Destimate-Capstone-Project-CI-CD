package config

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"myproject/routes"
)

func SetupRouter() *echo.Echo {
	db, err := InitializeDatabase()
	if err != nil {
		log.Fatal(err)
	}
	router := echo.New()
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Pre(middleware.RemoveTrailingSlash())
	routes.SetupRoutes(router, db)
	return router
}
