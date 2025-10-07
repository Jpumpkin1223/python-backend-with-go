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
# 2. 로그인 & JWT 인증 테스트
# ===========================================
echo -e "${YELLOW}[B] 로그인 & JWT 인증 테스트${NC}"

# Test 4: 로그인 성공
RESPONSE=$(curl -s -X POST "$BASE_URL/api/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user1@test.com",
    "password": "password123"
  }')

if echo "$RESPONSE" | grep -q "로그인 성공"; then
    USER1_TOKEN=$(echo "$RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    print_test 0 "로그인 성공 (토큰 획득)"
else
    print_test 1 "로그인 실패" "$RESPONSE"
fi

# Test 5: 두 번째 사용자 로그인
RESPONSE=$(curl -s -X POST "$BASE_URL/api/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user2@test.com",
    "password": "password456"
  }')

if echo "$RESPONSE" | grep -q "로그인 성공"; then
    USER2_TOKEN=$(echo "$RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    print_test 0 "두 번째 사용자 로그인 성공"
else
    print_test 1 "두 번째 사용자 로그인 실패" "$RESPONSE"
fi

# Test 6: 잘못된 비밀번호로 로그인 시도
RESPONSE=$(curl -s -X POST "$BASE_URL/api/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user1@test.com",
    "password": "wrongpassword"
  }')

if echo "$RESPONSE" | grep -q "invalid email or password"; then
    print_test 0 "잘못된 비밀번호 거부 확인"
else
    print_test 1 "잘못된 비밀번호 체크 실패" "$RESPONSE"
fi

# Test 7: 존재하지 않는 이메일로 로그인 시도
RESPONSE=$(curl -s -X POST "$BASE_URL/api/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "nonexistent@test.com",
    "password": "password123"
  }')

if echo "$RESPONSE" | grep -q "invalid email or password"; then
    print_test 0 "존재하지 않는 이메일 거부 확인"
else
    print_test 1 "존재하지 않는 이메일 체크 실패" "$RESPONSE"
fi

echo ""

# ===========================================
# 3. JWT 인증 필수 API 테스트
# ===========================================
echo -e "${YELLOW}[C] JWT 인증 필수 API 테스트${NC}"

# Test 8: 토큰 없이 팔로우 시도 (401 에러 확인)
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/users/$USER1_ID/follow" \
  -H "Content-Type: application/json" \
  -d "{\"follower_id\": $USER2_ID}")

HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
if [ "$HTTP_CODE" = "401" ]; then
    print_test 0 "토큰 없이 팔로우 거부 확인 (401)"
else
    print_test 1 "토큰 없이 팔로우 체크 실패 (HTTP $HTTP_CODE)" "$RESPONSE"
fi

# Test 9: 유효하지 않은 토큰으로 팔로우 시도
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/users/$USER1_ID/follow" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer invalid.token.here" \
  -d "{\"follower_id\": $USER2_ID}")

HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
if [ "$HTTP_CODE" = "401" ]; then
    print_test 0 "유효하지 않은 토큰 거부 확인 (401)"
else
    print_test 1 "유효하지 않은 토큰 체크 실패 (HTTP $HTTP_CODE)" "$RESPONSE"
fi

# Test 10: 토큰 없이 게시글 작성 시도
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\": $USER1_ID, \"content\": \"테스트\"}")

HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
if [ "$HTTP_CODE" = "401" ]; then
    print_test 0 "토큰 없이 게시글 작성 거부 확인 (401)"
else
    print_test 1 "토큰 없이 게시글 작성 체크 실패 (HTTP $HTTP_CODE)" "$RESPONSE"
fi

echo ""

# ===========================================
# 4. 팔로우 기능 테스트 (인증 포함)
# ===========================================
echo -e "${YELLOW}[D] 팔로우 기능 테스트 (인증 포함)${NC}"

