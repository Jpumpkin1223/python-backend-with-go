# Python Backend with Go

Go로 작성된 간단한 백엔드 서버입니다.

## 기능

- 기본 HTTP 서버
- 헬스 체크 엔드포인트 (`/health`)
- API 엔드포인트 예시 (`/api/hello`)
- 환경 변수를 통한 포트 설정

## 실행 방법

### 개발 환경에서 실행

```bash
# 의존성 설치
go mod tidy

# 서버 실행
go run main.go
```

### 빌드 후 실행

```bash
# 빌드
go build -o server main.go

# 실행
./server
```

## 환경 변수

- `PORT`: 서버 포트 (기본값: 8080)

## API 엔드포인트

- `GET /`: 기본 환영 메시지
- `GET /health`: 헬스 체크
- `GET /api/hello`: JSON 응답 예시

## 예시 요청

```bash
# 기본 페이지
curl http://localhost:8080/

# 헬스 체크
curl http://localhost:8080/health

# API 엔드포인트
curl http://localhost:8080/api/hello
```

## 개발

이 프로젝트는 Go 1.19 이상에서 테스트되었습니다.
