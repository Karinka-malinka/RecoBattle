package router

import (
	"net/http"

	"github.com/RecoBattle/cmd/config"
	"github.com/RecoBattle/internal/app/userapp"
	"github.com/RecoBattle/internal/controller/handler"
	"github.com/RecoBattle/internal/controller/handler/userhandler"
	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Router struct {
	Echo    *echo.Echo
	UserApp *userapp.Users
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func NewRouter(cfg config.ApiServer, handlers []handler.Handler, usersapp *userapp.Users) *Router {

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	r := &Router{
		Echo:    e,
		UserApp: usersapp,
	}

	//e.Use(middleware.Recover())
	e.Use(middleware.Decompress())
	e.Use(middleware.Gzip())

	publicGroup := e.Group("/api_public")
	privateGroup := e.Group("/api_private")

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

	privateGroup.Use(echojwt.WithConfig(restrictedConfig))

	for _, handler := range handlers {
		handler.RegisterHandler(e, publicGroup, privateGroup)
	}

	return r
}

func (rt *Router) TokenRefresher(c echo.Context, cfg config.ApiServer) error {

	cookie, err := c.Cookie("refresh_token")

	if err != nil {
		return c.String(http.StatusUnauthorized, "please check cookie")
	}

	tokenResponse, err := rt.UserApp.Token(c.Request().Context(), cookie.Value, cfg)

	if err != nil {
		if err.Error() == "401" {
			return c.String(http.StatusUnauthorized, "")
		}
		return c.String(http.StatusInternalServerError, "")
	}

	userhandler.SendResponceToken(c, tokenResponse)

	return nil
}
