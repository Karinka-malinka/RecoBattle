package userhandler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RecoBattle/cmd/config"
	"github.com/RecoBattle/internal/app/userapp"
	"github.com/RecoBattle/internal/database/mocks"
	"github.com/go-playground/validator"
	"github.com/golang/mock/gomock"
	echo "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var (
	userJSON = `{"login":"JonSnow","password":"1234"}`
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func TestUserHandler_Register(t *testing.T) {

	cfg := config.ApiServer{}
	ctx := context.Background()

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockUserStore(ctrl)
	m.EXPECT().Create(ctx, userapp.User{Username: "JonSnow", Password: "1234"}).Return(nil)

	// Setup
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	req := httptest.NewRequest(http.MethodPost, "/api_public/user/register", strings.NewReader(userJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	/*userStore, err := userdb.NewUserStore(ctx, m)
	if err != nil {
		logrus.Fatalf("error in creating user store table")
	}*/
	userApp := userapp.NewUser(m, cfg)

	//cfg.BaseShortAddr = ""
	//h := &UserHandler{}
	h := NewUserHandler(userApp)

	// Assertions
	if assert.NoError(t, h.Register(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}
