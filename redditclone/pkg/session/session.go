package session

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"redditclone/pkg/user"
)

type Session struct {
	ID        string
	UserID    string
	UserLogin string
}

type sessKey string

var (
	ErrNoAuth                  = errors.New("no session found")
	SessionKey         sessKey = "sessionKey"
	ExampleTokenSecret         = []byte("супер секретный ключ")
)

func NewSession(currUser user.User) *Session {
	randID := make([]byte, 16)
	_, err := rand.Read(randID)
	if err != nil {
		fmt.Println("error:", err)
	}
	return &Session{
		ID:        fmt.Sprintf("%x", randID),
		UserID:    currUser.ID,
		UserLogin: currUser.Login,
	}
}

func SessionFromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value(SessionKey).(*Session)
	if !ok || sess == nil {
		return nil, ErrNoAuth
	}
	return sess, nil
}

func ContextWithSession(ctx context.Context, sess *Session) context.Context {
	return context.WithValue(ctx, SessionKey, sess)
}

type SessRepo interface {
	Create(user user.User) (*Session, error)
	Check(w http.ResponseWriter, r *http.Request) (*Session, error)
	CreateToken(sess *Session) (string, error)
}
