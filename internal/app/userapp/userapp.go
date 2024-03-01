package userapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/RecoBattle/cmd/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type JWTCustomClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type User struct {
	UUID     uuid.UUID
	Username string
	Password string
}

type UserStore interface {
	Create(ctx context.Context, user User) error
	Read(ctx context.Context, username string) (*User, error)
}

type Users struct {
	userStore UserStore
	cfg       config.ApiServer
}

type LoginResponse struct {
	AccessToken  string
	RefreshToken string
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

	resp, err := ua.newTokenPair(cfg, user)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (ua *Users) Login(ctx context.Context, cfg config.ApiServer, user User) (*LoginResponse, error) {

	userInDB, err := ua.userStore.Read(ctx, user.Username)

	if err != nil {
		return nil, errors.New("401")
	}

	if !ua.checkHash(user, userInDB.Password) {
		return nil, errors.New("401")
	}

	resp, err := ua.newTokenPair(cfg, user)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (ua *Users) newTokenPair(cfg config.ApiServer, user User) (*LoginResponse, error) {

	accessToken, refreshToken := ua.getTokensWithClaims(cfg, user)

	accessTokenString, err := accessToken.SignedString([]byte(cfg.SecretKeyForAccessToken))
	if err != nil {
		logrus.Errorf("error in signedString access token. error: %v", err)
		return nil, err
	}
	refreshTokenString, err := refreshToken.SignedString([]byte(cfg.SecretKeyForRefreshToken))
	if err != nil {
		logrus.Errorf("error in signedString refresh token. error: %v", err)
		return nil, err
	}

	return &LoginResponse{AccessToken: accessTokenString, RefreshToken: refreshTokenString}, nil
}

func (ua *Users) getTokensWithClaims(cfg config.ApiServer, user User) (accessToken *jwt.Token, refreshToken *jwt.Token) {

	accessTokenClaims := &JWTCustomClaims{
		UserID: user.UUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(cfg.AccessTokenExpiresAt))),
		},
	}

	refreshTokenClaims := &JWTCustomClaims{
		UserID: user.UUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(cfg.RefreshTokenExpiresAt))),
		},
	}

	accessToken = jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	refreshToken = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	return accessToken, refreshToken
}

func (ua *Users) parseToken(tokenstr, secretKey string) (bool, *JWTCustomClaims, error) {

	token, err := jwt.ParseWithClaims(tokenstr, &JWTCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secretKey), nil

	})

	if err != nil {
		if !errors.Is(err, jwt.ErrTokenExpired) {
			logrus.Infof("error in parsing token. error: %v", err)
			return false, nil, err
		}
	}

	userClaims := token.Claims.(*JWTCustomClaims)

	return token.Valid, userClaims, nil
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
