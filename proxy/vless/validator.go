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
	Add(u *protocol.MemoryUser) error
	Del(email string) error
	GetByEmail(email string) *protocol.MemoryUser
	GetAll() map[string]*protocol.MemoryUser
	GetCount() int64
	GetAllEmails() []string
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

// Get a VLESS user with email, nil if user doesn't exist.
func (v *MemoryValidator) GetByEmail(email string) *protocol.MemoryUser {
	u, _ := v.email.Load(email)
	if u != nil {
		return u.(*protocol.MemoryUser)
	}
	return nil
}

func (v *MemoryValidator) GetAll() map[string]*protocol.MemoryUser {
	var user any
	var memoryUser *protocol.MemoryUser
	var id string
	var ok bool
	u := map[string]*protocol.MemoryUser{}

	v.email.Range(func(key, value any) bool {
		if id, ok = key.(string); !ok {
			return true
		}
		if user, ok = v.users.Load(id); !ok {
			return true
		}
		if memoryUser, ok = user.(*protocol.MemoryUser); ok {
			u[id] = memoryUser
		}
		return true
	})
	return u
}

// Get users count
func (v *MemoryValidator) GetCount() int64 {
	var c int64 = 0
	v.email.Range(func(key, value interface{}) bool {
		c++
		return true
	})
	return c
}

func (v *MemoryValidator) GetAllEmails() []string {
	emails := make([]string, 0)
	v.email.Range(func(key, value any) bool {
		email, ok := key.(string)
		if !ok {
			return true
		}
		emails = append(emails, email)
		return true
	})
	return emails
}
