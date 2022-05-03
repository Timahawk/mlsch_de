package main

import (
	"embed"
	"flag"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	_ "net/http/pprof"

	"github.com/Timahawk/go_watcher"
	"github.com/Timahawk/mlsch_de/pkg/chat"
	"github.com/Timahawk/mlsch_de/pkg/locator_v2"
	"github.com/Timahawk/mlsch_de/pkg/util"
	"go.uber.org/zap"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
)

var development *bool

//go:embed web/static/**/* web/static/*
var staticFS embed.FS

//go:embed web/templates/**/*
var templatesFS embed.FS

// the Loggger used throughout
var Logger *zap.Logger

func init() {
	development = flag.Bool("dev", true, "Run local")
	Logger = util.InitLogger()
}

type Server struct {
	*gin.Engine
}

func main() {

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
		go func() {
			util.Sugar.Info(http.ListenAndServe("localhost:6060", nil))
		}()
		util.Sugar.Fatal(r.Run(":8080"))

	} else {
		util.Sugar.Infof("Starting in Release Mode!")
		util.Sugar.Fatal(autotls.RunWithManager(r, &certManager))
	}
}

func mustFS() http.FileSystem {
	sub, err := fs.Sub(staticFS, "web/static")

	if err != nil {
		panic(err)
	}

	return http.FS(sub)
}

// SetupRouter does all the Routes setting.
// Extra function for easier testsetup.
func SetupRouter() *gin.Engine {

	util.Sugar.Infow("Started mlsch_de application")

	var r *gin.Engine

	if *development {
		r = gin.Default()
	} else {
		r = gin.New()
		// Not using extra timestamp.
		r.Use(ginzap.Ginzap(Logger, "", true))
		r.Use(ginzap.RecoveryWithZap(Logger, true))
	}

	//r := gin.New()
	// Not using extra timestamp.
	// r.Use(ginzap.Ginzap(Logger, "", true))
	// r.Use(ginzap.RecoveryWithZap(Logger, true))

	// *************************************************************** //
	// 						Files & Templates 						   //
	// *************************************************************** //

	// This is so that in dev Mode you can reload templates for better dev experience.
	if *development == true {
		util.Sugar.Infow("Loading templates from external FileSystem")
		r.LoadHTMLGlob("web/templates/**/*.html")
		r.Static("/static", "web/static")
	} else {
		util.Sugar.Infow("Loading templates from internal (embedded) FileSystem")
		templ := template.Must(template.New("").ParseFS(templatesFS, "web/templates/**/*.html"))
		r.SetHTMLTemplate(templ)
		r.StaticFS("/static", mustFS())
	}

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
	// 							LOCATOR-V2							   //
	// *************************************************************** //

	if err := locator_v2.LoadGames(); err != nil {
		util.Sugar.Fatalf("Fatal loading games %v", err)
	}

	if *development == true {
		util.Sugar.Infow("Creating Testing Lobby AAAAAAAA",
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
