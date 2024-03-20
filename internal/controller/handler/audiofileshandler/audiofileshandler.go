package audiofileshandler

import (
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/RecoBattle/internal/app/asr"
	"github.com/RecoBattle/internal/app/audiofilesapp"
	"github.com/RecoBattle/internal/controller/handler"
	"github.com/RecoBattle/internal/database"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type AudioFilesHandler struct {
	AudioFilesApp   *audiofilesapp.AudioFiles
	ASRRegistry     *asr.ASRRegistry
	PathFileStorage string
}

type RequestData struct {
	ASR      string `json:"asr" validate:"required"`
	FileName string `json:"file_name" validate:"required"`
	Audio    string `json:"audio" validate:"required"`
}

func NewAudioFilesHandler(audioFilesApp *audiofilesapp.AudioFiles, asrRegistry *asr.ASRRegistry, pathFileStorage string) *AudioFilesHandler {
	return &AudioFilesHandler{AudioFilesApp: audioFilesApp, ASRRegistry: asrRegistry, PathFileStorage: pathFileStorage}
}

func (lh *AudioFilesHandler) RegisterHandler(_ *echo.Echo, _, privateGroup *echo.Group) {

	privateGroup.POST("/asr/audiofile", lh.SetAudioFile)
	privateGroup.GET("/asr/audiofiles", lh.GetAudioFiles)
	privateGroup.GET("/asr/textfile/:uuid", lh.GetResultASR)
}

// SetAudioFile
//
//	@Summary      SetAudioFile
//	@Description  add audio file
//	@Param        json body RequestData
//	@Success      202 {string} the new wav file has been accepted for processing
//	@Failure      400 {string} invalid request format
//	@Failure      401 {string} the user is not authenticated
//	@Failure      409 {string} wav file has already been uploaded by this user
//	@Failure      422 {string} invalid ASR format or audio file type
//	@Failure      500 {string} internal server error
//	@Router       /api_private/asr/audiofile [post]
//
//	@Security JWT Token
func (lh *AudioFilesHandler) SetAudioFile(c echo.Context) error {

	userID, err := handler.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	ca := make(chan bool)
	errc := make(chan error)

	audioFile := new(RequestData)
	err = c.Bind(audioFile)
	if err != nil {
		log.Errorf("error in bind audio file request. error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(audioFile); err != nil {
		log.Errorf("error in validate audio file  request. error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	asr, ok := lh.ASRRegistry.GetService(audioFile.ASR)
	if !ok {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "")
	}

	newAudioFile := audiofilesapp.AudioFile{
		FileName: audioFile.FileName,
		ASR:      audioFile.ASR,
		UserID:   userID,
	}

	go func() {

		filePath := lh.PathFileStorage + audioFile.FileName
		file, err := os.Create(filePath)
		if err != nil {
			errc <- err
			return
		}

		defer file.Close()

		_, err = io.WriteString(file, audioFile.Audio)
		if err != nil {
			errc <- err
			return
		}

		data, err := os.ReadFile(filePath)

		if err != nil {
			errc <- err
			return
		}

		newAudioFile.FileID, err = lh.AudioFilesApp.Create(c.Request().Context(), newAudioFile)

		if err != nil {
			errc <- err
			return
		}

		go lh.AudioFilesApp.AddASRProcessing(newAudioFile, asr, data)

		ca <- true
	}()

	select {
	case <-ca:
		return c.String(http.StatusAccepted, "OK")
	case err := <-errc:
		log.Errorf("error: %v", err)
		var errConflict *database.ConflictError
		if errors.As(err, &errConflict) {
			return c.String(http.StatusConflict, "wav file has already been uploaded by this user")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}
}

// GetAudioFiles
//
//	@Summary      GetAudioFiles
//	@Description  get all files from DB
//	@Success      200 {object} an array of uploaded wav files
//	@Failure      204 {string} no data for an answer
//	@Failure      401 {string} the user is not authenticated
//	@Failure      500 {string} internal server error
//	@Router       /api_private/asr/audiofiles [get]
//
//	@Security JWT Token
func (lh *AudioFilesHandler) GetAudioFiles(c echo.Context) error {

	ca := make(chan []audiofilesapp.AudioFile, 1)
	errc := make(chan error)

	userID, err := handler.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	go func() {
		outputData, err := lh.AudioFilesApp.GetAudioFiles(c.Request().Context(), userID)

		if err != nil {
			errc <- err
			return
		}

		ca <- *outputData
	}()

	select {
	case result := <-ca:
		if len(result) == 0 {
			return echo.NewHTTPError(http.StatusNoContent)
		}
		return c.JSON(http.StatusOK, result)
	case err := <-errc:
		log.Errorf("error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}
}

// GetResultASR
//
//	@Summary      GetResultASR
//	@Description  get result ASR
//	@Success      200 {object} array with recognized text
//	@Failure      204 {string} no data for an answer
//	@Failure      401 {string} the user is not authenticated
//	@Failure      500 {string} internal server error
//	@Router       /api_private/asr/textfile/:uuid [get]
//
//	@Security JWT Token
func (lh *AudioFilesHandler) GetResultASR(c echo.Context) error {

	ca := make(chan []audiofilesapp.ResultASR, 1)
	errc := make(chan error)

	uuid := c.Param("uuid")

	go func() {
		outputData, err := lh.AudioFilesApp.GetResultASR(c.Request().Context(), uuid)

		if err != nil {
			errc <- err
			return
		}

		ca <- *outputData
	}()

	select {
	case result := <-ca:
		if len(result) == 0 {
			return echo.NewHTTPError(http.StatusNoContent)
		}
		return c.JSON(http.StatusOK, result)
	case err := <-errc:
		log.Errorf("error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}
}
