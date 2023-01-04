//go:build webapp

package webapp

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gilwo/Sh0r7/common"
	"github.com/gilwo/Sh0r7/shortener"
	"github.com/gilwo/Sh0r7/store"
	_ "github.com/gilwo/Sh0r7/webapp/backend"
	webappCommon "github.com/gilwo/Sh0r7/webapp/common"
	"github.com/gilwo/Sh0r7/webapp/frontend"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/maxence-charriere/go-app/v9/pkg/errors"
)

var (
	ShortLiveTokenExpiration = 20 * time.Minute
)

func init() {
	log.Println("webapp init invoked")
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
	sh0r7H *app.Handler = &app.Handler{
		Name:        "Sh0r7",
		Description: "Sh0r7 url and data shortener",
		Icon: app.Icon{
			// Default: "/web/sh0r7-website-favicon-color.png",
			Default: "logoS.png",
			Large:   "logoL.png",
		},
		LoadingLabel: "standby",
		Styles: []string{
			// "/web/sh0r7-main.css",
			"/web/main.css",
			"https://maxcdn.bootstrapcdn.com/bootstrap/3.4.1/css/bootstrap.min.css",
		},
		Title: "this is Sh0r7",
		RawHeaders: []string{
			`
			<link href="https://fonts.googleapis.com/css2?family=Share+Tech+Mono&display=swap" rel="stylesheet"/>
			<link href="https://fonts.googleapis.com/css?family=Cairo&display=swap" rel="stylesheet" />
			<link href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.10.2/font/bootstrap-icons.css"rel="stylesheet"  >
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
		webappCommon.ShortPath:   true,
		webappCommon.PrivatePath: true,
		webappCommon.PublicPath:  true,
		"/wasm_exec.js":          true,
		"/app.js":                true,
		"/manifest.webmanifest":  true,
		"/app-worker.js":         true,
		"/app.css":               true,
		"/web/app.wasm":          true,
		"/web/main.css":          true,
	}

	if gin.Mode() == gin.DebugMode {
		sh0r7H.Icon = app.Icon{
			Default: frontend.ImgSource,
			Large:   frontend.ImgSource,
		}
		webappServedPaths[frontend.ImgSource] = true
	}
}

func webappgenfunc(args ...interface{}) interface{} {
	if len(args) < 2 {
		log.Printf("webappgenfunc: not enought args (%d)\n", len(args))
		return false
	}
	cmd, ok := args[0].(string)
	if !ok {
		log.Printf("webappgenfunc: first arg is not a string")
		return false
	}
	switch cmd {
	case "SERVE":
		c, ok := args[1].(*gin.Context)
		if !ok {
			log.Printf("webappgenfunc: second arg is not a gin context")
			return false
		}
		// log.Printf("*********\n%#+v\n*********\n", c.Request)
		path := c.Request.URL.Path
		log.Printf("webappgenfunc: handling path: %s\n", path)
		if _, ok := webappServedPaths[path]; !ok {
			if fixPath(c) {
				return true
			}
			if checkPublicRedirect(c) {
				return true
			}
			if handlePublicRedirect(c) {
				return true
			}
			if checkPrivateRedirect(c) {
				return true
			}
			if handlePrivateRedirect(c) {
				return true
			}
			return false
		}
		if path == webappCommon.ShortPath {
			headerUpdate(c)
		}
		log.Printf("serving %s\n", c.Request.URL)
		sh0r7H.ServeHTTP(c.Writer, c.Request)
		return true
	}
	return false
}

func fixPath(c *gin.Context) bool {
	if c.Request.Referer() != "" {
		log.Printf("referer not empty (%s)\n", c.Request.Referer())
		return false
	}
	if strings.Contains(c.Request.URL.Path, webappCommon.PrivatePath) {
		redirect := c.Request.URL
		redirect.Path = webappCommon.PrivatePath
		c.Redirect(http.StatusFound, redirect.String())
		log.Printf("redirect with private (%s)\n", redirect)
		return true
	}
	log.Printf("path not fixed")
	return false
}

func checkPrivateRedirect(c *gin.Context) bool {
	if c.Param("ext") != "" {
		return false
	}
	privateKey := c.Param("short")
	if dataKey, err := store.StoreCtx.LoadDataMapping(privateKey + store.SuffixPrivate); err == nil {
		if info, err := store.StoreCtx.LoadDataMappingInfo(string(dataKey) + store.SuffixPublic); err == nil {
			if v, ok := info[store.FieldPrivate]; ok && v == privateKey {
				redirect := &url.URL{
					Host:     c.Request.Host,
					Scheme:   c.Request.URL.Scheme,
					Path:     webappCommon.PrivatePath,
					RawQuery: webappCommon.FPrivateKey + "=" + privateKey,
				}
				if salt, ok := info[store.FieldPrvPassSalt]; ok {
					redirect.RawQuery += "&" + webappCommon.PasswordProtected + "=" + url.QueryEscape(salt.(string))
				}
				if strings.HasSuffix(c.Request.Referer(), redirect.String()) {
					return false
				}
				if c.Request.Header.Get(webappCommon.FPrvPassToken) != "" {
					return false
				}
				if c.Request.URL.Query().Has(webappCommon.FPass) {
					return false
				}

				c.Redirect(http.StatusFound, redirect.String())
				log.Printf("redirect with private key (%s)\n", redirect)
				return true
			}
		}
	}
	log.Printf("path not checked redirected")
	return false
}

func checkPublicRedirect(c *gin.Context) bool {
	if c.Param("ext") != "" {
		return false
	}
	publiceKey := c.Param("short")
	log.Printf("path public check if need redirect <%s>\n", c.Request.URL)
	if info, err := store.StoreCtx.LoadDataMappingInfo(string(publiceKey) + store.SuffixPublic); err == nil {
		if v, ok := info[store.FieldPublic]; ok && v == publiceKey {
			log.Printf("original url: %+#v\n", c.Request.URL)
			redirect, err := url.ParseRequestURI(c.Request.RequestURI)
			if err != nil {
				log.Printf("failed to parse request uri: %s\n", err)
				return false
			}
			redirect.Path = webappCommon.PublicPath
			redirect.RawQuery = webappCommon.FPrivateKey + "=" + publiceKey
			log.Printf("redirect url: %+#v\n", c.Request.URL)

			if strings.HasSuffix(c.Request.Referer(), redirect.String()) {
				return false
			}
			if c.Request.Header.Get(webappCommon.FPrvPassToken) != "" {
				return false
			}
			if c.Request.URL.Query().Has(webappCommon.FPass) {
				return false
			}

			if salt, ok := info[store.FieldPrvPassSalt]; ok {
				redirect.RawQuery += "&" + webappCommon.PasswordProtected + "=" + url.QueryEscape(salt.(string))
				c.Redirect(http.StatusFound, redirect.String())
				log.Printf("redirect with public key (%s)\n", redirect)
				return true
			}
		}
	}
	log.Printf("path public not redirected <%s>\n", c.Request.URL)
	return false
}

func handlePrivateRedirect(c *gin.Context) bool {
	if strings.Contains(c.Request.URL.Path, webappCommon.PrivatePath) {
		if key, ok := c.GetQuery("key"); ok {
			if dataKey, err := store.StoreCtx.LoadDataMapping(key + store.SuffixPrivate); err == nil {
				if info, err := store.StoreCtx.LoadDataMappingInfo(string(dataKey) + store.SuffixPublic); err == nil {
					if v, ok := info[store.FieldPrivate]; ok && v == key {
						log.Printf("!! serving path: <%s>\n", c.Request.RequestURI)
						sh0r7H.ServeHTTP(c.Writer, c.Request)
						return true
					}
				}
			}
		}
	}
	log.Printf("!! path not served: <%s>\n", c.Request.URL)
	return false
}

func handlePublicRedirect(c *gin.Context) bool {
	if strings.Contains(c.Request.URL.Path, webappCommon.PublicPath) {
		if key, ok := c.GetQuery("key"); ok {
			if info, err := store.StoreCtx.LoadDataMappingInfo(string(key) + store.SuffixPublic); err == nil {
				if v, ok := info[store.FieldPublic]; ok && v == key {
					log.Printf("!! serving path: <%s>\n", c.Request.RequestURI)
					sh0r7H.ServeHTTP(c.Writer, c.Request)
					return true
				}
			}
		}
	}
	log.Printf("!! path not served: <%s>\n", c.Request.URL)
	return false

}

func headerUpdate(c *gin.Context) {
	if c.Request.Header.Get(webappCommon.FRequestTokenSeed) == "" {
		return
	}
	log.Println("===========================================")
	log.Printf("request : %s\n", c.Request.Header)
	log.Printf("request UA: %s\n", c.Request.Header.Get("User-Agent"))
	log.Printf("clientIP : %s\n", c.ClientIP())
	log.Printf("client request uri : %s\n", c.Request.RequestURI)
	log.Printf("client request url : %s\n", c.Request.URL)
	// seed := generateSeedAndStoreToken(fmt.Sprintf("%s", c.Request.Header))
	seedLen, tokenLen := 32, 40
	seed, token := generateSeedAndToken(c.Request.Header.Get("User-Agent"), seedLen, tokenLen)
	log.Println("/*/*/*/*/*/*/*/*/*/*/*/*/")
	log.Printf("seed: <%s> (%d)\n", seed, seedLen)
	log.Printf("token: <%s> (%d)\n", token, tokenLen)
	log.Println("/*/*/*/*/*/*/*/*/*/*/*/*/")

	log.Printf("generated seed : <%s>\n", seed)
	c.Writer.Header().Add(webappCommon.FSaltTokenID, seed)                        // seed
	c.Writer.Header().Add(webappCommon.FSaltTokenID, fmt.Sprintf("%d", tokenLen)) // tokenLen
	c.Writer.Header().Add(webappCommon.FSaltTokenID, "0")                         // token start pos
	log.Println("===========================================")
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
