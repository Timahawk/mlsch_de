package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Timahawk/go_watcher"
	"github.com/Timahawk/mlsch_de/pkg/chat"
	"github.com/Timahawk/mlsch_de/pkg/locator"
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

	r := SetupRouter()

	// *************************************************************** //
	// 							Certs								   //
	// *************************************************************** //

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

// SetupRouter does all the Routes setting.
// Extra function for easier testsetup.
func SetupRouter() *gin.Engine {

	r := gin.Default()

	// *************************************************************** //
	// 						Files & Templates 						   //
	// *************************************************************** //

	r.StaticFile("/favicon.ico", "web/static/favicon.ico")
	r.Static("/static", "web/static")
	r.LoadHTMLGlob("web/templates/**/*.html")

	// *************************************************************** //
	// 							Frontpage 							   //
	// *************************************************************** //

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "globals/index.html", nil)
	})

	// *************************************************************** //
	// 							GO WATCHER 							   //
	// *************************************************************** //

	watcher := r.Group("/watcher")
	go_watcher.Start(time.Second)
	{
		// watcher.GET("/echo", func(c *gin.Context) {
		// 	go_watcher.SendUpdates(c.Writer, c.Request)
		// })
		watcher.GET("/echo", gin.WrapF(go_watcher.SendUpdates))

		// watcher.GET("/", func(c *gin.Context) {
		// 	go_watcher.SendTemplate(c.Writer, c.Request)
		// })
		watcher.GET("/", gin.WrapF(go_watcher.SendTemplate))
	}

	// *************************************************************** //
	// 							CHATS 								   //
	// *************************************************************** //

	chats := r.Group("/chats")
	{
		chats.GET("/", func(c *gin.Context) {
			c.HTML(200, "chats/start.html", gin.H{"title": "Chats"})
		})
		chats.GET(":room/chat", chat.GetChatRoom)
		chats.GET(":room/ws", chat.GetRoomWebsocket)
		chats.POST("/", chat.PostCreateNewHub)
	}

	// *************************************************************** //
	// 							LOCATOR								   //
	// *************************************************************** //

	locators := r.Group("/locators")

	// World wide games
	locator.NewGame("world", "pkg/locator/worldcities.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90})
	locator.NewGame("large", "pkg/locator/capital_cities.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90})
	locator.NewGame("capitals", "pkg/locator/large_cities.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90})

	// Country specific games
	locator.NewGame("germany", "pkg/locator/german_cities.json", []float64{10.019531, 50.792047}, 1, 14, 1, []float64{-2.55, 42.18, 22.58, 58.86})

	{
		locators.GET("/", func(c *gin.Context) {
			c.HTML(200, "locators/start.html", gin.H{"title": "Locator"})
		})
		locators.GET("/:country", locator.HandleGame)
		locators.POST("/:country/submit", locator.HandleGameSubmit)
		locators.POST("/:country/newGuess", locator.HandleNewGuess)
	}
	return r
}
