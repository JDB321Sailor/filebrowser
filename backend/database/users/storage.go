package users

import (
	"sync"
	"time"

	"github.com/gtsteffaniak/filebrowser/backend/common/errors"
)

// StorageBackend is the interface to implement for a users storage.
type StorageBackend interface {
	GetBy(interface{}) (*User, error)
	Gets() ([]*User, error)
	Save(u *User, changePass bool, disableScopeChange bool) error
	Update(u *User, adminActor bool, fields ...string) error
	DeleteByID(uint) error
	DeleteByUsername(string) error
}

// Store is an interface for user storage.
type Store interface {
	Get(id interface{}) (user *User, err error)
	Gets() ([]*User, error)
	Update(user *User, adminActor bool, fields ...string) error
	Save(user *User, changePass bool, disableScopeChange bool) error
	Delete(id interface{}) error
	LastUpdate(id uint) int64
	AddApiKey(username uint, name string, key AuthToken) error
	DeleteApiKey(username uint, name string) error
}

// Storage is a users storage.
type Storage struct {
	back    StorageBackend
	updated map[uint]int64
	mux     sync.RWMutex
}

// NewStorage creates a users storage from a backend.
func NewStorage(back StorageBackend) *Storage {
	return &Storage{
		back:    back,
		updated: map[uint]int64{},
	}
}

// Get allows you to get a user by its name or username. The provided
// id must be a string for username lookup or a uint for id lookup. If id
// is neither, a ErrInvalidDataType will be returned.
func (s *Storage) Get(id interface{}) (user *User, err error) {
	user, err = s.back.GetBy(id)
	if err != nil {
		return
	}
	return user, err
}

// Gets gets a list of all users.
func (s *Storage) Gets() ([]*User, error) {
	users, err := s.back.Gets()
	if err != nil {
		return nil, err
	}
	return users, err
}

// Update updates a user in the database.
func (s *Storage) Update(user *User, adminIActor bool, fields ...string) error {
	err := s.back.Update(user, adminIActor, fields...)
	if err != nil {
		return err
	}

	s.mux.Lock()
	s.updated[user.ID] = time.Now().Unix()
	s.mux.Unlock()
	return nil
}

func (s *Storage) AddApiKey(userID uint, name string, key AuthToken) error {
	user, err := s.Get(userID)
	if err != nil {
		return err
	}
	// Initialize the ApiKeys map if it is nil
	if user.ApiKeys == nil {
		user.ApiKeys = make(map[string]AuthToken)
	}
	user.ApiKeys[name] = key
	err = s.Update(user, true, "ApiKeys")
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) DeleteApiKey(userID uint, name string) error {
	user, err := s.Get(userID)
	if err != nil {
		return err
	}
	// Initialize the ApiKeys map if it is nil
	if user.ApiKeys == nil {
		user.ApiKeys = make(map[string]AuthToken)
	}
	delete(user.ApiKeys, name)
	err = s.Update(user, true, "ApiKeys")
	if err != nil {
		return err
	}

	return nil
}

// Save saves the user in a storage.
func (s *Storage) Save(user *User, changePass, disableScopeChange bool) error {
	return s.back.Save(user, changePass, disableScopeChange)
}

// Delete allows you to delete a user by its name or username. The provided
// id must be a string for username lookup or a uint for id lookup. If id
// is neither, a ErrInvalidDataType will be returned.
func (s *Storage) Delete(id interface{}) error {
	switch id := id.(type) {
	case string:
		return s.back.DeleteByUsername(id)
	case uint:
		return s.back.DeleteByID(id)
	default:
		return errors.ErrInvalidDataType
	}
}

// LastUpdate gets the timestamp for the last update of an user.
func (s *Storage) LastUpdate(id uint) int64 {
	s.mux.RLock()
	defer s.mux.RUnlock()
	if val, ok := s.updated[id]; ok {
		return val
	}
	return 0
}
