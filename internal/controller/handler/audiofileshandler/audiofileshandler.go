package audiofileshandler

import (
	"io"
	"net/http"
	"os"

	"github.com/RecoBattle/internal/app/asr"
	"github.com/RecoBattle/internal/app/audiofilesapp"
	"github.com/RecoBattle/internal/controller/handler"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
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

func (lh *AudioFilesHandler) RegisterHandler(e *echo.Echo, publicGroup, privateGroup *echo.Group) {

	privateGroup.POST("/asr/audiofile", lh.SetAudioFile)
	privateGroup.GET("/asr/audiofiles", lh.GetAudioFiles)
	privateGroup.GET("/asr/textfile/:uuid", lh.GetResultASR)
}

// SetAudioFile
//
//	@Summary      SetAudioFile
//	@Description  add audio file
//	@Param        json body RequestData
//	@Success      200 {string} wav file has already been uploaded by this user
//	@Success      202 {string} the new wav file has been accepted for processing
//	@Failure      400 {string} invalid request format
//	@Failure      401 {string} the user is not authenticated
//	@Failure      422 {string} invalid ASR format or audio file type
//	@Failure      500 {string} internal server error
//	@Router       /api_private/asr/audiofile [post]
//
//	@Security JWT Token
func (lh *AudioFilesHandler) SetAudioFile(c echo.Context) error {

	ca := make(chan bool)
	errc := make(chan error)

	audioFile := new(RequestData)
	err := c.Bind(audioFile)
	if err != nil {
		logrus.Errorf("error in bind audio file request. error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(audioFile); err != nil {
		logrus.Errorf("error in bind audio file  request. error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userID := handler.GetUserID(c)

	asr, ok := lh.ASRRegistry.GetService(audioFile.ASR)
	if !ok {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "")
	}

	newAudioFile := audiofilesapp.AudioFile{
		FileName: audioFile.FileName,
		ASR:      audioFile.ASR,
		UserID:   userID,
	}

	go func() error {

		filePath := lh.PathFileStorage + audioFile.FileName
		file, err := os.Create(filePath)
		if err != nil {
			errc <- err
			return err
		}

		defer file.Close()

		_, err = io.WriteString(file, audioFile.Audio)
		if err != nil {
			errc <- err
			return err
		}

		err = lh.AudioFilesApp.Create(c.Request().Context(), newAudioFile)

		if err != nil {
			errc <- err
			return err
		}

		data, err := os.ReadFile(filePath)

		if err != nil {
			errc <- err
			return err
		}

		err = lh.AudioFilesApp.AddASRProcessing(c.Request().Context(), newAudioFile, asr, data)

		if err != nil {
			errc <- err
			return err
		}

		ca <- true
		return nil

	}()

	select {
	case <-ca:
		return c.String(http.StatusOK, "OK")
	case err := <-errc:
		logrus.Errorf("error: %v", err)
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

	userID := handler.GetUserID(c)

	go func() error {
		outputData, err := lh.AudioFilesApp.GetAudioFiles(c.Request().Context(), userID)

		if err != nil {
			errc <- err
			return err
		}

		ca <- *outputData
		return nil
	}()

	select {
	case result := <-ca:
		if len(result) == 0 {
			return echo.NewHTTPError(http.StatusNoContent)
		}
		return c.JSON(http.StatusOK, result)
	case err := <-errc:
		logrus.Errorf("error: %v", err)
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

	uuid := c.Param("id")

	go func() error {
		outputData, err := lh.AudioFilesApp.GetResultASR(c.Request().Context(), uuid)

		if err != nil {
			errc <- err
			return err
		}

		ca <- *outputData
		return nil
	}()

	select {
	case result := <-ca:
		if len(result) == 0 {
			return echo.NewHTTPError(http.StatusNoContent)
		}
		return c.JSON(http.StatusOK, result)
	case err := <-errc:
		logrus.Errorf("error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}
}