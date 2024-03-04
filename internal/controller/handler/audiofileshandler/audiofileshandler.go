package audiofileshandler

import (
	"io"
	"net/http"
	"os"

	"github.com/RecoBattle/internal/app/audiofilesapp"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type AudioFilesHandler struct {
	AudioFiles      *audiofilesapp.AudioFiles
	PathFileStorage string
}

type RequestData struct {
	ASR      string `json:"asr" validate:"required"`
	FileName string `json:"file_name" validate:"required"`
	Audio    string `json:"audio" validate:"required"`
}

func NewAudioFilesHandler(audioFiles *audiofilesapp.AudioFiles, pathFileStorage string) *AudioFilesHandler {
	return &AudioFilesHandler{AudioFiles: audioFiles, PathFileStorage: pathFileStorage}
}

func (lh *AudioFilesHandler) RegisterHandler(e *echo.Echo, publicGroup, privateGroup *echo.Group) {

	privateGroup.POST("/user/audiofile", lh.SetAudioFile)
	privateGroup.GET("/user/audiofiles", lh.GetAudioFiles)

}

// SetAudioFile
//
//	@Summary      SetAudioFile
//	@Description  add audio file
//	@Tags         AudioFile
//	@Accept       json
//	@Produce      json
//	@Param        json body RequestData
//	@Success      200 {string} wav file has already been uploaded by this user
//	@Success      202 {string} the new wav file has been accepted for processing
//	@Failure      400 {string} invalid request format
//	@Failure      401 {string} the user is not authenticated
//	@Failure      422 {string} invalid ASR format or audio file type
//	@Failure      500 {string} internal server error
//	@Router       /api_private/user/audiofile [post]
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

	go func() error {

		out, err := os.Create(lh.PathFileStorage + audioFile.FileName)
		if err != nil {
			errc <- err
			return err
		}

		defer out.Close()

		_, err = io.WriteString(out, audioFile.Audio)
		if err != nil {
			errc <- err
			return err
		}

		// Обработка аудиофайла пропущена для краткости

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
//	@Tags         User
//	@Accept       json
//	@Produce      json
//	@Success      200 {object} wav file has already been uploaded by this user
//	@Failure      204 {string} no data for an answer
//	@Failure      401 {string} the user is not authenticated
//	@Failure      422 {string} invalid ASR format or audio file type
//	@Failure      500 {string} internal server error
//	@Router       /api_private/user/audiofiles [get]
//
//	@Security JWT Token
func (lh *AudioFilesHandler) GetAudioFiles(c echo.Context) error {
	return nil
}
