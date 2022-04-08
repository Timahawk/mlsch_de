package main

//
// import (
// 	"net/http"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// )

// var r *gin.Engine

// func init() {
// 	r = SetupRouter()
// 	go r.Run()
// }

// func wrapper(w http.ResponseWriter, r *http.Request) {
// 	x := &gin.Context{Writer: w, Request: r}
// 	ginFunction(x)
// }

// func ginFunction(c *gin.Context) {
// 	c.JSON(200, nil)
// }

// func Test_Frontpage(t *testing.T) {

// 	// assert.HTTPStatusCode(t, go_watcher.SendTemplate, "GET", "localhost:8080", nil, 200)
// 	// assert.HTTPSuccess(t, go_watcher.SendUpdates, "GET", "localhost:8080", nil, "Frontpage failed")
// 	// assert.HTTPSuccess(t, gin.WrapF(handler), "GET", "localhost:8080", nil, "Frontpage failed")
// 	// assert.HTTPSuccess(t, GetChatRoom, "GET", "localhost:8080", nil, "Frontpage failed")

// 	assert.HTTPStatusCode(t, wrapper, "GET", "", nil, 200)
// 	assert.HTTPStatusCode(t, wrapper, "GET", "", nil, 200)

// }
