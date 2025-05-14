package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	requests := []struct {
		name  string
		count int
		want  int
	}{
		{
			name:  "count=0",
			count: 0,
			want:  0,
		},
		{
			name:  "count=1",
			count: 1,
			want:  1,
		},
		{
			name:  "count=2",
			count: 2,
			want:  2,
		},
		{
			name:  "count=100",
			count: 100,
			want:  len(cafeList["moscow"]),
		},
	}

	for _, req := range requests {
		t.Run(req.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/cafe?city=moscow&count="+string(req.count), nil)
			responseRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(mainHandle)
			handler.ServeHTTP(responseRecorder, request)

			require.Equal(t, http.StatusOK, responseRecorder.Code)
			
			body := strings.TrimSpace(responseRecorder.Body.String())
			var cafes []string
			if body != "" {
				cafes = strings.Split(body, ",")
			}
			
			assert.Equal(t, req.want, len(cafes))
		})
	}
}

func TestCafeSearch(t *testing.T) {
	requests := []struct {
		name      string
		search    string
		wantCount int
	}{
		{
			name:      "search=фасоль",
			search:    "фасоль",
			wantCount: 0,
		},
		{
			name:      "search=кофе",
			search:    "кофе",
			wantCount: 2,
		},
		{
			name:      "search=вилка",
			search:    "вилка",
			wantCount: 1,
		},
	}

	for _, req := range requests {
		t.Run(req.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/cafe?city=moscow&search="+req.search, nil)
			responseRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(mainHandle)
			handler.ServeHTTP(responseRecorder, request)

			require.Equal(t, http.StatusOK, responseRecorder.Code)
			
			body := strings.TrimSpace(responseRecorder.Body.String())
			var cafes []string
			if body != "" {
				cafes = strings.Split(body, ",")
			}
			
			assert.Equal(t, req.wantCount, len(cafes))

			searchLower := strings.ToLower(req.search)
			for _, cafe := range cafes {
				cafeLower := strings.ToLower(strings.TrimSpace(cafe))
				assert.True(t, strings.Contains(cafeLower, searchLower),
					"cafe '%s' should contain '%s'", cafe, req.search)
			}
		})
	}
}