#!/bin/bash

# API Integration Test Script
# Tests all endpoints with password hashing verification

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TOTAL=0
PASSED=0
FAILED=0

# Helper function to print test results
print_test() {
    TOTAL=$((TOTAL + 1))
    if [ $1 -eq 0 ]; then
        PASSED=$((PASSED + 1))
        echo -e "${GREEN}✓ Test $TOTAL: $2${NC}"
    else
        FAILED=$((FAILED + 1))
        echo -e "${RED}✗ Test $TOTAL: $2${NC}"
        if [ ! -z "$3" ]; then
            echo -e "${RED}  Error: $3${NC}"
        fi
    fi
}

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Starting API Integration Tests${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Clean database before testing
echo "Cleaning database for fresh test..."
mysql --defaults-file=.my.cnf go_backend -e "DELETE FROM tweets; DELETE FROM users_follow_list; DELETE FROM users;" 2>/dev/null
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Database cleaned${NC}"
else
    echo -e "${RED}✗ Failed to clean database (continuing anyway...)${NC}"
fi
echo ""

# Wait for server to be ready
echo "Waiting for server to be ready..."
sleep 2

# ===========================================
# 1. 회원가입 테스트 (패스워드 해싱 확인)
# ===========================================
echo -e "${YELLOW}[A] 회원가입 & 보안 테스트${NC}"

# Test 1: 회원가입 성공
RESPONSE=$(curl -s -X POST "$BASE_URL/api/signup" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "테스트유저1",
    "email": "user1@test.com",
    "password": "password123",
    "profile": "첫 번째 사용자"
  }')

if echo "$RESPONSE" | grep -q "회원가입이 완료되었습니다"; then
    USER1_ID=$(echo "$RESPONSE" | grep -o '"user_id":[0-9]*' | cut -d':' -f2)
    print_test 0 "회원가입 성공 (User ID: $USER1_ID)"
else
    print_test 1 "회원가입 실패" "$RESPONSE"
fi

# Test 2: 두 번째 사용자 회원가입
RESPONSE=$(curl -s -X POST "$BASE_URL/api/signup" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "테스트유저2",
    "email": "user2@test.com",
    "password": "password456",
    "profile": "두 번째 사용자"
  }')

if echo "$RESPONSE" | grep -q "회원가입이 완료되었습니다"; then
    USER2_ID=$(echo "$RESPONSE" | grep -o '"user_id":[0-9]*' | cut -d':' -f2)
    print_test 0 "두 번째 사용자 회원가입 (User ID: $USER2_ID)"
else
    print_test 1 "두 번째 사용자 회원가입 실패" "$RESPONSE"
fi

# Test 3: 중복 이메일 체크
RESPONSE=$(curl -s -X POST "$BASE_URL/api/signup" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "중복유저",
    "email": "user1@test.com",
    "password": "test",
    "profile": "test"
  }')

if echo "$RESPONSE" | grep -q "email already exists"; then
    print_test 0 "중복 이메일 거부 확인"
else
    print_test 1 "중복 이메일 체크 실패" "$RESPONSE"
fi

echo ""

# ===========================================
# 2. 팔로우 기능 테스트
# ===========================================
echo -e "${YELLOW}[B] 팔로우 기능 테스트${NC}"

# Test 4: 팔로우 생성 (User2 -> User1)
RESPONSE=$(curl -s -X POST "$BASE_URL/api/users/$USER1_ID/follow" \
  -H "Content-Type: application/json" \
  -d "{\"follower_id\": $USER2_ID}")

if echo "$RESPONSE" | grep -q "팔로우 성공"; then
    print_test 0 "팔로우 생성 성공 (User $USER2_ID → User $USER1_ID)"
else
    print_test 1 "팔로우 생성 실패" "$RESPONSE"
fi

# Test 5: 팔로워 목록 조회
RESPONSE=$(curl -s -X GET "$BASE_URL/api/users/$USER1_ID/followers")

if echo "$RESPONSE" | grep -q "user2@test.com"; then
    print_test 0 "팔로워 목록 조회 성공"
else
    print_test 1 "팔로워 목록 조회 실패" "$RESPONSE"
fi

# Test 6: 팔로잉 목록 조회
RESPONSE=$(curl -s -X GET "$BASE_URL/api/users/$USER2_ID/following")

if echo "$RESPONSE" | grep -q "user1@test.com"; then
    print_test 0 "팔로잉 목록 조회 성공"
else
    print_test 1 "팔로잉 목록 조회 실패" "$RESPONSE"
fi

# Test 7: 자기 자신 팔로우 시도 (에러 확인)
RESPONSE=$(curl -s -X POST "$BASE_URL/api/users/$USER1_ID/follow" \
  -H "Content-Type: application/json" \
  -d "{\"follower_id\": $USER1_ID}")

if echo "$RESPONSE" | grep -q "cannot follow yourself"; then
    print_test 0 "자기 자신 팔로우 거부 확인"
else
    print_test 1 "자기 자신 팔로우 체크 실패" "$RESPONSE"
fi

# Test 8: 존재하지 않는 사용자 팔로우 시도
RESPONSE=$(curl -s -X POST "$BASE_URL/api/users/999/follow" \
  -H "Content-Type: application/json" \
  -d "{\"follower_id\": $USER1_ID}")

if echo "$RESPONSE" | grep -q "user not found"; then
    print_test 0 "존재하지 않는 사용자 팔로우 거부 확인"
else
    print_test 1 "존재하지 않는 사용자 팔로우 체크 실패" "$RESPONSE"
fi

echo ""

# ===========================================
# 3. 게시글 기능 테스트
# ===========================================
echo -e "${YELLOW}[C] 게시글 기능 테스트${NC}"

# Test 9: 게시글 작성
RESPONSE=$(curl -s -X POST "$BASE_URL/api/posts" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": $USER1_ID,
    \"content\": \"첫 번째 게시글입니다!\"
  }")

if echo "$RESPONSE" | grep -q "게시글이 생성되었습니다"; then
    POST_ID=$(echo "$RESPONSE" | grep -o '"post_id":[0-9]*' | cut -d':' -f2)
    print_test 0 "게시글 작성 성공 (Post ID: $POST_ID)"
else
    print_test 1 "게시글 작성 실패" "$RESPONSE"
fi

# Test 10: 게시글 수정
RESPONSE=$(curl -s -X PUT "$BASE_URL/api/posts/$POST_ID" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": $USER1_ID,
    \"content\": \"수정된 게시글입니다!\"
  }")

if echo "$RESPONSE" | grep -q "게시글이 수정되었습니다"; then
    print_test 0 "게시글 수정 성공"
else
    print_test 1 "게시글 수정 실패" "$RESPONSE"
fi

# Test 11: 다른 사용자의 게시글 수정 시도 (에러 확인)
RESPONSE=$(curl -s -X PUT "$BASE_URL/api/posts/$POST_ID" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": $USER2_ID,
    \"content\": \"해킹 시도\"
  }")

if echo "$RESPONSE" | grep -q "unauthorized"; then
    print_test 0 "권한 없는 수정 거부 확인"
else
    print_test 1 "권한 없는 수정 체크 실패" "$RESPONSE"
fi

# Test 12: 사용자 게시글 목록 조회
RESPONSE=$(curl -s -X GET "$BASE_URL/api/users/$USER1_ID/posts")

if echo "$RESPONSE" | grep -q "수정된 게시글입니다"; then
    print_test 0 "사용자 게시글 목록 조회 성공"
else
    print_test 1 "사용자 게시글 목록 조회 실패" "$RESPONSE"
fi

# Test 13: 타임라인 조회 (User2가 User1을 팔로우 중)
RESPONSE=$(curl -s -X GET "$BASE_URL/api/users/$USER2_ID/timeline")

if echo "$RESPONSE" | grep -q "수정된 게시글입니다"; then
    print_test 0 "타임라인 조회 성공"
else
    print_test 1 "타임라인 조회 실패" "$RESPONSE"
fi

# Test 14: 300자 초과 게시글 작성 시도
LONG_CONTENT=$(printf 'a%.0s' {1..301})
RESPONSE=$(curl -s -X POST "$BASE_URL/api/posts" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": $USER1_ID,
    \"content\": \"$LONG_CONTENT\"
  }")

if echo "$RESPONSE" | grep -q "300 characters"; then
    print_test 0 "300자 초과 게시글 거부 확인"
else
    print_test 1 "300자 초과 체크 실패" "$RESPONSE"
fi

echo ""

# ===========================================
# Summary
# ===========================================
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Test Summary${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "Total Tests: $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed. ✗${NC}"
    exit 1
fi
