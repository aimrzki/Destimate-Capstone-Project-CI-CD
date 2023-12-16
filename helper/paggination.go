package helper

import (
	"github.com/labstack/echo/v4"
	"strconv"
)

func GetPaginationParams(c echo.Context) (int, int) {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(c.QueryParam("per_page"))
	if err != nil || perPage < 1 {
		perPage = 10 // Jika tidak ada nilai per_page, atur default ke 10
	}

	return page, perPage
}
