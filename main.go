package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "net/http/pprof"

	"github.com/Timahawk/go_watcher"
	"github.com/Timahawk/mlsch_de/pkg/chat"
	"github.com/Timahawk/mlsch_de/pkg/locator"
	"github.com/Timahawk/mlsch_de/pkg/locator_v2"
	"github.com/Timahawk/mlsch_de/pkg/util"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
)

var development *bool

func main() {

	development = flag.Bool("dev", true, "Run local")
	flag.Parse()

	// This must be set before router is created.
	if *development == false {
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
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	} else {
		fmt.Println("Starting in Release Mode!")
		log.Fatalln(autotls.RunWithManager(r, &certManager))
	}
}

// SetupRouter does all the Routes setting.
// Extra function for easier testsetup.
func SetupRouter() *gin.Engine {

	util.InitLogger()
	util.Sugar.Infof("Started mlsch_de application")

	r := gin.Default()
	//r := gin.New()
	// Not using extra timestamp.
	// r.Use(ginzap.Ginzap(Logger, "", true))
	// r.Use(ginzap.RecoveryWithZap(Logger, true))

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
		watcher.GET("/echo", gin.WrapF(go_watcher.SendUpdates))
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
	locator.NewGame("world", "data/cities/worldcities.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90})
	locator.NewGame("large", "data/cities/capital_cities.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90})
	locator.NewGame("capitals", "data/cities/large_cities.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90})

	// Country specific games
	locator.NewGame("germany", "data/cities/german_cities.json", []float64{10.019531, 50.792047}, 1, 14, 1, []float64{-2.55, 42.18, 22.58, 58.86})

	{
		locators.GET("/", func(c *gin.Context) {
			c.HTML(200, "locators/start.html", gin.H{"title": "Locator"})
		})
		locators.GET("/:country", locator.HandleGame)
		locators.POST("/:country/submit", locator.HandleGameSubmit)
		locators.POST("/:country/newGuess", locator.HandleNewGuess)
	}

	// *************************************************************** //
	// 							LOCATOR-IO							   //
	// *************************************************************** //

	// locator_io.LoadedGames["world"], _ = locator_io.NewGame("world", "data/cities/worldcities.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90})
	// locator_io.LoadedGames["large"], _ = locator_io.NewGame("large", "data/cities/large_cities.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90})
	// locator_io.LoadedGames["capitals"], _ = locator_io.NewGame("capitals", "data/cities/capital_cities.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90})

	// // util.Sugar.Infow("Loaded Games",
	// //	"Games", locator_io.LoadedGames)

	// locator_ioGroup := r.Group("/l")

	// {
	// 	locator_ioGroup.GET("/", locator_io.CreateLobbyGET)
	// 	locator_ioGroup.POST("/", locator_io.CreateLobbyPOST)
	// 	locator_ioGroup.GET("/:lobby", locator_io.GetWaitingroom)
	// 	locator_ioGroup.GET("/:lobby/ws", locator_io.Waitingroom_WS)
	// 	locator_ioGroup.GET("/:lobby/game", locator_io.PlayGame)
	// 	locator_ioGroup.GET("/:lobby/game/ws", locator_io.ServeLobby)
	// }

	// *************************************************************** //
	// 							LOCATOR-V2							   //
	// *************************************************************** //

	// https://boundingbox.klokantech.com/

	locator_v2.LoadedGames["world"], _ = locator_v2.NewGame("world", "data/cities/worldcities.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90})
	locator_v2.LoadedGames["cities_larger_250000"], _ = locator_v2.NewGame("cities_larger_250000", "data/cities/cities_larger_250000.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90})
	locator_v2.LoadedGames["capitals"], _ = locator_v2.NewGame("capitals", "data/cities/capital_cities.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90})
	// Germany
	locator_v2.LoadedGames["germany"], _ = locator_v2.NewGame("germany", "data/cities/german_cities.json", []float64{10.019531, 50.792047}, 1, 14, 1, []float64{-2.55, 42.18, 22.58, 58.86})
	locator_v2.LoadedGames["germany_larger25000"], _ = locator_v2.NewGame("germany_larger25000", "data/cities/german_cities_larger25000.json", []float64{10.019531, 50.792047}, 1, 14, 1, []float64{-2.55, 42.18, 22.58, 58.86})
	// Japan
	locator_v2.LoadedGames["japan_larger25000"], _ = locator_v2.NewGame("japan_larger25000", "data/cities/japan_cities_larger25000.json", []float64{138.3, 34.76}, 1, 14, 1, []float64{118.44, 20.8, 155.53, 52.0})
	// Region specific games
	locator_v2.LoadedGames["european_cities_larger_100000"], _ = locator_v2.NewGame("european_cities_larger_100000", "data/cities/european_cities_larger_100000.json", []float64{10.019531, 50.792047}, 1, 14, 1, []float64{-41.8, 27.0, 69.6, 73.7})
	locator_v2.LoadedGames["north_american_cities_larger_100000"], _ = locator_v2.NewGame("north_american_cities_larger_100000", "data/cities/north_american_cities_larger_100000.json", []float64{-100, 40}, 1, 14, 1, []float64{-180, -15, 40, 85})

	if *development == true {
		util.Sugar.Infow("Staring Testing Lobby AAAAAAAA",
			"development", *development)
		locator_v2.SetupTest()
	}
	locator_v2Group := r.Group("/locate")
	{
		locator_v2Group.GET("/", locator_v2.CreateOrJoinLobby)
		locator_v2Group.POST("/create", locator_v2.CreateLobbyPOST)
		locator_v2Group.POST("/join", locator_v2.JoinLobbyPOST)
		locator_v2Group.GET("/:lobby", locator_v2.WaitingRoom)
		locator_v2Group.GET("/:lobby/ws", locator_v2.WaitingRoomWS)
		locator_v2Group.GET("/:lobby/game", locator_v2.GameRoom)
		locator_v2Group.GET("/:lobby/game/ws", locator_v2.GameRoomWS)
	}

	return r
}
