package handler

import (
	"fmt"
	"net/http"

	"github.com/RecoBattle/internal/app/userapp"
	"github.com/labstack/echo/v4"
)

type Handler interface {
	RegisterHandler(*echo.Echo, *echo.Group, *echo.Group)
}

func GetUserID(c echo.Context) (string, error) {

	userID := c.Get("user").(string)

	if userID == "" {
		return "", fmt.Errorf("no user id")
	}

	return userID, nil
}

func SendResponceToken(c echo.Context, response *userapp.LoginResponse) {

	c.Response().Header().Set("Authorization", "Bearer "+response.AccessToken)

	writeAccessTokenCookie(c, response.AccessToken)
	writeRefreshTokenCookie(c, response.RefreshToken)
}

func writeAccessTokenCookie(c echo.Context, accessToken string) {

	cookie := new(http.Cookie)

	cookie.Name = "access_token"
	cookie.Value = accessToken
	cookie.HttpOnly = true
	cookie.SameSite = 3
	cookie.Path = "/"

	c.SetCookie(cookie)
}

func writeRefreshTokenCookie(c echo.Context, refreshToken string) {

	cookie := new(http.Cookie)

	cookie.Name = "refresh_token"
	cookie.Value = refreshToken
	cookie.HttpOnly = true
	cookie.SameSite = 3
	cookie.Path = "/"

	c.SetCookie(cookie)
}
