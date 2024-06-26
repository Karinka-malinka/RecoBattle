package userhandler

import (
	"errors"
	"net/http"

	"github.com/RecoBattle/internal/app/userapp"
	"github.com/RecoBattle/internal/controller/handler"
	"github.com/RecoBattle/internal/database"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
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

func (lh *UserHandler) RegisterHandler(_ *echo.Echo, publicGroup, _ *echo.Group) {

	publicGroup.POST("/user/register", lh.Register)
	publicGroup.POST("/user/login", lh.Login)

}

// Register для регистрации нового пользователя
//
//	@Summary      Register
//	@Description  register
//	@Param        json body RegisterRequest true "Register new User"
//	@Success      200 {object} RegisterResponse
//	@Failure      400 {string} please check request struct
//	@Failure      409 {string} login is busy
//	@Failure      500 {string} internal server error
//	@Router       /api_public/user/register [post]
func (lh *UserHandler) Register(c echo.Context) error {

	ca := make(chan *userapp.LoginResponse)
	errc := make(chan error)

	user := new(RegisterRequest)
	err := c.Bind(user)
	if err != nil {
		log.Errorf("error in bind register User request. error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(user); err != nil {
		log.Errorf("error in bind register User request. error: %v", err)
		return err
	}

	go func() {

		registerResult, err := lh.UserApp.Register(c.Request().Context(), userapp.User{Username: user.Username, Password: user.Password})

		if err != nil {
			errc <- err
			return
		}

		ca <- registerResult
	}()

	select {
	case result := <-ca:
		handler.SendResponceToken(c, result)
		return c.JSON(http.StatusOK, result)
	case err := <-errc:
		var errConflict *database.ConflictError
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
//	@Param        json body RegisterRequest true "Register new User"
//	@Success      200 {object} RegisterResponse
//	@Failure      400 {string} please check request struct
//	@Failure      401 {string} invalid username/password pair
//	@Failure      500 {string} internal server error
//	@Router       /api_public/user/register [post]
func (lh *UserHandler) Login(c echo.Context) error {

	ca := make(chan *userapp.LoginResponse)
	errc := make(chan error)

	user := new(RegisterRequest)
	err := c.Bind(user)
	if err != nil {
		log.Errorf("error in bind register User request. error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(user); err != nil {
		log.Errorf("error in bind register User request. error: %v", err)
		return err
	}

	go func() {

		registerResult, err := lh.UserApp.Login(c.Request().Context(), userapp.User{Username: user.Username, Password: user.Password})

		if err != nil {
			errc <- err
			return
		}

		ca <- registerResult
	}()

	select {
	case result := <-ca:
		handler.SendResponceToken(c, result)
		return c.JSON(http.StatusOK, result)
	case err := <-errc:
		if err.Error() == "401" {
			return c.String(http.StatusUnauthorized, "invalid username/password pair")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}

}
