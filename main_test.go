package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_CreateOrJoinLobby(t *testing.T) {

	// For less crazy printing.
	gin.SetMode(gin.ReleaseMode)
	router := SetupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/locate/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	// assert.Equal(t, "pong", w.Body.String())
}

func Test_CreateLobbyPost(t *testing.T) {

	// For less crazy printing.
	gin.SetMode(gin.ReleaseMode)
	router := SetupRouter()
	w := httptest.NewRecorder()

	form := url.Values{}
	form.Add("roundTime", "15")
	form.Add("reviewTime", "15")
	form.Add("rounds", "15")
	form.Add("gameset", "capitals")
	// form.Add("username", "test")

	// Should not work due to Missing Input
	req, _ := http.NewRequest("POST", "/locate/create", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, w.Body.String())
	assert.Equal(t, `{"status":"CreateLobbyPost failed, due to faulty Form Input."}`, w.Body.String())

	// Should now work due to Complete Input.
	form.Add("username", "test")

	req, _ = http.NewRequest("POST", "/locate/create", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
