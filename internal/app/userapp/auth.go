package userapp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RecoBattle/cmd/config"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/gommon/log"
)

type JWTCustomClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type LoginResponse struct {
	AccessToken  string
	RefreshToken string
}

func (ua *Users) newToken(user User, tokenExpiresAt uint, SecretKeyForToken string) (string, error) {

	token := ua.getTokensWithClaims(user, tokenExpiresAt)

	tokenString, err := token.SignedString([]byte(SecretKeyForToken))
	if err != nil {
		log.Errorf("error in signedString access token. error: %v", err)
		return "", err
	}

	return tokenString, nil
}

func (ua *Users) getTokensWithClaims(user User, tokenExpiresAt uint) (token *jwt.Token) {

	tokenClaims := &JWTCustomClaims{
		UserID: user.UUID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(tokenExpiresAt) * time.Minute)),
		},
	}

	token = jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	return token
}

func (ua *Users) Token(ctx context.Context, refreshToken string, cfg config.ApiServer) (*LoginResponse, error) {

	valid, userClaims, err := ParseToken(refreshToken, cfg.SecretKeyForRefreshToken)
	if err != nil {
		return nil, err
	}

	if valid {
		user, err := ua.userStore.GetUser(ctx, map[string]string{"uuid": userClaims.UserID})
		if err != nil {
			return nil, err
		}

		accessToken, err := ua.newToken(*user, cfg.AccessTokenExpiresAt, cfg.SecretKeyForAccessToken)
		if err != nil {
			return nil, err
		}

		return &LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
	}

	return nil, errors.New("401")
}

func ParseToken(tokenstr, secretKey string) (bool, *JWTCustomClaims, error) {

	token, err := jwt.ParseWithClaims(tokenstr, &JWTCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secretKey), nil

	})

	if err != nil {
		if !errors.Is(err, jwt.ErrTokenExpired) {
			log.Infof("error in parsing token. error: %v", err)
			return false, nil, err
		}
	}

	userClaims := token.Claims.(*JWTCustomClaims)

	return token.Valid, userClaims, nil
}
