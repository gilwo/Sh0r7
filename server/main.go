package server

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
	n, e := strconv.ParseInt(common.BuildTime, 10, 64)
	t := time.Unix(n, 0).UTC()
	log.Printf("BuildVersion: <%s> time: <%v> (%v)\n", common.BuildVersion, t, e)
	if strings.ToLower(os.Getenv("SH0R7_PRODUCTION")) == "true" {
		gin.SetMode(gin.ReleaseMode)
	}
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
	log.Printf("env Local<%s>, Redis<%s>, fallback<%s>\n", envLocal, envRedis, envFallback)
	if redisUrl := envRedis; redisUrl != "" && !*useLocal {
		fmt.Printf("redisURL: %s\n", redisUrl)
		if store.NewStoreRedis == nil {
			return errors.New("missing redis storage support")
		}
		store.StoreCtx = store.NewStoreRedis(redisUrl)
		log.Printf("redis storage initialized")
	} else if envLocal != "" {
		if store.NewStoreLocal == nil {
			return errors.New("missing local storage support")

		}
		store.StoreCtx = store.NewStoreLocal()
		log.Printf("local storage initialized")
	}
	if store.StoreCtx == nil {
		if strings.TrimSpace(strings.ToLower(envFallback)) == "true" || *useLocal {
			fmt.Println("no specific store defined - fallback to local storage")
			store.StoreCtx = store.NewStoreLocal()
			log.Printf("local fallback storage initialized")
		} else {
			return errors.New("missing storage support")
		}
	}
	err := store.StoreCtx.InitializeStore()
	if err != nil {
		return errors.Wrap(err, "store init failed, exiting...\n")
	}
	handler.StoreFavicon()
	log.Printf("storage sucessfully initialized")
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
	log.Printf("admin key: [%s], token: [%s]\n", adminKey, adTok)
}

func startServer() {
	addr := ":8080"
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
		log.Printf("original: %+#v\n", c.Request.URL)
		log.Printf("generic handler: short <%s>, ext <%s>\n", paramShort, paramExt)

		// if paramShort == "hc" {
		// 	if time.Now().Unix()%60 == 0 {
		// 		defer QueueMaintWork()
		// 	}
		// 	if paramExt == "dump" {
		// 		handler.HandleDumpMaint(c, DumpMaintList)
		// 		return
		// 	}
		// 	c.Status(200)
		// } else
		if paramShort == "favicon.ico" {
			c.FileFromFS(".", handler.HandleGetFavIcon())
		} else if paramExt == "info" {
			handler.HandleGetShortDataInfo(c)
		} else if paramExt == "data" {
			handler.HandleGetOriginData(c)
		} else if paramShort == "dump" && paramExt == "keys" {
			handler.HandleDumpKeys(c)
		} else if common.WebappGenFunc != nil && common.WebappGenFunc("SERVE", c).(bool) {
			return
			// } else if store.StoreCtx.CheckExistShortDataMapping(c.Request.URL.Path) {
			// 	data, err := store.StoreCtx.LoadDataMapping(c.Request.URL.Path)
			// 	if err == nil {
			// 		c.Data(200, "image/jpg", data)
			// 		return
			// 	}
			// 	c.JSON(http.StatusInternalServerError, "error")
		} else {
			handler.HandleShort(c)
		}
	}
	r.GET("/", genericHandleShort)
	r.GET("/:short", genericHandleShort)
	r.GET("/:short/*ext", genericHandleShort)

	handleHealthCheck := func(c *gin.Context) {
		log.Println("handleHC")
		t := time.Now().Unix()
		t10 := t % 10
		log.Printf("time %%10: %v (%v)\n", t, t10)
		if time.Now().Unix()%10 >= 5 {
			defer QueueMaintWork()
		}
		if strings.Trim(c.Param("hcparam"), "/") == "dump" {
			handler.HandleDumpMaint(c, DumpMaintList)
			return
		}
		c.Status(200)
	}
	r.GET("/hc", handleHealthCheck)
	r.GET("/hc/*hcparam", handleHealthCheck)

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
	registerTemplates(r)
	return r
}

func registerTemplates(r *gin.Engine) {
	templateShowPublic, err := template.New("public-show-no-lock").Parse(`<!doctype html>
<html lang="en"><head>
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=0, viewport-fit=cover">
<link type="text/css" rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.4.1/css/bootstrap.min.css">
</head><body>
<div><pre contenteditable="false">{{.Data}}</pre></div>
</body></html>`)
	if err != nil {
		panic(err)
	}
	r.SetHTMLTemplate(templateShowPublic)
}

var (
	OpenTelemetryServiceName string
)

func initUptrace() {
	// isLive := true
	// serviceName := os.Getenv("RENDER_SERVICE_NAME")
	// serviceID := os.Getenv("RENDER_SERVICE_ID")
	// instanceID := os.Getenv("RENDER_INSTANCE_ID")
	// serviceVersion := "dev"
	// the system service name that will be shown in uptrace will be in the following form:
	// [<service name>|<development host>].<deploy type>.sh0r7.me[.debug|.test]
	// where
	// service name - relevant for deployed instance
	// development host - relevant for local development running instance
	// .debug / .test - relevant for gin mode
	// deploy type - must be defined
	prefix := ""
	version := common.BuildVersion
	deploy := ""
	switch deploy = os.Getenv("SH0R7_DEPLOY"); deploy {
	case "production", "development":
	case "staging", "testing":
		panic("deploy type not ready yet: " + deploy)
	case "localdev":
		if _, ok := os.LookupEnv("RENDER"); ok {
			panic("invalid deploy type " + deploy + " for render environment")
		}
		version += ":build@" + common.BuildTime
	default:
		panic("deploy type not familiar: [" + deploy + "]")
	}
	if _, ok := os.LookupEnv("RENDER"); ok {
		prefix = os.Getenv("RENDER_SERVICE_NAME")
		if len(version) == 0 {
			version = os.Getenv("RENDER_GIT_COMMIT")
		}
	} else if _, ok := os.LookupEnv("SH0R7__DEV_ENV"); ok {
		log.Default().SetFlags(log.Flags() | log.Llongfile)
		common.IsDevEnv = true
		prefix = os.Getenv("SH0R7_DEV_HOST")
		version += ":source@" + common.SourceTime
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
			uptrace.WithDeploymentEnvironment(deploy),
		)
		options = append(options,
			uptrace.WithServiceVersion(version),
		)
		uptrace.ConfigureOpentelemetry(options...)
		log.Printf("** using uptrace OTEL service using [%s]\n", OpenTelemetryServiceName)
	}
	metrics.GlobalMeter = metrics.InitGlobalMeter(OpenTelemetryServiceName)
	// fmt.Printf("host: %+#v\n", os.Getenv("SH0R7_DEV_HOST"))
	// fmt.Printf("allenv: %+#v\n", os.Environ())

	// metrics.GlobalMeter.IncMeterCounter(metrics.CreationFailure)
	// metrics.GlobalMeter.IncMeterCounter(metrics.VisitPublic)
	// metrics.GlobalMeter.IncMeterCounter(metrics.VisitPublic)
	// metrics.GlobalMeter.IncMeterCounter(metrics.VisitPublic)
	// fmt.Printf("global:\n%s\n", metrics.GlobalMeter.Dump())
	// <-time.After(60 * time.Second)

	// panic("123")
}
