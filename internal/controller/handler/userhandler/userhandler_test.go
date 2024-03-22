package userhandler

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RecoBattle/cmd/config"
	"github.com/RecoBattle/internal/app/userapp"
	"github.com/RecoBattle/internal/controller/handler"
	"github.com/RecoBattle/internal/controller/router"
	"github.com/RecoBattle/internal/database"
	"github.com/RecoBattle/internal/database/mocks"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const ConfigASR = "../../../../cmd/config/config.toml"
const userID = "2d53b244-8844-40a6-ab37-e5b89019af0a"

func getEchoContext(mockUserStore *mocks.MockUserStore, reqBody string) (echo.Context, *UserHandler) {

	var registeredHandlers []handler.Handler

	cfg := config.NewConfig()

	cnf, err := cfg.GetConfig(ConfigASR)
	if err != nil {
		log.Fatalf("cnf is not set. Error: %v", err)
	}

	userApp := userapp.NewUser(mockUserStore, cnf.ApiServer)
	userHandler := NewUserHandler(userApp)

	registeredHandlers = append(registeredHandlers, userHandler)

	e := router.NewRouter(cnf.ApiServer, registeredHandlers, userApp).Echo

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return c, userHandler
}

func TestUserHandler_Register(t *testing.T) {

	reqBody := `{"login": "testuser", "password": "testpassword"}`

	userID := uuid.MustParse(userID)

	user := &userapp.User{
		UUID:     userID,
		Username: "testuser",
		Password: "cff17119871bdcd21a5638b1134ec1bcc9be47e0ce3bcd9863a2a24a68c862b5",
	}

	mockUserStore := new(mocks.MockUserStore)
	mockUserStore.On("Create", mock.Anything, *user).Return(nil)

	t.Run("Bad request", func(t *testing.T) {

		reqBody := `{"login": "testuser", "pass": "testpassword"}`

		c, userHandler := getEchoContext(mockUserStore, reqBody)

		err := userHandler.Register(c)
		assert.Error(t, err)
		httpError := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	})

	t.Run("Successful Registration", func(t *testing.T) {

		c, userHandler := getEchoContext(mockUserStore, reqBody)

		if assert.NoError(t, userHandler.Register(c)) {
			assert.Equal(t, http.StatusOK, c.Response().Status)
		}

	})

	t.Run("Conflict", func(t *testing.T) {

		mockUserStore := new(mocks.MockUserStore)
		mockUserStore.On("Create", mock.Anything, *user).Return(database.NewErrorConflict(errors.New("409")))

		c, userHandler := getEchoContext(mockUserStore, reqBody)

		if assert.NoError(t, userHandler.Register(c)) {
			assert.Equal(t, http.StatusConflict, c.Response().Status)
		}

	})

}

func TestUserHandler_Login(t *testing.T) {

	reqBody := `{"login": "testuser", "password": "testpassword"}`

	userID := uuid.MustParse(userID)
	user := &userapp.User{
		UUID:     userID,
		Username: "testuser",
		Password: "cff17119871bdcd21a5638b1134ec1bcc9be47e0ce3bcd9863a2a24a68c862b5",
	}

	t.Run("Invalid username/password pair", func(t *testing.T) {

		mockUserStore := new(mocks.MockUserStore)
		mockUserStore.On("GetUser", mock.Anything, mock.Anything).Return(&userapp.User{}, errors.New("401"))

		c, userHandler := getEchoContext(mockUserStore, reqBody)

		if assert.NoError(t, userHandler.Login(c)) {
			assert.Equal(t, http.StatusUnauthorized, c.Response().Status)
		}

	})

	t.Run("Successful Login", func(t *testing.T) {

		mockUserStore := new(mocks.MockUserStore)
		mockUserStore.On("GetUser", mock.Anything, map[string]string{"login": "testuser"}).Return(user, nil)

		c, userHandler := getEchoContext(mockUserStore, reqBody)

		if assert.NoError(t, userHandler.Login(c)) {
			assert.Equal(t, http.StatusOK, c.Response().Status)
		}

	})

}
