package repository

import (
	"fmt"
	"sort"
	"sync"

	"gorm.io/gorm"
	"python-backend-with-go/models"
)

// PostRepository defines the interface for post data operations
type PostRepository interface {
	Create(post *models.Post) error
	Update(post *models.Post) error
	Delete(postID int) error
	GetByID(postID int) (models.Post, error)
	GetByUserID(userID int) ([]models.Post, error)
	GetByUserIDs(userIDs []int) ([]models.Post, error)
}

// GormPostRepository implements PostRepository using GORM
type GormPostRepository struct {
	db *gorm.DB
}

// NewGormPostRepository creates a new GORM post repository
func NewGormPostRepository(db *gorm.DB) *GormPostRepository {
	return &GormPostRepository{db: db}
}

// Create adds a new post to the database
func (r *GormPostRepository) Create(post *models.Post) error {
	return r.db.Create(post).Error
}

// Update updates an existing post in the database
func (r *GormPostRepository) Update(post *models.Post) error {
	result := r.db.Model(post).Where("id = ?", post.ID).Update("tweet", post.Content)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("post not found")
	}
	return nil
}

// Delete removes a post from the database
func (r *GormPostRepository) Delete(postID int) error {
	result := r.db.Delete(&models.Post{}, postID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("post not found")
	}
	return nil
}

// GetByID retrieves a post by ID
func (r *GormPostRepository) GetByID(postID int) (models.Post, error) {
	var post models.Post
	err := r.db.First(&post, postID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Post{}, fmt.Errorf("post not found")
		}
		return models.Post{}, err
	}
	return post, nil
}

// GetByUserID retrieves all posts by a specific user
func (r *GormPostRepository) GetByUserID(userID int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// GetByUserIDs retrieves all posts by multiple users
func (r *GormPostRepository) GetByUserIDs(userIDs []int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Where("user_id IN ?", userIDs).Order("created_at DESC").Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// InMemoryPostRepository implements PostRepository using in-memory storage
type InMemoryPostRepository struct {
	posts      map[int]models.Post
	userPosts  map[int][]int // key: userID, value: list of post IDs
	nextPostID int
	mu         sync.RWMutex
}

// NewInMemoryPostRepository creates a new in-memory post repository
func NewInMemoryPostRepository() *InMemoryPostRepository {
	return &InMemoryPostRepository{
		posts:      make(map[int]models.Post),
		userPosts:  make(map[int][]int),
		nextPostID: 1,
	}
}

// Create adds a new post to the repository
func (r *InMemoryPostRepository) Create(post *models.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post.ID = r.nextPostID
	r.posts[post.ID] = *post
	r.userPosts[post.UserID] = append(r.userPosts[post.UserID], post.ID)
	r.nextPostID++
	return nil
}

// Update updates an existing post in the repository
func (r *InMemoryPostRepository) Update(post *models.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.posts[post.ID]; !exists {
		return fmt.Errorf("post not found")
	}

	r.posts[post.ID] = *post
	return nil
}

// Delete removes a post from the repository
func (r *InMemoryPostRepository) Delete(postID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, exists := r.posts[postID]
	if !exists {
		return fmt.Errorf("post not found")
	}

	// Remove from posts map
	delete(r.posts, postID)

	// Remove from userPosts index
	userPostIDs := r.userPosts[post.UserID]
	for i, id := range userPostIDs {
		if id == postID {
			r.userPosts[post.UserID] = append(userPostIDs[:i], userPostIDs[i+1:]...)
			break
		}
	}

	return nil
}

// GetByID retrieves a post by ID
func (r *InMemoryPostRepository) GetByID(postID int) (models.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, exists := r.posts[postID]
	if !exists {
		return models.Post{}, fmt.Errorf("post not found")
	}
	return post, nil
}

// GetByUserID retrieves all posts by a specific user
func (r *InMemoryPostRepository) GetByUserID(userID int) ([]models.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	postIDs := r.userPosts[userID]
	posts := make([]models.Post, 0, len(postIDs))

	for _, postID := range postIDs {
		if post, exists := r.posts[postID]; exists {
			posts = append(posts, post)
		}
	}

	// Sort by created_at descending (newest first)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})

	return posts, nil
}

// GetByUserIDs retrieves all posts by multiple users
func (r *InMemoryPostRepository) GetByUserIDs(userIDs []int) ([]models.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	posts := make([]models.Post, 0)

	for _, userID := range userIDs {
		postIDs := r.userPosts[userID]
		for _, postID := range postIDs {
			if post, exists := r.posts[postID]; exists {
				posts = append(posts, post)
			}
		}
	}

	// Sort by created_at descending (newest first)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})

	return posts, nil
}

// GetNextPostID returns the next available post ID
func (r *InMemoryPostRepository) GetNextPostID() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.nextPostID
}
