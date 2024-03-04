package userhandler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RecoBattle/internal/app/userapp"
	"github.com/RecoBattle/internal/database/mocks"
	"github.com/golang/mock/gomock"
	echo "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var (
	userJSON = `{"login":"JonSnow","password":"1234"}`
)

func TestUserHandler_Register(t *testing.T) {

	//cfg := config.ApiServer{}
	//ctx := context.Background()

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockUserStore(ctrl)

	// Setup
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/api_public/user/register", strings.NewReader(userJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	/*userStore, err := userdb.NewUserStore(ctx, m)
	if err != nil {
		logrus.Fatalf("error in creating user store table")
	}*/
	userApp := userapp.NewUser(m)

	//cfg.BaseShortAddr = ""
	//h := &UserHandler{}
	h := NewUserHandler(userApp)

	// Assertions
	if assert.NoError(t, h.Register(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}
