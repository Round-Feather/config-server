package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func Live(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"status": "available"})
}

func Ready(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"status": "available"})
}
