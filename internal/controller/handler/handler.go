package handler

import (
	"github.com/RecoBattle/internal/app/userapp"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler interface {
	RegisterHandler(*echo.Echo, *echo.Group, *echo.Group)
}

func GetUserID(c echo.Context) uuid.UUID {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*userapp.JWTCustomClaims)
	return claims.UserID
}
