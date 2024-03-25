package qualitycontrolhandler

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RecoBattle/cmd/config"
	"github.com/RecoBattle/internal/app/qualitycontrolapp"
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
const PathTestFile = "../../../../testfile/test.wav"
const userID = "2d53b244-8844-40a6-ab37-e5b89019af0a"
const fileID = "1d35b422-7755-50a7-ab73-e4b98091af1a"

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

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("user", userID)

	c.SetPath("/api_private/qualitycontrol/:id_file")
	c.SetParamNames("id_file")
	c.SetParamValues(fileID)

	return c, qcHandler
}

func TestQCHandler_SetIdealText(t *testing.T) {

	ID := uuid.MustParse("2d53b244-8844-40a6-ab37-e5b89019af0a")

	qualityControl := qualitycontrolapp.IdealText{
		UUID:       ID,
		FileID:     fileID,
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

		reqBody = `{"id_file": "` + fileID + `", "ChannelTag": "1", "Text":"Hi"}`
		c, qcHandler = getEchoContext(mockQCStore, reqBody)

		if assert.NoError(t, qcHandler.SetIdealText(c)) {
			assert.Equal(t, http.StatusOK, c.Response().Status)
		}

	})

	t.Run("Conflict", func(t *testing.T) {

		mockQCStore := new(mocks.MockQualityControlStore)
		mockQCStore.On("Create", mock.Anything, qualityControl).Return(database.NewErrorConflict(errors.New("409")))
		reqBody = `{"id_file": "` + fileID + `", "ChannelTag": "1", "Text":"Hi"}`
		c, qcHandler = getEchoContext(mockQCStore, reqBody)

		if assert.NoError(t, qcHandler.SetIdealText(c)) {
			assert.Equal(t, http.StatusConflict, c.Response().Status)
		}
	})
}

func TestQCHandler_QualityControl(t *testing.T) {

	var data []qualitycontrolapp.QualityControl

	t.Run("No content", func(t *testing.T) {

		mockQCStore := new(mocks.MockQualityControlStore)
		mockQCStore.On("GetTextASRIdeal", mock.Anything, fileID).Return(data, "Hi", nil)

		c, qcHandler := getEchoContext(mockQCStore, "")

		err := qcHandler.QualityControl(c)
		assert.Error(t, err)
		httpError := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusNoContent, httpError.Code)
	})

	resASR := qualitycontrolapp.QualityControl{
		ASR:       "yandexSpeachKit",
		TestIdeal: "Hi",
		TextASR:   "Hi",
		Quality:   90,
	}
	data = append(data, resASR)

	t.Run("Successful", func(t *testing.T) {

		mockQCStore := new(mocks.MockQualityControlStore)
		mockQCStore.On("GetTextASRIdeal", mock.Anything, fileID).Return(data, "Hi", nil)

		c, qcHandler := getEchoContext(mockQCStore, "")

		if assert.NoError(t, qcHandler.QualityControl(c)) {
			assert.Equal(t, http.StatusOK, c.Response().Status)
		}

	})

}
