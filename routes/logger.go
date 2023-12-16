package routes

import (
	"github.com/labstack/echo/v4"
	"log"
	"time"
)

func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			end := time.Now()
			latency := end.Sub(start)
			latencyHuman := latency.String()
			status := c.Response().Status

			log.Printf("[%s] %s %s %s %d %s", end.Format("2006/01/02 - 15:04:05"), c.RealIP(), c.Request().Method, c.Path(), status, latencyHuman)

			return err
		}
	}
}
