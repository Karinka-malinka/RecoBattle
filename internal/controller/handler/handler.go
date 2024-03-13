package handler

import (
	"github.com/RecoBattle/internal/app/userapp"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

type Handler interface {
	RegisterHandler(*echo.Echo, *echo.Group, *echo.Group)
}

func GetUserID(c echo.Context) string {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*userapp.JWTCustomClaims)
	return claims.UserID
}
