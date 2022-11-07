package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gilwo/5H0r7/handler"
	"github.com/gilwo/5H0r7/store"
	"github.com/gin-gonic/gin"
)

var (
	useLocal = flag.Bool("local", false, "force use local storage")
)

func main() {
	flag.Parse()

	// Note store initialization happens here
	envLocal := os.Getenv("SH0R7_STORE_LOCAL")
	envRedis := os.Getenv("SH0R7_STORE_REDIS")
	envFallback := os.Getenv("SH0R7_STORE_FALLBACK")
	if redisUrl := envRedis; redisUrl != "" && !*useLocal {
		// fmt.Printf("redisURL: %s\n", redisUrl)
		if store.NewStoreRedis == nil {
			fmt.Println("missing redis storage support")
			os.Exit(1)
		}
		store.StoreCtx = store.NewStoreRedis(redisUrl)
	} else if envLocal != "" {
		if store.NewStoreLocal == nil {
			fmt.Println("missing local storage support")
			os.Exit(1)
		}
		store.StoreCtx = store.NewStoreLocal()
	}
	if store.StoreCtx == nil {
		if strings.TrimSpace(strings.ToLower(envFallback)) == "true" || *useLocal {
			fmt.Println("no specific store defined - fallback to local storage")
			store.StoreCtx = store.NewStoreLocal()
		} else {
			fmt.Println("missing storage support")
			os.Exit(1)
		}
	}
	err := store.StoreCtx.InitializeStore()
	if err != nil {
		fmt.Printf("store init failed, exiting...: %s\n", err)
		panic(err)
	}
	handler.StoreFavicon()

	addr := ":9808"
	envProd := os.Getenv("SH0R7_PRODUCTION")
	if strings.ToLower(envProd) == "true" {
		gin.SetMode(gin.ReleaseMode)
	}
	addr = ":8080"
	if envAddr := os.Getenv("SH0R7_ADDR"); envAddr != "" {
		addr = envAddr
	}

	err = GinInit().Run(addr)
	if err != nil {
		panic(fmt.Sprintf("Failed to start the web server - Error: %v", err))
	}

}

func GinInit() *gin.Engine {

	// gin endpoints
	// ----------------
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the URL Shortener API",
		})
	})

	r.GET("/:short", func(c *gin.Context) {
		if c.Param("short") == "favicon.ico" {
			c.FileFromFS(".", handler.HandleGetFavIcon())
			return
		}
	})

	// create short for data
	r.POST("/create-short-data", func(c *gin.Context) {
		handler.HandleCreateShortDataImproved1(c)
	})

	// retrieve meta data on short
	r.GET("/:short/info", func(c *gin.Context) {
		// TODO: get more info... not only the long url ...
		handler.HandleGetShortDataInfo(c)
	})

	// retrieve original data on store on short
	r.GET("/:short/data", func(c *gin.Context) {
		handler.HandleGetOriginData(c)
	})

	// update data on short
	r.PATCH("/:short", func(c *gin.Context) {
		handler.UpdateShortData(c)
	})
	// update data on short
	r.DELETE("/:short", func(c *gin.Context) {
		handler.DeleteShortData(c)
	})
	return r
}