# Test 11: 팔로우 생성 (User2 -> User1) with JWT
RESPONSE=$(curl -s -X POST "$BASE_URL/api/users/$USER1_ID/follow" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER2_TOKEN" \
  -d "{\"follower_id\": $USER2_ID}")

if echo "$RESPONSE" | grep -q "팔로우 성공"; then
    print_test 0 "JWT 인증으로 팔로우 생성 성공 (User $USER2_ID → User $USER1_ID)"
else
    print_test 1 "팔로우 생성 실패" "$RESPONSE"
fi

# Test 12: 팔로워 목록 조회 (공개 API)
RESPONSE=$(curl -s -X GET "$BASE_URL/api/users/$USER1_ID/followers")

if echo "$RESPONSE" | grep -q "user2@test.com"; then
    print_test 0 "팔로워 목록 조회 성공"
else
    print_test 1 "팔로워 목록 조회 실패" "$RESPONSE"
fi

# Test 13: 팔로잉 목록 조회 (공개 API)
RESPONSE=$(curl -s -X GET "$BASE_URL/api/users/$USER2_ID/following")

if echo "$RESPONSE" | grep -q "user1@test.com"; then
    print_test 0 "팔로잉 목록 조회 성공"
else
    print_test 1 "팔로잉 목록 조회 실패" "$RESPONSE"
fi

# Test 14: 자기 자신 팔로우 시도 (에러 확인)
RESPONSE=$(curl -s -X POST "$BASE_URL/api/users/$USER1_ID/follow" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER1_TOKEN" \
  -d "{\"follower_id\": $USER1_ID}")

if echo "$RESPONSE" | grep -q "cannot follow yourself"; then
    print_test 0 "자기 자신 팔로우 거부 확인"
else
    print_test 1 "자기 자신 팔로우 체크 실패" "$RESPONSE"
fi

# Test 15: 언팔로우 (User2 -> User1) with JWT
RESPONSE=$(curl -s -X DELETE "$BASE_URL/api/users/$USER1_ID/follow" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER2_TOKEN" \
  -d "{\"follower_id\": $USER2_ID}")

if echo "$RESPONSE" | grep -q "언팔로우 성공"; then
    print_test 0 "JWT 인증으로 언팔로우 성공"
else
    print_test 1 "언팔로우 실패" "$RESPONSE"
fi

# Test 16: 다시 팔로우 (테스트를 위해)
RESPONSE=$(curl -s -X POST "$BASE_URL/api/users/$USER1_ID/follow" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER2_TOKEN" \
  -d "{\"follower_id\": $USER2_ID}")

if echo "$RESPONSE" | grep -q "팔로우 성공"; then
    print_test 0 "다시 팔로우 성공 (게시글 테스트 준비)"
else
    print_test 1 "다시 팔로우 실패" "$RESPONSE"
fi

echo ""

# ===========================================
# 5. 게시글 기능 테스트 (인증 포함)
# ===========================================
echo -e "${YELLOW}[E] 게시글 기능 테스트 (인증 포함)${NC}"

# Test 17: 게시글 작성 with JWT
RESPONSE=$(curl -s -X POST "$BASE_URL/api/posts" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER1_TOKEN" \
  -d "{
    \"user_id\": $USER1_ID,
    \"content\": \"첫 번째 게시글입니다!\"
  }")

if echo "$RESPONSE" | grep -q "게시글이 생성되었습니다"; then
    POST_ID=$(echo "$RESPONSE" | grep -o '"post_id":[0-9]*' | cut -d':' -f2)
    print_test 0 "JWT 인증으로 게시글 작성 성공 (Post ID: $POST_ID)"
else
    print_test 1 "게시글 작성 실패" "$RESPONSE"
fi

# Test 18: 게시글 수정 with JWT
RESPONSE=$(curl -s -X PUT "$BASE_URL/api/posts/$POST_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER1_TOKEN" \
  -d "{
    \"user_id\": $USER1_ID,
    \"content\": \"수정된 게시글입니다!\"
  }")

if echo "$RESPONSE" | grep -q "게시글이 수정되었습니다"; then
    print_test 0 "JWT 인증으로 게시글 수정 성공"
else
    print_test 1 "게시글 수정 실패" "$RESPONSE"
fi

# Test 19: 다른 사용자의 게시글 수정 시도 (에러 확인)
RESPONSE=$(curl -s -X PUT "$BASE_URL/api/posts/$POST_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER2_TOKEN" \
  -d "{
    \"user_id\": $USER2_ID,
    \"content\": \"해킹 시도\"
  }")

if echo "$RESPONSE" | grep -q "unauthorized"; then
    print_test 0 "권한 없는 게시글 수정 거부 확인"
else
    print_test 1 "권한 없는 수정 체크 실패" "$RESPONSE"
fi

# Test 20: 사용자 게시글 목록 조회 (공개 API)
RESPONSE=$(curl -s -X GET "$BASE_URL/api/users/$USER1_ID/posts")

if echo "$RESPONSE" | grep -q "수정된 게시글입니다"; then
    print_test 0 "사용자 게시글 목록 조회 성공"
else
    print_test 1 "사용자 게시글 목록 조회 실패" "$RESPONSE"
fi

# Test 21: 타임라인 조회 (User2가 User1을 팔로우 중) (공개 API)
RESPONSE=$(curl -s -X GET "$BASE_URL/api/users/$USER2_ID/timeline")

if echo "$RESPONSE" | grep -q "수정된 게시글입니다"; then
    print_test 0 "타임라인 조회 성공"
else
    print_test 1 "타임라인 조회 실패" "$RESPONSE"
fi

# Test 22: 300자 초과 게시글 작성 시도
LONG_CONTENT=$(printf 'a%.0s' {1..301})
RESPONSE=$(curl -s -X POST "$BASE_URL/api/posts" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER1_TOKEN" \
  -d "{
    \"user_id\": $USER1_ID,
    \"content\": \"$LONG_CONTENT\"
  }")

if echo "$RESPONSE" | grep -q "300 characters"; then
    print_test 0 "300자 초과 게시글 거부 확인"
else
    print_test 1 "300자 초과 체크 실패" "$RESPONSE"
fi

# Test 23: 게시글 삭제 with JWT
RESPONSE=$(curl -s -X DELETE "$BASE_URL/api/posts/$POST_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER1_TOKEN" \
  -d "{\"user_id\": $USER1_ID}")

if echo "$RESPONSE" | grep -q "게시글이 삭제되었습니다"; then
    print_test 0 "JWT 인증으로 게시글 삭제 성공"
else
    print_test 1 "게시글 삭제 실패" "$RESPONSE"
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
