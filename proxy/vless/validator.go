package vless

import (
	"strings"
	"sync"

	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/uuid"
)

type Validator interface {
	Get(id uuid.UUID) *protocol.MemoryUser
	Load(idOrEmail string) (memoryUser *protocol.MemoryUser, exists bool)
	Add(u *protocol.MemoryUser) error
	Del(email string) error
	GetAllIDs() []string
	GetAll() map[string]*protocol.MemoryUser
}

// MemoryValidator stores valid VLESS users.
type MemoryValidator struct {
	// Considering email's usage here, map + sync.Mutex/RWMutex may have better performance.
	email sync.Map
	users sync.Map
}

// Add a VLESS user, Email must be empty or unique.
func (v *MemoryValidator) Add(u *protocol.MemoryUser) error {
	if u.Email != "" {
		_, loaded := v.email.LoadOrStore(strings.ToLower(u.Email), u)
		if loaded {
			return errors.New("User ", u.Email, " already exists.")
		}
	}
	v.users.Store(u.Account.(*MemoryAccount).ID.UUID(), u)
	return nil
}

// Del a VLESS user with a non-empty Email.
func (v *MemoryValidator) Del(e string) error {
	if e == "" {
		return errors.New("Email must not be empty.")
	}
	le := strings.ToLower(e)
	u, _ := v.email.Load(le)
	if u == nil {
		return errors.New("User ", e, " not found.")
	}
	v.email.Delete(le)
	v.users.Delete(u.(*protocol.MemoryUser).Account.(*MemoryAccount).ID.UUID())
	return nil
}

// Get a VLESS user with UUID, nil if user doesn't exist.
func (v *MemoryValidator) Get(id uuid.UUID) *protocol.MemoryUser {
	u, _ := v.users.Load(id)
	if u != nil {
		return u.(*protocol.MemoryUser)
	}
	return nil
}

func (v *MemoryValidator) Load(idOrEmail string) (memoryUser *protocol.MemoryUser, exists bool) {
	var user any
	if user, exists = v.email.Load(idOrEmail); exists {
		return user.(*protocol.MemoryUser), exists
	}
	if id, err := uuid.ParseString(idOrEmail); err == nil {
		if user, exists = v.users.Load(id); exists {
			return user.(*protocol.MemoryUser), exists
		}
	}
	return nil, false
}

func (v *MemoryValidator) GetAllIDs() []string {
	users := []string{}
	v.email.Range(func(key, value any) bool {
		id, ok := key.(string)
		if !ok {
			return true
		}
		users = append(users, id)
		return true
	})
	return users
}

func (v *MemoryValidator) GetAll() map[string]*protocol.MemoryUser {
	var user any
	var memoryUser *protocol.MemoryUser
	var id string
	var ok bool
	users := map[string]*protocol.MemoryUser{}

	v.email.Range(func(key, value any) bool {
		if id, ok = key.(string); !ok {
			return true
		}
		if user, ok = v.users.Load(id); !ok {
			return true
		}
		if memoryUser, ok = user.(*protocol.MemoryUser); ok {
			users[id] = memoryUser
		}
		return true
	})
	return users
}
