package qualitycontrolhandler

import (
	"errors"
	"net/http"

	"github.com/RecoBattle/internal/app/qualitycontrolapp"
	"github.com/RecoBattle/internal/database"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type QCHandler struct {
	QCApp *qualitycontrolapp.QualityControls
}

type RequestData struct {
	FileID     string `json:"id_file" validate:"required"`
	ChannelTag string `json:"channelTag" validate:"required"`
	Text       string `json:"text" validate:"required"`
}

func NewQCHandler(qcApp *qualitycontrolapp.QualityControls) *QCHandler {
	return &QCHandler{QCApp: qcApp}
}

func (lh *QCHandler) RegisterHandler(_ *echo.Echo, _, privateGroup *echo.Group) {

	privateGroup.POST("/qualitycontrol/ideal", lh.SetIdealText)
	privateGroup.GET("/qualitycontrol/:id_file", lh.QualityControl)
}

// SetIdealText
//
//	@Summary      SetIdealText
//	@Description  add audio file
//	@Param        json body RequestData
//	@Success      200 {string} wav file has already been uploaded by this user
//	@Failure      400 {string} invalid request format
//	@Failure      401 {string} the user is not authenticated
//	@Failure      409 {string} invalid ASR format or audio file type
//	@Failure      500 {string} internal server error
//	@Router       /api_private/qualitycontrol/ideal [post]
//
//	@Security JWT Token
func (lh *QCHandler) SetIdealText(c echo.Context) error {

	ca := make(chan bool)
	errc := make(chan error)

	idealText := new(RequestData)
	err := c.Bind(idealText)
	if err != nil {
		log.Errorf("error in bind ideal text request. error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(idealText); err != nil {
		log.Errorf("error in validate ideal text  request. error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	go func() {

		err = lh.QCApp.Create(c.Request().Context(), qualitycontrolapp.IdealText{FileID: idealText.FileID, ChannelTag: idealText.ChannelTag, Text: idealText.Text})

		if err != nil {
			errc <- err
			return
		}

		ca <- true
	}()

	select {
	case <-ca:
		return c.String(http.StatusOK, "OK")
	case err := <-errc:
		log.Errorf("error: %v", err)
		var errConflict *database.ConflictError
		if errors.As(err, &errConflict) {
			return c.String(http.StatusConflict, "")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}
}

// QualityControl
//
//	@Summary      QualityControl
//	@Description  add audio file
//	@Success      200 {object} wav file has already been uploaded by this user
//	@Failure      204 {string} no data
//	@Failure      401 {string} the user is not authenticated
//	@Failure      500 {string} internal server error
//	@Router       /api_private/qualitycontrol/:id_file [get]
//
//	@Security JWT Token
func (lh *QCHandler) QualityControl(c echo.Context) error {

	ca := make(chan []qualitycontrolapp.QualityControl)
	errc := make(chan error)

	fileID := c.Param("id_file")

	go func() {

		outputData, err := lh.QCApp.QualityControl(c.Request().Context(), fileID)

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
