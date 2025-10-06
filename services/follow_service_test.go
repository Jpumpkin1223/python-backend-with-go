package services

import (
	"testing"

	"python-backend-with-go/models"
	"python-backend-with-go/repository"
)

func setupFollowServiceTest(t *testing.T) (*FollowService, *UserService) {
	userRepo := repository.NewInMemoryUserRepository()
	followRepo := repository.NewInMemoryFollowRepository()
	userService := NewUserService(userRepo)
	followService := NewFollowService(followRepo, userRepo)

	// Create test users
	for i := 1; i <= 3; i++ {
		userRepo.Create(models.User{
			ID:    i,
			Name:  "User" + string(rune('0'+i)),
			Email: "user" + string(rune('0'+i)) + "@test.com",
		})
	}

	return followService, userService
}

func TestFollowService_Follow(t *testing.T) {
	tests := []struct {
		name        string
		followerID  int
		followingID int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful follow",
			followerID:  1,
			followingID: 2,
			expectError: false,
		},
		{
			name:        "cannot follow yourself",
			followerID:  1,
			followingID: 1,
			expectError: true,
			errorMsg:    "cannot follow yourself",
		},
		{
			name:        "follower user not found",
			followerID:  999,
			followingID: 2,
			expectError: true,
			errorMsg:    "follower user not found",
		},
		{
			name:        "following user not found",
			followerID:  1,
			followingID: 999,
			expectError: true,
			errorMsg:    "following user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			followService, _ := setupFollowServiceTest(t)

			resp, err := followService.Follow(tt.followerID, tt.followingID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp.FollowerID != tt.followerID || resp.FollowingID != tt.followingID {
					t.Errorf("Expected follower_id=%d, following_id=%d, got follower_id=%d, following_id=%d",
						tt.followerID, tt.followingID, resp.FollowerID, resp.FollowingID)
				}
			}
		})
	}
}

func TestFollowService_Follow_Duplicate(t *testing.T) {
	followService, _ := setupFollowServiceTest(t)

	// First follow
	_, err := followService.Follow(1, 2)
	if err != nil {
		t.Fatalf("Failed to create first follow: %v", err)
	}

	// Try to follow again
	_, err = followService.Follow(1, 2)
	if err == nil {
		t.Error("Expected error for duplicate follow, got none")
	} else if err.Error() != "already following this user" {
		t.Errorf("Expected 'already following this user', got '%s'", err.Error())
	}
}

func TestFollowService_Unfollow(t *testing.T) {
	followService, _ := setupFollowServiceTest(t)

	// Create follow first
	_, err := followService.Follow(1, 2)
	if err != nil {
		t.Fatalf("Failed to create follow: %v", err)
	}

	// Unfollow
	resp, err := followService.Unfollow(1, 2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp.FollowerID != 1 || resp.FollowingID != 2 {
		t.Errorf("Expected follower_id=1, following_id=2, got follower_id=%d, following_id=%d",
			resp.FollowerID, resp.FollowingID)
	}
}

func TestFollowService_Unfollow_NotFollowing(t *testing.T) {
	followService, _ := setupFollowServiceTest(t)

	// Try to unfollow without following
	_, err := followService.Unfollow(1, 2)
	if err == nil {
		t.Error("Expected error for non-existent follow, got none")
	} else if err.Error() != "follow relationship not found" {
		t.Errorf("Expected 'follow relationship not found', got '%s'", err.Error())
	}
}

func TestFollowService_GetFollowers(t *testing.T) {
	followService, _ := setupFollowServiceTest(t)

	// User 2 and 3 follow User 1
	followService.Follow(2, 1)
	followService.Follow(3, 1)

	resp, err := followService.GetFollowers(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.Count != 2 {
		t.Errorf("Expected count 2, got %d", resp.Count)
	}
	if len(resp.Users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(resp.Users))
	}
}

func TestFollowService_GetFollowing(t *testing.T) {
	followService, _ := setupFollowServiceTest(t)

	// User 1 follows User 2 and 3
	followService.Follow(1, 2)
	followService.Follow(1, 3)

	resp, err := followService.GetFollowing(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.Count != 2 {
		t.Errorf("Expected count 2, got %d", resp.Count)
	}
	if len(resp.Users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(resp.Users))
	}
}

func TestFollowService_GetFollowStatus(t *testing.T) {
	followService, _ := setupFollowServiceTest(t)

	// User 1 follows User 2
	followService.Follow(1, 2)

	// Check status (following)
	resp := followService.GetFollowStatus(1, 2)
	if !resp.IsFollowing {
		t.Error("Expected IsFollowing=true, got false")
	}

	// Check status (not following)
	resp = followService.GetFollowStatus(2, 1)
	if resp.IsFollowing {
		t.Error("Expected IsFollowing=false, got true")
	}
}
