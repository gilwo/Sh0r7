package server

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/uptrace/uptrace-go/uptrace"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/gilwo/Sh0r7/common"
	"github.com/gilwo/Sh0r7/handler"
	"github.com/gilwo/Sh0r7/metrics"
	"github.com/gilwo/Sh0r7/shortener"
	"github.com/gilwo/Sh0r7/store"
	_ "github.com/gilwo/Sh0r7/webapp"
)

func init() {
	common.MainServer = mainServer
	initUptrace()
}

var (
	useLocal  = flag.Bool("local", false, "force use local storage")
	useWebapp = flag.Bool("webapp", false, "include webapp")

	mainCtx context.Context
)

func mainServer() {
	flag.Parse()
	err := webappInit()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = storageInit()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// metrics.MetricGlobalCounter = metrics.NewMetricGlobal()
	// metrics.MetricProcessor = metrics.NewMetricContext()
	// metrics.MetricProcessor.StartProcessing()
	// metrics.MetricProcessor.EnableDisableDump()
	adTokenSet()
	startServer()
}

func webappInit() error {
	envWebapp := os.Getenv("SH0R7_WEBAPP")
	if envWebapp != "" || *useWebapp {
		if common.WebappInit == nil {
			return errors.New("webapp code not included in this build")
		}
		common.WebappInit()
	}
	return nil
}

func storageInit() error {
	envLocal := os.Getenv("SH0R7_STORE_LOCAL")
	envRedis := os.Getenv("SH0R7_STORE_REDIS")
	envFallback := os.Getenv("SH0R7_STORE_FALLBACK")
	if redisUrl := envRedis; redisUrl != "" && !*useLocal {
		fmt.Printf("redisURL: %s\n", redisUrl)
		if store.NewStoreRedis == nil {
			return errors.New("missing redis storage support")
		}
		store.StoreCtx = store.NewStoreRedis(redisUrl)
	} else if envLocal != "" {
		if store.NewStoreLocal == nil {
			return errors.New("missing local storage support")

		}
		store.StoreCtx = store.NewStoreLocal()
	}
	if store.StoreCtx == nil {
		if strings.TrimSpace(strings.ToLower(envFallback)) == "true" || *useLocal {
			fmt.Println("no specific store defined - fallback to local storage")
			store.StoreCtx = store.NewStoreLocal()
		} else {
			return errors.New("missing storage support")
		}
	}
	err := store.StoreCtx.InitializeStore()
	if err != nil {
		return errors.Wrap(err, "store init failed, exiting...\n")
	}
	handler.StoreFavicon()
	return nil
}

func adTokenSet() {
	var err error
	adminKey := os.Getenv("SH0R7_ADMIN_KEY")
	if adminKey == "" {
		adminKey = uuid.NewString()
	}
	adTok := shortener.GenerateTokenTweaked(adminKey, 0, 32, 0)
	if !store.StoreCtx.CheckExistShortDataMapping(adTok) {
		err = store.StoreCtx.SaveDataMapping([]byte(""), adTok, -1)
		if err != nil {
			panic(err)
		}
	}
	err = store.StoreCtx.SetMetaDataMapping(adTok, store.FieldBlocked, store.IsBLOCKED)
	if err != nil {
		panic(err)
	}
}

func startServer() {
	addr := ":9808"
	envProd := os.Getenv("SH0R7_PRODUCTION")
	if strings.ToLower(envProd) == "true" {
		gin.SetMode(gin.ReleaseMode)
	}
	addr = ":8080"
	if envAddr := os.Getenv("SH0R7_ADDR"); envAddr != "" {
		addr = envAddr
	}

	ctx, cancel := ContextWithSignals(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	defer exit()
	mainCtx = ctx

	srv := &http.Server{
		Addr:    addr,
		Handler: GinInit(),
	}
	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()
	go func() {
		fmt.Println("** server starting **")
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				fmt.Printf("failed serving with gin: %s\n", err)
				os.Exit(1)
			}
		}
	}()
	<-ctx.Done()
	time.Sleep(100 * time.Millisecond)
	// metrics.MetricProcessor.StopProcessing()
	fmt.Println("** server down **")
	// err = GinInit().Run(addr)
	// if err != nil {
	// 	panic(fmt.Sprintf("Failed to start the web server - Error: %v", err))
	// }

}

func ContextWithSignals(parent context.Context, sig ...os.Signal) (ctx context.Context, cancel func()) {
	ctx, cancel = context.WithCancel(parent)
	c := make(chan os.Signal, 1)
	signal.Notify(c, sig...)

	go func() {
		<-c
		close(c)
		cancel()
	}()

	return ctx, cancel
}

func exit() {
	err := recover()
	if err != nil {
		fmt.Printf("command failed: %s", err)
		os.Exit(-1)
	}
}

