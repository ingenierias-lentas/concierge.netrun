package server

import (
	//"fmt"
	//"github.com/julienschmidt/httprouter"
	//"github.com/justinas/alice"
	//"github.com/justinas/nosurf"
	"database/sql"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/throttled/throttled"
	"github.com/throttled/throttled/store/memstore"
	"go.uber.org/zap"
	"net/http"
	"time"
)

//TODO should store secrets in a vault
var jwtSecret []byte

var db *sql.DB

var Logger *zap.Logger

func SetJwtSecret(secret []byte) {
	jwtSecret = secret
}

func GetJwtSecret() []byte {
	return jwtSecret
}

func SetDb(sqlDb *sql.DB) {
	db = sqlDb
}

func GetDb() *sql.DB {
	return db
}

func SetLogger(log *zap.Logger) {
	Logger = log
}

func handler(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

// Rate-limit to 20 requests per path per minute with max burst of 5 requests
func rateLimiter() throttled.HTTPRateLimiter {
	store, err := memstore.New(65536)
	if err != nil {
		Logger.Fatal(err.Error())
	}

	// Max burst of 5 which refills at 20 tokens per minute
	quota := throttled.RateQuota{MaxRate: throttled.PerMin(20), MaxBurst: 5}

	rateLimiter, err := throttled.NewGCRARateLimiter(store, quota)
	if err != nil {
		Logger.Fatal(err.Error())
	}

	httpRateLimiter := throttled.HTTPRateLimiter{
		RateLimiter: rateLimiter,
		VaryBy:      &throttled.VaryBy{Path: true},
	}

	return httpRateLimiter
}

func InitServer() *gin.Engine {
	router := gin.Default()

	router.Use(ginzap.Ginzap(Logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(Logger, true))

	errcsoolCors := cors.New(cors.Config{
		AllowOrigins: []string{
			"https://errcsool.com",
			"http://192.168.1.37:8020",
			"http://localhost:8020",
		},
		AllowMethods:     []string{"POST", "GET", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})

	accessRouter := router.Group("/access")
	accessRouter.Use(errcsoolCors)
	accessRouter.POST("/signin", Signin)
	accessRouter.POST("/signup", Signup)
	accessRouter.POST("/signout", Signout)

	commandRouter := router.Group("/command")
	commandRouter.Use(errcsoolCors)
	commandRouter.POST("/newcommand", VerifyToken(), CheckGroup(), IsAdmin(), NewCommand)
	commandRouter.POST("/deletecommand", VerifyToken(), CheckGroup(), CanWrite(), DeleteCommand)
	commandRouter.POST("/runcommand", VerifyToken(), CheckGroup(), CanExecute(), RunCommand)
	commandRouter.POST("/killcommand", VerifyToken(), CheckGroup(), CanExecute(), KillCommand)

	router.GET("/ping", handler)

	return router
}

func RunServer(portString string, router *gin.Engine) {
	throttler := rateLimiter()
	http.ListenAndServe(portString, throttler.RateLimit(router))
}
