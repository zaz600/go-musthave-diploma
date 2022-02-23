package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

const key = "secureSecretText" // TODO взять из волта

type UserClaims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	jwt.StandardClaims
}

func createToken(userID string, sessionID string) (string, error) {
	claims := UserClaims{
		UserID:    userID,
		SessionID: sessionID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    "Gophermart",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(key))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func getClaims(jwtToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		jwtToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(key), nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, errors.New("couldn't parse claims")
	}
	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return nil, errors.New("jwt is expired")
	}
	return claims, nil
}

func SetJWT(w http.ResponseWriter, session *entity.Session) error {
	jwtToken, err := createToken(session.UID, session.SessionID)
	if err != nil {
		return err
	}
	w.Header().Set("Authorization", "Bearer "+jwtToken)
	return nil
}

func GetClaims(r *http.Request) (*UserClaims, error) {
	jwtToken := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", 1)
	if jwtToken == "" {
		return nil, ErrTokenNotFound
	}
	claims, err := getClaims(jwtToken)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
