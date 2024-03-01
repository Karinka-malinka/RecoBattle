package router

import (
	"net/http"

	"github.com/RecoBattle/cmd/config"
	"github.com/RecoBattle/internal/app/userapp"
	"github.com/RecoBattle/internal/controller/handler"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Router struct {
	Echo *echo.Echo
}

func NewRouter(cfg config.ApiServer, handlers []handler.Handler) *Router {

	e := echo.New()

	r := &Router{
		Echo: e,
	}

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:          true,
		LogStatus:       true,
		LogMethod:       true,
		LogResponseSize: true,
		LogLatency:      true,
		LogError:        true,
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.Decompress())
	e.Use(middleware.Gzip())

	restrictedConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &userapp.JWTCustomClaims{}
		},
		SigningKey: []byte(cfg.SecretKeyForAccessToken),
		ErrorHandler: func(c echo.Context, err error) error {
			return r.TokenRefresher(c, cfg)
		},
		ContinueOnIgnoredError: true,
		TokenLookup:            "header:Authorization:Bearer,cookie:access_token",
	}

	e.Use(echojwt.WithConfig(restrictedConfig))

	apiUserGroup := e.Group("/api/user")

	for _, handler := range handlers {
		handler.RegisterHandler(e, apiUserGroup)
	}

	return r
}

func (rt *Router) TokenRefresher(c echo.Context, cfg config.ApiServer) error {

	/*
		cookie, err := c.Cookie("refresh_token")

		var user *userapp.User


			if err != nil {
				user, err = rt.UserAPP.Register(c.Request().Context(), cfg)
			} else {
				user, err = rt.UserAPP.UpdateToken(c.Request().Context(), cookie.Value, cfg)
				if err != nil {
					user, err = rt.UserAPP.Register(c.Request().Context(), cfg)
				}
			}

			if err != nil {
				if errors.Is(err, jwt.ErrSignatureInvalid) {
					c.Response().Writer.WriteHeader(http.StatusUnauthorized)
				} else {
					c.Response().Writer.WriteHeader(http.StatusInternalServerError)
				}
			}

			sendResponceToken(c, user.Token)
	*/
	return nil
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
