package handler

import "github.com/labstack/echo/v4"

type Handler interface {
	RegisterHandler(*echo.Echo, *echo.Group)
}
