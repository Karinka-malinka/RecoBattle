package userhandler

import (
	"errors"
	"net/http"

	"github.com/RecoBattle/cmd/config"
	"github.com/RecoBattle/internal/app/userapp"
	"github.com/RecoBattle/internal/controller/router"
	"github.com/RecoBattle/internal/database/userdb"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	UserApp *userapp.Users
	cfg     config.ApiServer
}

type RegisterRequest struct {
	Username string `json:"login" binding:"required" validate:"required"`
	Password string `json:"password" binding:"required" validate:"required"`
}

func NewUserHandler(userapp *userapp.Users, cfg *config.ConfigData) *UserHandler {
	return &UserHandler{UserApp: userapp}
}

func (lh *UserHandler) RegisterHandler(e *echo.Echo, apiGroup *echo.Group) {

	apiGroup.POST("/register", lh.Register)
	apiGroup.POST("/login", lh.Login)

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
//	@Router       /api/user/register [post]
//
//	@Security JWT Token
func (lh *UserHandler) Register(c echo.Context) error {

	ca := make(chan *userapp.LoginResponse)
	errc := make(chan error)

	user := new(RegisterRequest)
	err := c.Bind(user)
	if err != nil {
		logrus.Errorf("error in bind register User request. error: %v", err)
		return c.JSON(http.StatusBadRequest, "please check request struct")
	}

	go func() error {

		registerResult, err := lh.UserApp.Register(c.Request().Context(), lh.cfg, userapp.User{Username: user.Username, Password: user.Password})

		if err != nil {
			errc <- err
			return err
		}

		ca <- registerResult
		return nil
	}()

	select {
	case result := <-ca:
		router.SendResponceToken(c, result)
		return c.String(http.StatusOK, "")
	case err := <-errc:
		var errConflict *userdb.ErrConflict
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
//	@Router       /api/user/register [post]
//
//	@Security JWT Token
func (lh *UserHandler) Login(c echo.Context) error {

	ca := make(chan *userapp.LoginResponse)
	errc := make(chan error)

	user := new(RegisterRequest)
	err := c.Bind(user)
	if err != nil {
		logrus.Errorf("error in bind register User request. error: %v", err)
		return c.JSON(http.StatusBadRequest, "please check request struct")
	}

	go func() error {

		registerResult, err := lh.UserApp.Login(c.Request().Context(), lh.cfg, userapp.User{Username: user.Username, Password: user.Password})

		if err != nil {
			errc <- err
			return err
		}

		ca <- registerResult
		return nil
	}()

	select {
	case result := <-ca:
		router.SendResponceToken(c, result)
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
