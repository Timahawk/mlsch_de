package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Timahawk/go_watcher"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	development := flag.Bool("dev", true, "Run local")
	flag.Parse()

	// This must be set before router is created.
	if !*development {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	watcher := r.Group("/watcher")
	go_watcher.Start(time.Second)
	{
		watcher.GET("/echo", func(c *gin.Context) {
			go_watcher.SendUpdates(c.Writer, c.Request)
		})
		watcher.GET("/", func(c *gin.Context) {
			go_watcher.SendTemplate(c.Writer, c.Request)
		})
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"Status": "Worked"})
	})

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("mlsch.de"), //Your domain here
		Cache:      autocert.DirCache("certs"),         //Folder for storing certificates
	}

	// Check if development (default) or Prod.
	if *development {
		log.Fatalln(r.Run())
	} else {
		fmt.Println("Starting in Release Mode!")
		log.Fatalln(autotls.RunWithManager(r, &certManager))
	}
}