func GinInit() *gin.Engine {

	// gin endpoints
	// ----------------
	r := gin.Default()
	if OpenTelemetryServiceName != "" {
		r.Use(otelgin.Middleware(OpenTelemetryServiceName))
	}

	// r.GET("/", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{
	// 		"message": "Welcome to the URL Shortener API",
	// 	})
	// })

	r.MaxMultipartMemory = 1 << 20 // 1 MiB

	genericHandleShort := func(c *gin.Context) {
		paramShort := c.Param("short")
		paramExt := strings.Trim(c.Param("ext"), "/")
		log.Printf("generic handler: short <%s>, ext <%s>\n", paramShort, paramExt)

		// if common.WebappGenFunc != nil {
		// 	res := common.WebappGenFunc("SERVE", c)
		// 	if res.(bool) {
		// 		return
		// 	}
		// }

		if paramShort == "hc" {
			defer QueueMaintWork()
			c.String(200, "")
		} else if paramShort == "favicon.ico" {
			c.FileFromFS(".", handler.HandleGetFavIcon())
		} else if paramExt == "info" {
			handler.HandleGetShortDataInfo(c)
		} else if paramExt == "data" {
			handler.HandleGetOriginData(c)
		} else if gin.Mode() == gin.DebugMode && paramShort == "dump" && paramExt == "keys" {
			handler.HandleDumpKeys(c)
		} else if common.WebappGenFunc != nil && common.WebappGenFunc("SERVE", c).(bool) {
			return
		} else {
			handler.HandleShort(c)
		}
	}
	r.GET("/", genericHandleShort)
	r.GET("/:short", genericHandleShort)
	r.GET("/:short/*ext", genericHandleShort)

	// create short for url
	r.POST("/create-short-url", func(c *gin.Context) {
		handler.HandleCreateShortUrl(c)
	})

	r.POST("/admin/upload", func(c *gin.Context) {
		handler.HandleUploadFile(c)
	})
	// create short for data
	r.POST("/create-short-data", func(c *gin.Context) {
		handler.HandleCreateShortData(c)
	})

	// // retrieve meta data on short
	// r.GET("/:short/info", func(c *gin.Context) {
	// 	// TODO: get more info... not only the long url ...
	// 	handler.HandleGetShortDataInfo(c)
	// })

	// // retrieve original data on store on short
	// r.GET("/:short/data", func(c *gin.Context) {
	// 	handler.HandleGetOriginData(c)
	// })

	// update data on short
	r.PATCH("/:short", func(c *gin.Context) {
		handler.HandleUpdateShort(c)
	})
	// update data on short
	r.DELETE("/:short", func(c *gin.Context) {
		handler.RemoveShortData(c)
	})
	return r
}

var (
	OpenTelemetryServiceName string
)

func initUptrace() {
	// the system service name that will be shown in uptrace will be in the following form:
	// [<service name>|<development host>].<deploy type>.sh0r7.me[.debug|.test]
	// where
	// service name - relevant for deployed instance
	// development host - relevant for local development running instance
	// .debug / .test - relevant for gin mode
	// deploy type - must be defined
	prefix := ""
	version := ""
	deploy := ""
	switch deploy = os.Getenv("SH0R7_DEPLOY"); deploy {
	case "prod", "dev":
	case "stage", "test":
		panic("deploy type not ready yet: " + deploy)
	case "localdev":
		if _, ok := os.LookupEnv("RENDER"); ok {
			panic("invalid deploy type " + deploy + " for render environment")
		}
		version = deploy // TODO - add build time ??
	default:
		panic("deploy type not familiar: [" + deploy + "]")
	}
	if _, ok := os.LookupEnv("RENDER"); ok {
		prefix = os.Getenv("RENDER_SERVICE_NAME")
		version = os.Getenv("RENDER_GIT_COMMIT")
		if deploy == "prod" {
			version = "v.0.0.0-alpha" // TODO : grab from tag
		}
	} else if _, ok := os.LookupEnv("SH0R7__DEV_ENV"); ok {
		prefix = os.Getenv("SH0R7_DEV_HOST")
		version = deploy
	}

	OpenTelemetryServiceName = prefix + "." + deploy + ".sh0r7.me"
	switch gin.Mode() {
	case gin.DebugMode:
		OpenTelemetryServiceName += ".debug"
	case gin.TestMode:
		OpenTelemetryServiceName += ".test"
	case gin.ReleaseMode:
	}
	if otelEnv := os.Getenv("SH0R7_OTEL_UPTRACE"); otelEnv != "" {
		options := []uptrace.Option{}
		options = append(options,
			uptrace.WithDSN(otelEnv),
			uptrace.WithServiceName(OpenTelemetryServiceName),
			uptrace.WithDeploymentEnvironment("development"),
		)
		options = append(options,
			uptrace.WithServiceVersion(version),
		)
		uptrace.ConfigureOpentelemetry(options...)
		log.Printf("** using uptrace OTEL service using [%s]\n", OpenTelemetryServiceName)
	}
	metrics.GlobalMeter = metrics.InitGlobalMeter(OpenTelemetryServiceName)
}
