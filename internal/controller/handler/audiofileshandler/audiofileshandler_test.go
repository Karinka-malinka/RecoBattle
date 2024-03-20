package audiofileshandler

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/RecoBattle/cmd/config"
	"github.com/RecoBattle/internal/app/asr"
	yandexspeachkit "github.com/RecoBattle/internal/app/asr/yandexSpeachKit"
	"github.com/RecoBattle/internal/app/audiofilesapp"
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

var registeredHandlers []handler.Handler

const ConfigASR = "../../../../cmd/config/config.toml"
const PathTestFile = "../../../../testfile/test.wav"
const userID = "2d53b244-8844-40a6-ab37-e5b89019af0a"

/*
type Config struct {
	userApp     userapp.Users
	cnf         config.ApiServer
	asrRegistry asr.ASRRegistry
}
*/

func getEchoContext(mockAudioFileStore *mocks.MockAudioFileStore, reqBody string) (echo.Context, *AudioFilesHandler) {

	cfg := config.NewConfig()

	cnf, err := cfg.GetConfig(ConfigASR)
	if err != nil {
		log.Fatalf("cnf is not set. Error: %v", err)
	}

	mockUserStore := new(mocks.MockUserStore)

	userApp := userapp.NewUser(mockUserStore, cnf.ApiServer)

	asrRegistry := asr.ASRRegistry{Services: make(map[string]asr.ASR)}

	yandexASR := yandexspeachkit.NewYandexASRStore(cnf.YandexAsr)
	asrRegistry.AddService("yandexSpeachKit", yandexASR)

	audiofilesApp := audiofilesapp.NewAudioFile(mockAudioFileStore)
	audiofilesHandler := NewAudioFilesHandler(audiofilesApp, &asrRegistry, cfg.PathFileStorage)

	registeredHandlers = append(registeredHandlers, audiofilesHandler)

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

	return c, audiofilesHandler
}

func getAudiofile() audiofilesapp.AudioFile {

	ID := uuid.MustParse("2d53b244-8844-40a6-ab37-e5b89019af0a")

	audioFile := audiofilesapp.AudioFile{
		UUID:     ID,
		FileName: "testfile.wav",
		FileID:   "efc4ec14fd3fae7710335da2df3e14e5d0f031ed8e252005e501acb55e9f37d4",
		ASR:      "yandexSpeachKit",
		UserID:   userID,
	}

	return audioFile
}

func TestAudioFilesHandler_SetAudioFile(t *testing.T) {

	audioFile := getAudiofile()
	reqBody := `{"asr": "yandexSpeachKit", "file_name": "testfile.wav", "audio":""}`

	mockAudioFileStore := new(mocks.MockAudioFileStore)
	mockAudioFileStore.On("CreateFile", mock.Anything, audioFile).Return(nil)
	mockAudioFileStore.On("CreateASR", mock.Anything, audioFile).Return(nil)

	c, audiofilesHandler := getEchoContext(mockAudioFileStore, reqBody)

	t.Run("Bad request", func(t *testing.T) {

		err := audiofilesHandler.SetAudioFile(c)
		assert.Error(t, err)
		httpError := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)

	})

	data, err := os.ReadFile(PathTestFile)
	if err != nil {
		log.Fatalf("Error reading file: %s", err)
	}

	audioBase64 := base64.StdEncoding.EncodeToString(data)
	reqBody = fmt.Sprintf(`{"asr": "yandexSpeachKit", "file_name": "testfile.wav", "audio":"%s"}`, audioBase64)

	c, audiofilesHandler = getEchoContext(mockAudioFileStore, reqBody)

	t.Run("Successful", func(t *testing.T) {

		if assert.NoError(t, audiofilesHandler.SetAudioFile(c)) {
			assert.Equal(t, http.StatusAccepted, c.Response().Status)
		}

	})

	t.Run("Unauthorized", func(t *testing.T) {

		c.Set("user", nil)
		err := audiofilesHandler.SetAudioFile(c)
		assert.Error(t, err)
		httpError := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusUnauthorized, httpError.Code)
	})

	t.Run("Unprocessable entity", func(t *testing.T) {

		reqBodyIncorrect := fmt.Sprintf(`{"asr": "3iTech", "file_name": "testfile.wav", "audio":"%s"}`, audioBase64)

		c, audiofilesHandler := getEchoContext(mockAudioFileStore, reqBodyIncorrect)

		err := audiofilesHandler.SetAudioFile(c)
		assert.Error(t, err)
		httpError := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusUnprocessableEntity, httpError.Code)
	})

	t.Run("Conflict", func(t *testing.T) {

		mockAudioFileStore := new(mocks.MockAudioFileStore)
		mockAudioFileStore.On("CreateFile", mock.Anything, audioFile).Return(database.NewErrorConflict(errors.New("409")))

		c, audiofilesHandler := getEchoContext(mockAudioFileStore, reqBody)

		if assert.NoError(t, audiofilesHandler.SetAudioFile(c)) {
			assert.Equal(t, http.StatusConflict, c.Response().Status)
		}
	})
}

func TestAudioFilesHandler_GetAudioFiles(t *testing.T) {

	var files []audiofilesapp.AudioFile

	mockAudioFileStore := new(mocks.MockAudioFileStore)
	mockAudioFileStore.On("GetAudioFiles", mock.Anything, userID).Return(&files, nil)

	c, audiofilesHandler := getEchoContext(mockAudioFileStore, "")

	t.Run("No content", func(t *testing.T) {

		err := audiofilesHandler.GetAudioFiles(c)
		assert.Error(t, err)
		httpError := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusNoContent, httpError.Code)
	})

	audioFile := getAudiofile()
	files = append(files, audioFile)

	t.Run("Successful", func(t *testing.T) {

		mockAudioFileStore := new(mocks.MockAudioFileStore)
		mockAudioFileStore.On("GetAudioFiles", mock.Anything, userID).Return(&files, nil)

		c, audiofilesHandler := getEchoContext(mockAudioFileStore, "")

		if assert.NoError(t, audiofilesHandler.GetAudioFiles(c)) {
			assert.Equal(t, http.StatusOK, c.Response().Status)
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {

		c.Set("user", nil)

		err := audiofilesHandler.GetAudioFiles(c)
		assert.Error(t, err)
		httpError := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusUnauthorized, httpError.Code)

	})
}

func TestAudioFilesHandler_GetResultASR(t *testing.T) {

	var resASR []audiofilesapp.ResultASR

	mockAudioFileStore := new(mocks.MockAudioFileStore)
	mockAudioFileStore.On("GetResultASR", mock.Anything, userID).Return(&resASR, nil)

	c, audiofilesHandler := getEchoContext(mockAudioFileStore, "")

	t.Run("No content", func(t *testing.T) {

		err := audiofilesHandler.GetResultASR(c)
		assert.Error(t, err)
		httpError := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusNoContent, httpError.Code)
	})

	ID := uuid.MustParse("2d53b244-8844-40a6-ab37-e5b89019af0a")
	r := audiofilesapp.ResultASR{
		UUID:       ID,
		ChannelTag: "1",
		Text:       "res",
	}
	resASR = append(resASR, r)

	t.Run("Successful", func(t *testing.T) {

		mockAudioFileStore := new(mocks.MockAudioFileStore)
		mockAudioFileStore.On("GetResultASR", mock.Anything, userID).Return(&resASR, nil)

		c, audiofilesHandler := getEchoContext(mockAudioFileStore, "")

		if assert.NoError(t, audiofilesHandler.GetResultASR(c)) {
			assert.Equal(t, http.StatusOK, c.Response().Status)
		}
	})
}
