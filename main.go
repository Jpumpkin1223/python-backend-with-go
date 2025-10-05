package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// 환경 변수에서 포트 가져오기 (기본값: 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 기본 라우트 핸들러
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "안녕하세요! Go 백엔드 서버입니다. 🚀\n")
		fmt.Fprintf(w, "요청 경로: %s\n", r.URL.Path)
		fmt.Fprintf(w, "요청 메서드: %s\n", r.Method)
	})

	// 헬스 체크 엔드포인트
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// API 엔드포인트 예시
	http.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "안녕하세요! API가 정상적으로 작동합니다.", "status": "success"}`)
	})

	log.Printf("서버가 포트 %s에서 시작됩니다...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
