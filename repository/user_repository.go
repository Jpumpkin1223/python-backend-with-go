package repository

import (
	"fmt"
	"sync"

	"python-backend-with-go/models"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user models.User) error
	GetByID(id int) (models.User, error)
	GetByEmail(email string) (models.User, error)
	EmailExists(email string) bool
	GetNextUserID() int
}

// InMemoryUserRepository implements UserRepository using in-memory storage
type InMemoryUserRepository struct {
	users      map[int]models.User
	nextUserID int
	mu         sync.RWMutex
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:      make(map[int]models.User),
		nextUserID: 1,
	}
}

// Create adds a new user to the repository
func (r *InMemoryUserRepository) Create(user models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID] = user
	r.nextUserID++
	return nil
}

// GetByID retrieves a user by ID
func (r *InMemoryUserRepository) GetByID(id int) (models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return models.User{}, fmt.Errorf("user not found")
	}
	return user, nil
}

// GetByEmail retrieves a user by email
func (r *InMemoryUserRepository) GetByEmail(email string) (models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return models.User{}, fmt.Errorf("user not found")
}

// EmailExists checks if an email already exists
func (r *InMemoryUserRepository) EmailExists(email string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return true
		}
	}
	return false
}

// GetNextUserID returns the next available user ID
func (r *InMemoryUserRepository) GetNextUserID() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.nextUserID
}
