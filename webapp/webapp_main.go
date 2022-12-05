//go:build webapp

package webapp

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gilwo/Sh0r7/common"
	"github.com/gilwo/Sh0r7/shortener"
	"github.com/gilwo/Sh0r7/store"
	_ "github.com/gilwo/Sh0r7/webapp/backend"
	webappCommon "github.com/gilwo/Sh0r7/webapp/common"
	_ "github.com/gilwo/Sh0r7/webapp/frontend"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/maxence-charriere/go-app/v9/pkg/errors"
)

var (
	ShortLiveTokenExpiration = 20 * time.Minute
)

func init() {
	fmt.Println("webapp init invoked")
	common.WebappInit = webappInit
	common.WebappGenFunc = webappgenfunc
	if expireEnv, ok := os.LookupEnv("SH0R7_WEBAPP_TOKEN_EXPIRATION_SHORT_LIVE"); ok {
		if expire, err := time.ParseDuration(expireEnv); err != nil {
			log.Printf("failed to parse duration from env")
		} else {
			ShortLiveTokenExpiration = expire
		}
		log.Printf("expire duration loaded from env and set to %s\n", ShortLiveTokenExpiration)
	}
}

var (
	helloH *app.Handler = &app.Handler{
		Name:        "Sh0r7",
		Description: "Sh0r7 url and data shortener",
		Icon: app.Icon{
			// Default: "/web/sh0r7-website-favicon-color.png",
			Default: "/web/short-giraffe-0.jpg",
			Large:   "/web/sh0r7-logo-color-on-transparent-background.png",
		},
		Styles: []string{
			// "/web/sh0r7-main.css",
			// "https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css",
		},
		Title: "this is Sh0r7",
		RawHeaders: []string{
			`<head>
			<link href="https://fonts.googleapis.com/css2?family=Share+Tech+Mono&display=swap" rel="stylesheet"/>
			<link href="https://fonts.googleapis.com/css?family=Cairo&display=swap" rel="stylesheet" />
			`,
		},
	}

	webappServedPaths map[string]bool
)

func webappInit() {

	if webappCommon.WebappFront == nil {
		panic("cant find webapp front")
	}
	webappCommon.WebappFront()

	webappServedPaths = map[string]bool{
		webappCommon.ShortPath:                 true,
		"/wasm_exec.js":                        true,
		"/app.js":                              true,
		"/manifest.webmanifest":                true,
		"/app-worker.js":                       true,
		"/app.css":                             true,
		"/web/app.wasm":                        true,
		"/web/sh0r7-main.css":                  true,
		"/web/sh0r7-website-favicon-color.png": true,
		"/web/short-giraffe-0.jpg":             true,
		"/web/sh0r7-logo-color-on-transparent-background.png": true,
	}

}

func webappgenfunc(args ...interface{}) interface{} {
	if len(args) < 2 {
		fmt.Printf("webappgenfunc: not enought args (%d)\n", len(args))
		return false
	}
	cmd, ok := args[0].(string)
	if !ok {
		fmt.Printf("webappgenfunc: first arg is not a string")
		return false
	}
	switch cmd {
	case "SERVE":
		c, ok := args[1].(*gin.Context)
		if !ok {
			fmt.Printf("webappgenfunc: second arg is not a gin context")
			return false
		}
		// fmt.Printf("*********\n%#+v\n*********\n", c.Request)
		path := c.Request.URL.Path
		fmt.Printf("webappgenfunc: handling path: %s\n", path)
		if _, ok := webappServedPaths[path]; !ok {
			return false
		}
		if path == webappCommon.ShortPath {
			headerUpdate(c)
		}
		helloH.ServeHTTP(c.Writer, c.Request)
		return true
	}
	return false
}

func headerUpdate(c *gin.Context) {
	if c.Request.Header.Get("RTS") == "" {
		return
	}
	fmt.Println("===========================================")
	fmt.Printf("request : %s\n", c.Request.Header)
	fmt.Printf("request UA: %s\n", c.Request.Header.Get("User-Agent"))
	fmt.Printf("clientIP : %s\n", c.ClientIP())
	fmt.Printf("client request uri : %s\n", c.Request.RequestURI)
	fmt.Printf("client request url : %s\n", c.Request.URL)
	// seed := generateSeedAndStoreToken(fmt.Sprintf("%s", c.Request.Header))
	seedLen, tokenLen := 32, 40
	seed, token := generateSeedAndToken(c.Request.Header.Get("User-Agent"), seedLen, tokenLen)
	fmt.Println("/*/*/*/*/*/*/*/*/*/*/*/*/")
	fmt.Printf("seed: <%s> (%d)\n", seed, seedLen)
	fmt.Printf("token: <%s> (%d)\n", token, tokenLen)
	fmt.Println("/*/*/*/*/*/*/*/*/*/*/*/*/")

	fmt.Printf("generated seed : <%s>\n", seed)
	c.Writer.Header().Add("stid", seed)                        // seed
	c.Writer.Header().Add("stid", fmt.Sprintf("%d", tokenLen)) // tokenLen
	c.Writer.Header().Add("stid", "0")                         // token start pos
	fmt.Println("===========================================")
}
func generateSeedAndToken(input string, seedLen, tokenLen int) (string, string) {
	var seed, token string
	c := 0
	for {
		seed = shortener.GenerateTokenTweaked(uuid.NewString(), 0, seedLen, 0)
		token = shortener.GenerateTokenTweaked(input+seed, 0, tokenLen, 0)
		err := store.StoreCtx.SaveDataMapping([]byte(""), token, ShortLiveTokenExpiration)
		if err == nil {
			break
		}
		c += 1
		if c > 200 {
			panic(errors.Newf("too many retries (%d) on token saving", c))
		}
	}
	return seed, token
}
