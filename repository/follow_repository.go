package repository

import (
	"fmt"
	"sync"

	"python-backend-with-go/models"
)

// FollowRepository defines the interface for follow data operations
type FollowRepository interface {
	Create(follow models.Follow) error
	Delete(followerID, followingID int) error
	Exists(followerID, followingID int) bool
	GetFollowers(userID int) ([]int, error)
	GetFollowing(userID int) ([]int, error)
}

// InMemoryFollowRepository implements FollowRepository using in-memory storage
type InMemoryFollowRepository struct {
	follows       map[string]models.Follow // key: "followerID:followingID"
	userFollowers map[int]map[int]bool     // key: userID, value: set of follower IDs
	userFollowing map[int]map[int]bool     // key: userID, value: set of following IDs
	mu            sync.RWMutex
}

// NewInMemoryFollowRepository creates a new in-memory follow repository
func NewInMemoryFollowRepository() *InMemoryFollowRepository {
	return &InMemoryFollowRepository{
		follows:       make(map[string]models.Follow),
		userFollowers: make(map[int]map[int]bool),
		userFollowing: make(map[int]map[int]bool),
	}
}

// Create adds a new follow relationship
func (r *InMemoryFollowRepository) Create(follow models.Follow) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	followKey := fmt.Sprintf("%d:%d", follow.FollowerID, follow.FollowingID)
	r.follows[followKey] = follow

	// Update indexes
	if r.userFollowers[follow.FollowingID] == nil {
		r.userFollowers[follow.FollowingID] = make(map[int]bool)
	}
	r.userFollowers[follow.FollowingID][follow.FollowerID] = true

	if r.userFollowing[follow.FollowerID] == nil {
		r.userFollowing[follow.FollowerID] = make(map[int]bool)
	}
	r.userFollowing[follow.FollowerID][follow.FollowingID] = true

	return nil
}

// Delete removes a follow relationship
func (r *InMemoryFollowRepository) Delete(followerID, followingID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	followKey := fmt.Sprintf("%d:%d", followerID, followingID)
	if _, exists := r.follows[followKey]; !exists {
		return fmt.Errorf("follow relationship not found")
	}

	delete(r.follows, followKey)

	// Update indexes
	if r.userFollowers[followingID] != nil {
		delete(r.userFollowers[followingID], followerID)
	}
	if r.userFollowing[followerID] != nil {
		delete(r.userFollowing[followerID], followingID)
	}

	return nil
}

// Exists checks if a follow relationship exists
func (r *InMemoryFollowRepository) Exists(followerID, followingID int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	followKey := fmt.Sprintf("%d:%d", followerID, followingID)
	_, exists := r.follows[followKey]
	return exists
}

// GetFollowers returns the list of follower IDs for a user
func (r *InMemoryFollowRepository) GetFollowers(userID int) ([]int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	followerIDs := r.userFollowers[userID]
	result := make([]int, 0, len(followerIDs))
	for id := range followerIDs {
		result = append(result, id)
	}
	return result, nil
}

// GetFollowing returns the list of following IDs for a user
func (r *InMemoryFollowRepository) GetFollowing(userID int) ([]int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	followingIDs := r.userFollowing[userID]
	result := make([]int, 0, len(followingIDs))
	for id := range followingIDs {
		result = append(result, id)
	}
	return result, nil
}
