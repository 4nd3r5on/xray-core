package trojan

import (
	"strings"
	"sync"

	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/protocol"
)

// Validator stores valid trojan users.
type Validator struct {
	// Considering email's usage here, map + sync.Mutex/RWMutex may have better performance.
	email sync.Map
	users sync.Map
}

// Add a trojan user, Email must be empty or unique.
func (v *Validator) Add(u *protocol.MemoryUser) error {
	if u.Email != "" {
		_, loaded := v.email.LoadOrStore(strings.ToLower(u.Email), u)
		if loaded {
			return errors.New("User ", u.Email, " already exists.")
		}
	}
	v.users.Store(hexString(u.Account.(*MemoryAccount).Key), u)
	return nil
}

// Del a trojan user with a non-empty Email.
func (v *Validator) Del(e string) error {
	if e == "" {
		return errors.New("Email must not be empty.")
	}
	le := strings.ToLower(e)
	u, _ := v.email.Load(le)
	if u == nil {
		return errors.New("User ", e, " not found.")
	}
	v.email.Delete(le)
	v.users.Delete(hexString(u.(*protocol.MemoryUser).Account.(*MemoryAccount).Key))
	return nil
}

// Get a trojan user with hashed key, nil if user doesn't exist.
func (v *Validator) Get(hash string) *protocol.MemoryUser {
	u, _ := v.users.Load(hash)
	if u != nil {
		return u.(*protocol.MemoryUser)
	}
	return nil
}

func (v *Validator) Load(hashOrEmail string) (memoryUser *protocol.MemoryUser, exists bool) {
	var user any
	if user, exists = v.email.Load(hashOrEmail); exists {
		return user.(*protocol.MemoryUser), exists
	}
	if user, exists = v.users.Load(hashOrEmail); exists {
		return user.(*protocol.MemoryUser), exists
	}
	return nil, false
}

func (v *Validator) GetAllIDs() []string {
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

func (v *Validator) GetAll() map[string]*protocol.MemoryUser {
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
