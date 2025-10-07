package repository

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
	"python-backend-with-go/models"
)

// FollowRepository defines the interface for follow data operations
type FollowRepository interface {
	Create(follow models.Follow) error
	Delete(userID, followUserID int) error
	Exists(userID, followUserID int) bool
	GetFollowers(userID int) ([]int, error)
	GetFollowing(userID int) ([]int, error)
}

// GormFollowRepository implements FollowRepository using GORM
type GormFollowRepository struct {
	db *gorm.DB
}

// NewGormFollowRepository creates a new GORM follow repository
func NewGormFollowRepository(db *gorm.DB) *GormFollowRepository {
	return &GormFollowRepository{db: db}
}

// Create adds a new follow relationship
func (r *GormFollowRepository) Create(follow models.Follow) error {
	return r.db.Create(&follow).Error
}

// Delete removes a follow relationship
func (r *GormFollowRepository) Delete(userID, followUserID int) error {
	result := r.db.Where("user_id = ? AND follow_user_id = ?", userID, followUserID).Delete(&models.Follow{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("follow relationship not found")
	}
	return nil
}

// Exists checks if a follow relationship exists
func (r *GormFollowRepository) Exists(userID, followUserID int) bool {
	var count int64
	r.db.Model(&models.Follow{}).Where("user_id = ? AND follow_user_id = ?", userID, followUserID).Count(&count)
	return count > 0
}

// GetFollowers returns the list of follower IDs for a user
func (r *GormFollowRepository) GetFollowers(userID int) ([]int, error) {
	var follows []models.Follow
	err := r.db.Where("follow_user_id = ?", userID).Find(&follows).Error
	if err != nil {
		return nil, err
	}

	followerIDs := make([]int, len(follows))
	for i, follow := range follows {
		followerIDs[i] = follow.UserID
	}
	return followerIDs, nil
}

// GetFollowing returns the list of following IDs for a user
func (r *GormFollowRepository) GetFollowing(userID int) ([]int, error) {
	var follows []models.Follow
	err := r.db.Where("user_id = ?", userID).Find(&follows).Error
	if err != nil {
		return nil, err
	}

	followingIDs := make([]int, len(follows))
	for i, follow := range follows {
		followingIDs[i] = follow.FollowUserID
	}
	return followingIDs, nil
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

	followKey := fmt.Sprintf("%d:%d", follow.UserID, follow.FollowUserID)
	r.follows[followKey] = follow

	// Update indexes
	if r.userFollowers[follow.FollowUserID] == nil {
		r.userFollowers[follow.FollowUserID] = make(map[int]bool)
	}
	r.userFollowers[follow.FollowUserID][follow.UserID] = true

	if r.userFollowing[follow.UserID] == nil {
		r.userFollowing[follow.UserID] = make(map[int]bool)
	}
	r.userFollowing[follow.UserID][follow.FollowUserID] = true

	return nil
}

// Delete removes a follow relationship
func (r *InMemoryFollowRepository) Delete(userID, followUserID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	followKey := fmt.Sprintf("%d:%d", userID, followUserID)
	if _, exists := r.follows[followKey]; !exists {
		return fmt.Errorf("follow relationship not found")
	}

	delete(r.follows, followKey)

	// Update indexes
	if r.userFollowers[followUserID] != nil {
		delete(r.userFollowers[followUserID], userID)
	}
	if r.userFollowing[userID] != nil {
		delete(r.userFollowing[userID], followUserID)
	}

	return nil
}

// Exists checks if a follow relationship exists
func (r *InMemoryFollowRepository) Exists(userID, followUserID int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	followKey := fmt.Sprintf("%d:%d", userID, followUserID)
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
