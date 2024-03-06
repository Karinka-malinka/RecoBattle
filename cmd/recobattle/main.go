package main

import (
	"context"
	"os"
	"os/signal"

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
	"github.com/RecoBattle/internal/controller/server"
	"github.com/RecoBattle/internal/database"
	"github.com/RecoBattle/internal/database/audiofilesdb"
	"github.com/RecoBattle/internal/database/qualitycontroldb"
	"github.com/RecoBattle/internal/database/userdb"
	"github.com/RecoBattle/internal/logger"
	"github.com/sirupsen/logrus"
)

func main() {

	logger.SetUpLogger()

	cfg := config.NewConfig()

	parseFlags(cfg)

	cnf, err := cfg.GetConfig(cfg.ConfigASR)
	if err != nil {
		logrus.Fatalf("cnf is not set. Error: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	defer cancel()

	db, err := database.NewDB(ctx, cfg.DatabaseDSN)
	if err != nil {
		logrus.Fatalf("error in open database. error: %v", err)
	}

	//Add ASR
	asrRegistry := asr.ASRRegistry{Services: make(map[string]asr.ASR)}

	yandexASR := yandexspeachkit.ServiceASRYandex{Name: "yandexSpeachKit"}
	asrRegistry.AddService("yandexSpeachKit", yandexASR)

	//Init storage and services
	userStore, err := userdb.NewUserStore(ctx, db.DB)
	if err != nil {
		logrus.Fatalf("error in creating user store table")
	}
	userApp := userapp.NewUser(userStore)

	audiofileStore, err := audiofilesdb.NewAudioFileStore(ctx, db.DB)
	if err != nil {
		logrus.Fatalf("error in creating audiofile store table")
	}
	audiofilesApp := audiofilesapp.NewAudioFile(audiofileStore)

	qcStore, err := qualitycontroldb.NewQCStore(ctx, db.DB)
	if err != nil {
		logrus.Fatalf("error in creating user store table")
	}
	qcApp := qualitycontrolapp.NewQualityControl(qcStore)

	//Add Actions to Handlers to slice
	var registeredHandlers []handler.Handler

	userHandler := userhandler.NewUserHandler(userApp)
	registeredHandlers = append(registeredHandlers, userHandler)

	audiofilesHandler := audiofileshandler.NewAudioFilesHandler(audiofilesApp, &asrRegistry, cfg.PathFileStorage)
	registeredHandlers = append(registeredHandlers, audiofilesHandler)

	qcHandler := qualitycontrolhandler.NewQCHandler(qcApp)
	registeredHandlers = append(registeredHandlers, qcHandler)

	appRouter := router.NewRouter(cnf.ApiServer, registeredHandlers)
	appServer := server.NewServer(cfg.RunAddr, appRouter.Echo)

	go appServer.Start(ctx)

	<-ctx.Done()
	appServer.Stop(ctx)
}
