package session

import (
	"fmt"
	"net/http"
	"redditclone/pkg/user"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type SessionsManager struct{}

func NewSessionsManager() *SessionsManager {
	return &SessionsManager{}
}

func (sm *SessionsManager) Check(w http.ResponseWriter, r *http.Request) (*Session, error) {

	hashSecretGetter := func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, fmt.Errorf("bad sign method")
		}
		return ExampleTokenSecret, nil
	}

	inToken := r.Header.Get("authorization")
	if inToken == "" {
		return nil, ErrNoAuth
	}
	inToken = strings.Split(inToken, " ")[1]
	token, errJwt := jwt.Parse(inToken, hashSecretGetter)
	if errJwt != nil {
		return nil, errJwt
	}
	payload, ok := token.Claims.(jwt.MapClaims)

	if ok {
		sessClaims := payload["user"].(map[string]interface{})

		sess := &Session{}
		sess.UserID = sessClaims["id"].(string)
		sess.UserLogin = sessClaims["username"].(string)

		return sess, nil
	}
	return nil, ErrNoAuth
}

func (sm *SessionsManager) Create(curUser user.User) (*Session, error) {
	sess := NewSession(curUser)
	return sess, nil
}

func (sm *SessionsManager) CreateToken(sess *Session) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": map[string]interface{}{
			"username": sess.UserLogin,
			"id":       sess.UserID,
		},
		"iat": time.Now().Unix(),
		"exp": time.Now().Unix() + 1200,
	})
	tokenString, err := token.SignedString(ExampleTokenSecret)

	return tokenString, err
}
