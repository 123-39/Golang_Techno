package user

import (
	"crypto/md5"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrNoUser  = errors.New("no user found")
	ErrBadPass = errors.New("invald password")
)

type UserMemoryRepository struct {
	data  map[string]User
	mutex sync.Mutex
}

func NewMemoryRepo() *UserMemoryRepository {
	return &UserMemoryRepository{
		data:  make(map[string]User),
		mutex: sync.Mutex{},
	}
}

func (repo *UserMemoryRepository) Authorize(login, pass string) (User, error) {
	user, ok := repo.data[login]
	if !ok {
		return User{}, ErrNoUser
	}

	pass = HashPass(pass)
	if user.password != pass {
		return user, ErrBadPass
	}

	return user, nil
}

func (repo *UserMemoryRepository) AddUser(login, pass string) (User, error) {
	pass = HashPass(pass)
	repo.mutex.Lock()
	repo.data[login] = User{
		ID:       uuid.New().String(),
		Login:    login,
		password: pass,
	}
	repo.mutex.Unlock()
	user, ok := repo.data[login]
	if !ok {
		return User{}, ErrNoUser
	}
	if user.password != pass {
		return User{}, ErrBadPass
	}
	return user, nil
}

func HashPass(data string) string {
	data += ""
	dataHash := fmt.Sprintf("%x", md5.Sum([]byte(data)))
	return dataHash
}
