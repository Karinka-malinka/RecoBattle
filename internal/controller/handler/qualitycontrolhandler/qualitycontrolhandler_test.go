package qualitycontrolhandler

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/RecoBattle/cmd/config"
	"github.com/RecoBattle/internal/app/qualitycontrolapp"
	"github.com/RecoBattle/internal/app/userapp"
	"github.com/RecoBattle/internal/controller/handler"
	"github.com/RecoBattle/internal/controller/router"
	"github.com/RecoBattle/internal/database"
	"github.com/RecoBattle/internal/database/mocks"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const ConfigASR = "../../../../cmd/config/config.toml"
const PathTestFile = "../../../../testfile/test.wav"
const userID = "2d53b244-8844-40a6-ab37-e5b89019af0a"

func getEchoContext(mockQCStore *mocks.MockQualityControlStore, reqBody string) (echo.Context, *QCHandler) {

	var registeredHandlers []handler.Handler

	cfg := config.NewConfig()

	cnf, err := cfg.GetConfig(ConfigASR)
	if err != nil {
		log.Fatalf("cnf is not set. Error: %v", err)
	}

	mockUserStore := new(mocks.MockUserStore)

	userApp := userapp.NewUser(mockUserStore, cnf.ApiServer)

	qcApp := qualitycontrolapp.NewQualityControl(mockQCStore)
	qcHandler := NewQCHandler(qcApp)
	registeredHandlers = append(registeredHandlers, qcHandler)

	e := router.NewRouter(cnf.ApiServer, registeredHandlers, userApp).Echo

	req := httptest.NewRequest(http.MethodPost, "/api_private/asr/audiofile", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tokenClaims := &userapp.JWTCustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cnf.ApiServer.AccessTokenExpiresAt) * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	c.Set("user", token)

	return c, qcHandler
}

func TestQCHandler_SetIdealText(t *testing.T) {

	ID := uuid.MustParse("2d53b244-8844-40a6-ab37-e5b89019af0a")

	qualityControl := qualitycontrolapp.IdealText{
		UUID:       ID,
		FileID:     "2d53b244-8844-40a6-ab37-e5b89019af0a",
		ChannelTag: "1",
		Text:       "Hi",
	}

	mockQCStore := new(mocks.MockQualityControlStore)
	mockQCStore.On("Create", mock.Anything, qualityControl).Return(nil)
	reqBody := `{"id_file": "", "ChannelTag": "1", "Text":"Hi"}`
	c, qcHandler := getEchoContext(mockQCStore, reqBody)

	t.Run("Bad request Validate", func(t *testing.T) {

		err := qcHandler.SetIdealText(c)
		assert.Error(t, err)
		httpError := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	})

	t.Run("Successful", func(t *testing.T) {

		reqBody = `{"id_file": "2d53b244-8844-40a6-ab37-e5b89019af0a", "ChannelTag": "1", "Text":"Hi"}`
		c, qcHandler = getEchoContext(mockQCStore, reqBody)

		if assert.NoError(t, qcHandler.SetIdealText(c)) {
			assert.Equal(t, http.StatusOK, c.Response().Status)
		}

	})

	t.Run("Conflict", func(t *testing.T) {

		mockQCStore := new(mocks.MockQualityControlStore)
		mockQCStore.On("Create", mock.Anything, qualityControl).Return(database.NewErrorConflict(errors.New("409")))
		reqBody = `{"id_file": "2d53b244-8844-40a6-ab37-e5b89019af0a", "ChannelTag": "1", "Text":"Hi"}`
		c, qcHandler = getEchoContext(mockQCStore, reqBody)

		if assert.NoError(t, qcHandler.SetIdealText(c)) {
			assert.Equal(t, http.StatusConflict, c.Response().Status)
		}
	})

}
