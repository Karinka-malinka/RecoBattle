package userapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/RecoBattle/cmd/config"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type User struct {
	UUID     uuid.UUID
	Username string
	Password string
}

type UserStore interface {
	Create(ctx context.Context, user User) error
	GetByName(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
}

type Users struct {
	userStore UserStore
	cfg       config.ApiServer
}

func NewUser(userStore UserStore) *Users {
	return &Users{
		userStore: userStore,
	}
}

func (ua *Users) Register(ctx context.Context, cfg config.ApiServer, user User) (*LoginResponse, error) {

	user.UUID = uuid.New()
	user.Password = hex.EncodeToString(ua.writeHash(user.Username, user.Password))

	if err := ua.userStore.Create(ctx, user); err != nil {
		return nil, err
	}

	accessToken, err := ua.newToken(user, cfg.AccessTokenExpiresAt, cfg.SecretKeyForAccessToken)
	if err != nil {
		return nil, err
	}

	refreshToken, err := ua.newToken(user, cfg.RefreshTokenExpiresAt, cfg.SecretKeyForRefreshToken)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (ua *Users) Login(ctx context.Context, cfg config.ApiServer, user User) (*LoginResponse, error) {

	userInDB, err := ua.userStore.GetByName(ctx, user.Username)

	if err != nil {
		return nil, err
	}

	if !ua.checkHash(user, userInDB.Password) {
		return nil, errors.New("401")
	}

	accessToken, err := ua.newToken(user, cfg.AccessTokenExpiresAt, cfg.SecretKeyForAccessToken)
	if err != nil {
		return nil, err
	}

	refreshToken, err := ua.newToken(user, cfg.RefreshTokenExpiresAt, cfg.SecretKeyForRefreshToken)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (ua *Users) checkHash(user User, userHash string) bool {
	check1 := ua.writeHash(user.Username, user.Password)
	check2, err := hex.DecodeString(userHash)

	if err != nil {
		logrus.Printf("Error in decode user hash. error: %v\n", err)
	}

	return hmac.Equal(check2, check1)
}

func (ua *Users) writeHash(username string, password string) []byte {
	hash := hmac.New(sha256.New, []byte(ua.cfg.SecretKeyForHashingPassword))
	hash.Write([]byte(fmt.Sprintf("%s:%s:%s", username, password, ua.cfg.SecretKeyForHashingPassword)))

	return hash.Sum(nil)
}
