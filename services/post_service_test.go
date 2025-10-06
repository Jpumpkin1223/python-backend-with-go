package services

import (
	"strings"
	"testing"

	"python-backend-with-go/models"
	"python-backend-with-go/repository"
)

func setupPostServiceTest(t *testing.T) (*PostService, *UserService, *FollowService) {
	userRepo := repository.NewInMemoryUserRepository()
	followRepo := repository.NewInMemoryFollowRepository()
	postRepo := repository.NewInMemoryPostRepository()

	userService := NewUserService(userRepo)
	followService := NewFollowService(followRepo, userRepo)
	postService := NewPostService(postRepo, userRepo, followRepo)

	// Create test users
	for i := 1; i <= 3; i++ {
		userRepo.Create(models.User{
			ID:    i,
			Name:  "User" + string(rune('0'+i)),
			Email: "user" + string(rune('0'+i)) + "@test.com",
		})
	}

	return postService, userService, followService
}

func TestPostService_CreatePost(t *testing.T) {
	tests := []struct {
		name        string
		request     models.CreatePostRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful post creation",
			request: models.CreatePostRequest{
				UserID:  1,
				Content: "안녕하세요, 첫 게시글입니다!",
			},
			expectError: false,
		},
		{
			name: "missing user_id",
			request: models.CreatePostRequest{
				Content: "테스트 게시글",
			},
			expectError: true,
			errorMsg:    "user_id is required",
		},
		{
			name: "user not found",
			request: models.CreatePostRequest{
				UserID:  999,
				Content: "테스트 게시글",
			},
			expectError: true,
			errorMsg:    "user not found",
		},
		{
			name: "missing content",
			request: models.CreatePostRequest{
				UserID: 1,
			},
			expectError: true,
			errorMsg:    "content is required",
		},
		{
			name: "content too long",
			request: models.CreatePostRequest{
				UserID:  1,
				Content: strings.Repeat("a", 301),
			},
			expectError: true,
			errorMsg:    "content must be 300 characters or less",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postService, _, _ := setupPostServiceTest(t)

			resp, err := postService.CreatePost(tt.request)

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
				if resp.PostID != 1 {
					t.Errorf("Expected post_id 1, got %d", resp.PostID)
				}
			}
		})
	}
}

func TestPostService_UpdatePost(t *testing.T) {
	postService, _, _ := setupPostServiceTest(t)

	// Create a post first
	createResp, err := postService.CreatePost(models.CreatePostRequest{
		UserID:  1,
		Content: "원본 게시글",
	})
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	tests := []struct {
		name        string
		postID      int
		request     models.UpdatePostRequest
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful update",
			postID: createResp.PostID,
			request: models.UpdatePostRequest{
				UserID:  1,
				Content: "수정된 게시글",
			},
			expectError: false,
		},
		{
			name:   "unauthorized update",
			postID: createResp.PostID,
			request: models.UpdatePostRequest{
				UserID:  2,
				Content: "다른 사용자가 수정 시도",
			},
			expectError: true,
			errorMsg:    "unauthorized to update this post",
		},
		{
			name:   "post not found",
			postID: 999,
			request: models.UpdatePostRequest{
				UserID:  1,
				Content: "존재하지 않는 게시글",
			},
			expectError: true,
			errorMsg:    "post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := postService.UpdatePost(tt.postID, tt.request)

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
				if resp.Post.Content != tt.request.Content {
					t.Errorf("Expected content '%s', got '%s'", tt.request.Content, resp.Post.Content)
				}
			}
		})
	}
}

func TestPostService_DeletePost(t *testing.T) {
	postService, _, _ := setupPostServiceTest(t)

	// Create a post first
	createResp, err := postService.CreatePost(models.CreatePostRequest{
		UserID:  1,
		Content: "삭제할 게시글",
	})
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	// Test unauthorized deletion
	_, err = postService.DeletePost(createResp.PostID, 2)
	if err == nil {
		t.Error("Expected error for unauthorized deletion, got none")
	} else if err.Error() != "unauthorized to delete this post" {
		t.Errorf("Expected 'unauthorized to delete this post', got '%s'", err.Error())
	}

	// Test successful deletion
	resp, err := postService.DeletePost(createResp.PostID, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp.PostID != createResp.PostID {
		t.Errorf("Expected post_id %d, got %d", createResp.PostID, resp.PostID)
	}

	// Verify post is deleted
	_, err = postService.DeletePost(createResp.PostID, 1)
	if err == nil {
		t.Error("Expected error for deleting non-existent post, got none")
	}
}

func TestPostService_GetUserPosts(t *testing.T) {
	postService, _, _ := setupPostServiceTest(t)

	// Create posts for user 1
	postService.CreatePost(models.CreatePostRequest{
		UserID:  1,
		Content: "첫 번째 게시글",
	})
	postService.CreatePost(models.CreatePostRequest{
		UserID:  1,
		Content: "두 번째 게시글",
	})

	// Get user posts
	resp, err := postService.GetUserPosts(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.Count != 2 {
		t.Errorf("Expected count 2, got %d", resp.Count)
	}
	if len(resp.Posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(resp.Posts))
	}
}

func TestPostService_GetTimeline(t *testing.T) {
	postService, _, followService := setupPostServiceTest(t)

	// User 1 follows User 2 and 3
	followService.Follow(1, 2)
	followService.Follow(1, 3)

	// User 2 and 3 create posts
	postService.CreatePost(models.CreatePostRequest{
		UserID:  2,
		Content: "User 2의 게시글",
	})
	postService.CreatePost(models.CreatePostRequest{
		UserID:  3,
		Content: "User 3의 게시글",
	})

	// Get timeline for User 1
	resp, err := postService.GetTimeline(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.Count != 2 {
		t.Errorf("Expected count 2, got %d", resp.Count)
	}
	if len(resp.Posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(resp.Posts))
	}

	// Verify posts are from followed users
	for _, post := range resp.Posts {
		if post.UserID != 2 && post.UserID != 3 {
			t.Errorf("Unexpected user_id %d in timeline", post.UserID)
		}
	}
}

func TestPostService_GetTimeline_EmptyWhenNotFollowing(t *testing.T) {
	postService, _, _ := setupPostServiceTest(t)

	// User 2 creates a post
	postService.CreatePost(models.CreatePostRequest{
		UserID:  2,
		Content: "User 2의 게시글",
	})

	// Get timeline for User 1 (not following anyone)
	resp, err := postService.GetTimeline(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.Count != 0 {
		t.Errorf("Expected count 0, got %d", resp.Count)
	}
	if len(resp.Posts) != 0 {
		t.Errorf("Expected 0 posts, got %d", len(resp.Posts))
	}
}
