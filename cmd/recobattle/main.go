package main

import (
	"context"
	"log"
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
)

func main() {

	logger.SetUpLogger()

	cfg := config.NewConfig()

	parseFlags(cfg)

	cnf, err := cfg.GetConfig(cfg.ConfigASR)
	if err != nil {
		log.Fatalf("cnf is not set. Error: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	defer cancel()

	db, err := database.NewDB(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("error in open database. error: %v", err)
	}

	//Add ASR
	asrRegistry := asr.ASRRegistry{Services: make(map[string]asr.ASR)}

	yandexASR := yandexspeachkit.NewYandexASRStore(cnf.YandexAsr)
	asrRegistry.AddService("yandexSpeachKit", yandexASR)

	//Init storage and services
	userStore := userdb.NewUserStore(db.DB)
	userApp := userapp.NewUser(userStore, cnf.ApiServer)

	audiofileStore := audiofilesdb.NewAudioFileStore(db.DB)
	audiofilesApp := audiofilesapp.NewAudioFile(audiofileStore)

	qcStore := qualitycontroldb.NewQCStore(db.DB)
	qcApp := qualitycontrolapp.NewQualityControl(qcStore)

	//Add Actions to Handlers to slice
	var registeredHandlers []handler.Handler

	userHandler := userhandler.NewUserHandler(userApp)
	registeredHandlers = append(registeredHandlers, userHandler)

	audiofilesHandler := audiofileshandler.NewAudioFilesHandler(audiofilesApp, &asrRegistry, cfg.PathFileStorage)
	registeredHandlers = append(registeredHandlers, audiofilesHandler)

	qcHandler := qualitycontrolhandler.NewQCHandler(qcApp)
	registeredHandlers = append(registeredHandlers, qcHandler)

	appRouter := router.NewRouter(cnf.ApiServer, registeredHandlers, userApp)
	appServer := server.NewServer(cfg.RunAddr, appRouter.Echo)

	go appServer.Start(ctx)

	<-ctx.Done()
	appServer.Stop(ctx)
}
