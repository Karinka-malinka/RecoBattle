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

var registeredHandlers []handler.Handler

const ConfigASR = "../../../../cmd/config/config.toml"

func TestUserHandler_Register(t *testing.T) {

	cfg := config.NewConfig()

	cnf, err := cfg.GetConfig(ConfigASR)
	if err != nil {
		log.Fatalf("cnf is not set. Error: %v", err)
	}

	reqBody := `{"login": "testuser", "password": "testpassword"}`

	userID := uuid.MustParse("2d53b244-8844-40a6-ab37-e5b89019af0a")

	user := &userapp.User{
		UUID:     userID,
		Username: "testuser",
		Password: "cff17119871bdcd21a5638b1134ec1bcc9be47e0ce3bcd9863a2a24a68c862b5",
	}

	t.Run("Successful Registration", func(t *testing.T) {

		mockUserStore := new(mocks.MockUserStore)

		userApp := userapp.NewUser(mockUserStore, cnf.ApiServer)
		userHandler := NewUserHandler(userApp)

		registeredHandlers = append(registeredHandlers, userHandler)

		e := router.NewRouter(cnf.ApiServer, registeredHandlers, userApp).Echo

		req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUserStore.On("Create", mock.Anything, *user).Return(nil)

		if assert.NoError(t, userHandler.Register(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

	})

	t.Run("Conflict", func(t *testing.T) {

		mockUserStore := new(mocks.MockUserStore)

		userApp := userapp.NewUser(mockUserStore, cnf.ApiServer)
		userHandler := NewUserHandler(userApp)

		registeredHandlers = append(registeredHandlers, userHandler)

		e := router.NewRouter(cnf.ApiServer, registeredHandlers, userApp).Echo

		req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUserStore.On("Create", mock.Anything, *user).Return(database.NewErrorConflict(errors.New("409")))

		if assert.NoError(t, userHandler.Register(c)) {
			assert.Equal(t, http.StatusConflict, rec.Code)
		}

	})

	t.Run("Bad request", func(t *testing.T) {

		reqBody := `{"login": "testuser", "pass": "testpassword"}`

		mockUserStore := new(mocks.MockUserStore)

		userApp := userapp.NewUser(mockUserStore, cnf.ApiServer)
		userHandler := NewUserHandler(userApp)

		registeredHandlers = append(registeredHandlers, userHandler)

		e := router.NewRouter(cnf.ApiServer, registeredHandlers, userApp).Echo

		req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUserStore.On("Create", mock.Anything, *user).Return(nil)

		err := userHandler.Register(c)
		assert.Error(t, err)
		httpError := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	})
}

func TestUserHandler_Login(t *testing.T) {

	cfg := config.NewConfig()

	cnf, err := cfg.GetConfig(ConfigASR)
	if err != nil {
		log.Fatalf("cnf is not set. Error: %v", err)
	}

	reqBody := `{"login": "testuser", "password": "testpassword"}`

	userID := uuid.MustParse("2d53b244-8844-40a6-ab37-e5b89019af0a")
	user := &userapp.User{
		UUID:     userID,
		Username: "testuser",
		Password: "cff17119871bdcd21a5638b1134ec1bcc9be47e0ce3bcd9863a2a24a68c862b5",
	}

	t.Run("Invalid username/password pair", func(t *testing.T) {

		mockUserStore := new(mocks.MockUserStore)

		userApp := userapp.NewUser(mockUserStore, cnf.ApiServer)
		userHandler := NewUserHandler(userApp)

		registeredHandlers = append(registeredHandlers, userHandler)

		e := router.NewRouter(cnf.ApiServer, registeredHandlers, userApp).Echo

		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUserStore.On("GetUser", mock.Anything, mock.Anything).Return(&userapp.User{}, errors.New("401"))

		if assert.NoError(t, userHandler.Login(c)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}

	})

	t.Run("Successful Login", func(t *testing.T) {

		mockUserStore := new(mocks.MockUserStore)

		userApp := userapp.NewUser(mockUserStore, cnf.ApiServer)
		userHandler := NewUserHandler(userApp)

		registeredHandlers = append(registeredHandlers, userHandler)

		e := router.NewRouter(cnf.ApiServer, registeredHandlers, userApp).Echo

		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUserStore.On("GetUser", mock.Anything, map[string]string{"login": "testuser"}).Return(user, nil)

		if assert.NoError(t, userHandler.Login(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

	})

}
