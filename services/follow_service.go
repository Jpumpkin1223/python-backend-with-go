package services

import (
	"fmt"
	"time"

	"python-backend-with-go/models"
	"python-backend-with-go/repository"
)

// FollowService handles follow business logic
type FollowService struct {
	followRepo repository.FollowRepository
	userRepo   repository.UserRepository
}

// NewFollowService creates a new follow service
func NewFollowService(followRepo repository.FollowRepository, userRepo repository.UserRepository) *FollowService {
	return &FollowService{
		followRepo: followRepo,
		userRepo:   userRepo,
	}
}

// Follow creates a follow relationship
func (s *FollowService) Follow(followerID, followingID int) (models.FollowResponse, error) {
	// Validate IDs
	if followerID == 0 || followingID == 0 {
		return models.FollowResponse{}, fmt.Errorf("follower_id and following user ID are required")
	}

	// Check if trying to follow themselves
	if followerID == followingID {
		return models.FollowResponse{}, fmt.Errorf("cannot follow yourself")
	}

	// Check if both users exist
	if _, err := s.userRepo.GetByID(followerID); err != nil {
		return models.FollowResponse{}, fmt.Errorf("follower user not found")
	}
	if _, err := s.userRepo.GetByID(followingID); err != nil {
		return models.FollowResponse{}, fmt.Errorf("following user not found")
	}

	// Check if already following
	if s.followRepo.Exists(followerID, followingID) {
		return models.FollowResponse{}, fmt.Errorf("already following this user")
	}

	// Create follow relationship
	now := time.Now()
	follow := models.Follow{
		UserID:       followerID,
		FollowUserID: followingID,
		CreatedAt:    now,
	}

	if err := s.followRepo.Create(follow); err != nil {
		return models.FollowResponse{}, fmt.Errorf("failed to create follow: %w", err)
	}

	return models.FollowResponse{
		Message:     "팔로우 성공",
		FollowerID:  followerID,
		FollowingID: followingID,
		CreatedAt:   now.Format(time.RFC3339),
	}, nil
}

// Unfollow removes a follow relationship
func (s *FollowService) Unfollow(followerID, followingID int) (models.FollowResponse, error) {
	// Validate IDs
	if followerID == 0 || followingID == 0 {
		return models.FollowResponse{}, fmt.Errorf("follower_id and following user ID are required")
	}

	// Delete follow relationship
	if err := s.followRepo.Delete(followerID, followingID); err != nil {
		return models.FollowResponse{}, err
	}

	return models.FollowResponse{
		Message:     "언팔로우 성공",
		FollowerID:  followerID,
		FollowingID: followingID,
	}, nil
}

// GetFollowers retrieves followers for a user
func (s *FollowService) GetFollowers(userID int) (models.FollowListResponse, error) {
	// Check if user exists
	if _, err := s.userRepo.GetByID(userID); err != nil {
		return models.FollowListResponse{}, fmt.Errorf("user not found")
	}

	// Get follower IDs
	followerIDs, err := s.followRepo.GetFollowers(userID)
	if err != nil {
		return models.FollowListResponse{}, fmt.Errorf("failed to get followers: %w", err)
	}

	// Convert to user info
	users := make([]models.UserInfo, 0, len(followerIDs))
	for _, id := range followerIDs {
		user, err := s.userRepo.GetByID(id)
		if err != nil {
			continue
		}
		users = append(users, models.UserInfo{
			ID:      user.ID,
			Name:    user.Name,
			Email:   user.Email,
			Profile: user.Profile,
		})
	}

	return models.FollowListResponse{
		Users: users,
		Count: len(users),
	}, nil
}

// GetFollowing retrieves following for a user
func (s *FollowService) GetFollowing(userID int) (models.FollowListResponse, error) {
	// Check if user exists
	if _, err := s.userRepo.GetByID(userID); err != nil {
		return models.FollowListResponse{}, fmt.Errorf("user not found")
	}

	// Get following IDs
	followingIDs, err := s.followRepo.GetFollowing(userID)
	if err != nil {
		return models.FollowListResponse{}, fmt.Errorf("failed to get following: %w", err)
	}

	// Convert to user info
	users := make([]models.UserInfo, 0, len(followingIDs))
	for _, id := range followingIDs {
		user, err := s.userRepo.GetByID(id)
		if err != nil {
			continue
		}
		users = append(users, models.UserInfo{
			ID:      user.ID,
			Name:    user.Name,
			Email:   user.Email,
			Profile: user.Profile,
		})
	}

	return models.FollowListResponse{
		Users: users,
		Count: len(users),
	}, nil
}

// GetFollowStatus checks if a user is following another user
func (s *FollowService) GetFollowStatus(followerID, followingID int) models.FollowStatusResponse {
	isFollowing := s.followRepo.Exists(followerID, followingID)
	return models.FollowStatusResponse{
		IsFollowing: isFollowing,
		FollowerID:  followerID,
		FollowingID: followingID,
	}
}
