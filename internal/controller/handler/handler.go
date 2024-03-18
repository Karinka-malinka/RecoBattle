package handler

import (
	"fmt"

	"github.com/RecoBattle/internal/app/userapp"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler interface {
	RegisterHandler(*echo.Echo, *echo.Group, *echo.Group)
}

func GetUserID(c echo.Context) (string, error) {

	var userID string

	user := c.Get("user")

	if user != nil {
		u := user.(*jwt.Token)
		claims := u.Claims.(*userapp.JWTCustomClaims)
		userID = claims.UserID
		return userID, nil
	} else {
		u := c.Get("userID").(uuid.UUID)
		if u != uuid.Nil {
			return u.String(), nil
		}
	}

	return "", fmt.Errorf("no token")
}
