package userhandler

import (
	"errors"
	"net/http"

	"github.com/RecoBattle/internal/app/userapp"
	"github.com/RecoBattle/internal/database"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	UserApp *userapp.Users
}

type RegisterRequest struct {
	Username string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func NewUserHandler(userapp *userapp.Users) *UserHandler {
	return &UserHandler{UserApp: userapp}
}

func (lh *UserHandler) RegisterHandler(e *echo.Echo, publicGroup, privateGroup *echo.Group) {

	publicGroup.POST("/user/register", lh.Register)
	publicGroup.POST("/user/login", lh.Login)

}

// Register для регистрации нового пользователя
//
//	@Summary      Register
//	@Description  register
//	@Tags         User
//	@Accept       json
//	@Produce      json
//	@Param        json body RegisterRequest true "Register new User"
//	@Success      200 {object} RegisterResponse
//	@Failure      400 {string} please check request struct
//	@Failure      409 {string} login is busy
//	@Failure      500 {string} internal server error
//	@Router       /api_public/user/register [post]
//
//	@Security JWT Token
func (lh *UserHandler) Register(c echo.Context) error {

	ca := make(chan *userapp.LoginResponse)
	errc := make(chan error)

	user := new(RegisterRequest)
	err := c.Bind(user)
	if err != nil {
		logrus.Errorf("error in bind register User request. error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(user); err != nil {
		logrus.Errorf("error in bind register User request. error: %v", err)
		return err
	}

	go func() error {

		registerResult, err := lh.UserApp.Register(c.Request().Context(), userapp.User{Username: user.Username, Password: user.Password})

		if err != nil {
			errc <- err
			return err
		}

		ca <- registerResult
		return nil
	}()

	select {
	case result := <-ca:
		SendResponceToken(c, result)
		return c.String(http.StatusOK, "")
	case err := <-errc:
		var errConflict *database.ErrConflict
		if errors.As(err, &errConflict) {
			return c.String(http.StatusConflict, "")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}
}

// Login для пользователя
//
//	@Summary      Login
//	@Description  login
//	@Tags         User
//	@Accept       json
//	@Produce      json
//	@Param        json body RegisterRequest true "Register new User"
//	@Success      200 {object} RegisterResponse
//	@Failure      400 {string} please check request struct
//	@Failure      401 {string} invalid username/password pair
//	@Failure      500 {string} internal server error
//	@Router       /api_public/user/register [post]
//
//	@Security JWT Token
func (lh *UserHandler) Login(c echo.Context) error {

	ca := make(chan *userapp.LoginResponse)
	errc := make(chan error)

	user := new(RegisterRequest)
	err := c.Bind(user)
	if err != nil {
		logrus.Errorf("error in bind register User request. error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(user); err != nil {
		logrus.Errorf("error in bind register User request. error: %v", err)
		return err
	}

	go func() error {

		registerResult, err := lh.UserApp.Login(c.Request().Context(), userapp.User{Username: user.Username, Password: user.Password})

		if err != nil {
			errc <- err
			return err
		}

		ca <- registerResult
		return nil
	}()

	select {
	case result := <-ca:
		SendResponceToken(c, result)
		return c.String(http.StatusOK, "")
	case err := <-errc:
		if err.Error() == "401" {
			return c.String(http.StatusUnauthorized, "invalid username/password pair")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}

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
