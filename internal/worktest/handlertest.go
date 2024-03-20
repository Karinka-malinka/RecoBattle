package worktest

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/RecoBattle/cmd/config"
	"github.com/RecoBattle/internal/app/asr"
	yandexspeachkit "github.com/RecoBattle/internal/app/asr/yandexSpeachKit"
	"github.com/RecoBattle/internal/app/audiofilesapp"
	"github.com/RecoBattle/internal/app/qualitycontrolapp"
	"github.com/RecoBattle/internal/app/userapp"
	"github.com/RecoBattle/internal/controller/handler"
	"github.com/RecoBattle/internal/controller/handler/audiofileshandler"
	"github.com/RecoBattle/internal/controller/handler/qualitycontrolhandler"
	"github.com/RecoBattle/internal/controller/handler/userhandler"
	"github.com/RecoBattle/internal/controller/router"
	"github.com/RecoBattle/internal/database/mocks"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

const ConfigASR = "../../../../cmd/config/config.toml"
const PathTestFile = "../../../../testfile/test.wav"
const userID = "2d53b244-8844-40a6-ab37-e5b89019af0a"

type Handlers struct {
	UserHandler       *userhandler.UserHandler
	AudioFilesHandler *audiofileshandler.AudioFilesHandler
	QCHandler         *qualitycontrolhandler.QCHandler
}

func GetEchoContext(mockUserStore *mocks.MockUserStore, mockAudioFileStore *mocks.MockAudioFileStore, mockQCStore *mocks.MockQualityControlStore, reqBody string) (echo.Context, *Handlers) {

	var registeredHandlers []handler.Handler

	cfg := config.NewConfig()

	cnf, err := cfg.GetConfig(ConfigASR)
	if err != nil {
		log.Fatalf("cnf is not set. Error: %v", err)
	}

	userApp := userapp.NewUser(mockUserStore, cnf.ApiServer)
	userHandler := userhandler.NewUserHandler(userApp)
	registeredHandlers = append(registeredHandlers, userHandler)

	asrRegistry := asr.ASRRegistry{Services: make(map[string]asr.ASR)}

	yandexASR := yandexspeachkit.NewYandexASRStore(cnf.YandexAsr)
	asrRegistry.AddService("yandexSpeachKit", yandexASR)

	audiofilesApp := audiofilesapp.NewAudioFile(mockAudioFileStore)
	audiofilesHandler := audiofileshandler.NewAudioFilesHandler(audiofilesApp, &asrRegistry, cfg.PathFileStorage)
	registeredHandlers = append(registeredHandlers, audiofilesHandler)

	qcApp := qualitycontrolapp.NewQualityControl(mockQCStore)
	qcHandler := qualitycontrolhandler.NewQCHandler(qcApp)
	registeredHandlers = append(registeredHandlers, qcHandler)

	e := router.NewRouter(cnf.ApiServer, registeredHandlers, userApp).Echo

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
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

	return c, &Handlers{UserHandler: userHandler, AudioFilesHandler: audiofilesHandler, QCHandler: qcHandler}
}
