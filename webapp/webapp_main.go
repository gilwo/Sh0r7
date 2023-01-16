//go:build webapp

package webapp

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
		webappCommon.RemovePath:  true,
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

		// redirect path when we in dev build
		if path == "/" && webappCommon.ShortPath == webappCommon.DevShortPath {
			redirect, err := url.ParseRequestURI(c.Request.RequestURI)
			redirect.Path = webappCommon.ShortPath
			if err == nil {
				c.Redirect(http.StatusFound, redirect.String())
				return true
			}
		}
		if _, ok := webappServedPaths[path]; !ok {
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
			if checkRemoveRedirect(c) {
				return true
			}
			if handleRemoveRedirect(c) {
				return true
			}
			return false
		}
		if path == webappCommon.ShortPath {
			headerUpdate(c)
		}
		log.Printf("serving %s\n", c.Request.URL)
		if path == webappCommon.ShortPath && queryUpdate(c) {
			return true
		}
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
					RawQuery: webappCommon.FShortKey + "=" + privateKey,
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
func getPublicInfo(publiceKey string) map[string]interface{} {
	if info, err := store.StoreCtx.LoadDataMappingInfo(string(publiceKey) + store.SuffixPublic); err == nil {
		return info
	}
	// try get the named hash ... ?
	hashedName := shortener.GenerateShortDataTweakedWithStore2NotRandom(publiceKey+store.SuffixPublic, 0, webappCommon.HashLengthNamedFixedSize, 0, 0, store.StoreCtx)
	if info, err := store.StoreCtx.LoadDataMappingInfo(string(hashedName) + store.SuffixPublic); err == nil {
		return info
	}
	return nil
}

func checkPublicRedirect(c *gin.Context) bool {
	if c.Param("ext") != "" {
		return false
	}
	publiceKey := c.Param("short")
	log.Printf("path public check if need redirect <%s>\n", c.Request.URL)
	if info := getPublicInfo(publiceKey); info != nil {
		v, ok := info[store.FieldPublic]
		v2, ok2 := info[store.FieldNamedPublic]
		if (ok && v == publiceKey) || (ok2 && v2 == publiceKey) {
			log.Printf("original url: %+#v\n", c.Request.URL)
			redirect, err := url.ParseRequestURI(c.Request.RequestURI)
			if err != nil {
				log.Printf("failed to parse request uri: %s\n", err)
				return false
			}
			redirect.Path = webappCommon.PublicPath
			redirect.RawQuery = webappCommon.FShortKey + "=" + publiceKey
			log.Printf("redirect url: %+#v\n", c.Request.URL)

			if strings.HasSuffix(c.Request.Referer(), redirect.String()) {
				return false
			}
			if c.Request.Header.Get(webappCommon.FPubPassToken) != "" {
				return false
			}
			if c.Request.URL.Query().Has(webappCommon.FPass) {
				return false
			}

			if salt, ok := info[store.FieldPubPassSalt]; ok {
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

func checkRemoveRedirect(c *gin.Context) bool {
	if c.Param("ext") != "" {
		return false
	}
	removeKey := c.Param("short")

	log.Printf("path remove check if need redirect <%s>\n", c.Request.URL)
	if dataKey, err := store.StoreCtx.LoadDataMapping(removeKey + store.SuffixRemove); err == nil {
		if info, err := store.StoreCtx.LoadDataMappingInfo(string(dataKey) + store.SuffixPublic); err == nil {
			if v, ok := info[store.FieldRemove]; ok && v == removeKey {

				log.Printf("original url: %+#v\n", c.Request.URL)
				redirect, err := url.ParseRequestURI(c.Request.RequestURI)
				if err != nil {
					log.Printf("failed to parse request uri: %s\n", err)
					return false
				}
				redirect.Path = webappCommon.RemovePath
				redirect.RawQuery = webappCommon.FShortKey + "=" + removeKey
				log.Printf("redirect url: %+#v\n", c.Request.URL)

				if strings.HasSuffix(c.Request.Referer(), redirect.String()) {
					return false
				}
				if c.Request.Header.Get(webappCommon.FRemPassToken) != "" {
					return false
				}
				if c.Request.URL.Query().Has(webappCommon.FPass) {
					return false
				}

				if salt, ok := info[store.FieldRemPassSalt]; ok {
					redirect.RawQuery += "&" + webappCommon.PasswordProtected + "=" + url.QueryEscape(salt.(string))
					c.Redirect(http.StatusFound, redirect.String())
					log.Printf("redirect with remove key (%s)\n", redirect)
					return true
				}
			}
		}
	}
	log.Printf("path remove not redirected <%s>\n", c.Request.URL)
	return false
}

func handlePrivateRedirect(c *gin.Context) bool {
	if strings.Contains(c.Request.URL.Path, webappCommon.PrivatePath) {
		if key, ok := c.GetQuery(webappCommon.FShortKey); ok {
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
		if key, ok := c.GetQuery(webappCommon.FShortKey); ok {
			if info := getPublicInfo(key); info != nil {
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

func handleRemoveRedirect(c *gin.Context) bool {
	if strings.Contains(c.Request.URL.Path, webappCommon.RemovePath) {
		if key, ok := c.GetQuery(webappCommon.FShortKey); ok {
			if dataKey, err := store.StoreCtx.LoadDataMapping(key + store.SuffixRemove); err == nil {
				if info, err := store.StoreCtx.LoadDataMappingInfo(string(dataKey) + store.SuffixPublic); err == nil {
					if v, ok := info[store.FieldRemove]; ok && v == key {
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
func checkSaltTokenStillValid(c *gin.Context) bool {
	qVals := c.Request.URL.Query()
	if len(qVals) < 1 { // all values under the same field
		return false
	}
	stid, ok := qVals[webappCommon.FSaltTokenID]
	if !ok {
		log.Printf("not field %s\n", webappCommon.FSaltTokenID)
		return false
	}
	x, err := shortener.Base64SE.Decode(stid[0])
	if err != nil {
		app.Logf("problem with stid : %s\n", err)
		return false
	}
	stid = strings.Split(string(x), "$")
	seed := stid[0]
	tokenLen, err := strconv.Atoi(stid[1])
	if err != nil {
		log.Printf("problem with number convertion: %s\n", err)
		return false
	}
	tokenStartPos, err := strconv.Atoi(stid[2])
	if err != nil {
		log.Printf("problem with number convertion: %s\n", err)
		return false
	}

	// calculate the token
	ua := c.Request.UserAgent()
	token := shortener.GenerateTokenTweaked(ua+seed, tokenStartPos, tokenLen, 0)

	// now check the validity of the token
	info, err := store.StoreCtx.LoadDataMappingInfo(token)
	if err != nil {
		log.Printf("failed to get info for token: <%s>\n", token)
		return false
	}

	v, ok := info[store.FieldTime]
	if !ok {
		log.Printf("failed to get created time value for on token: <%s>\n", token)
		return false
	}
	when := v.(string)
	// if time.Parse(when)
	t, err := time.Parse(time.RFC3339, when)
	if err != nil {
		log.Printf("failed to get parsed created time (%s) value for on token: <%s>\n", when, token)
		return false
	}
	log.Printf("created time for key: <%s> : before parse [%s], after parse [%s]\n", token, when, t)
	v, ok = info[store.FieldTTL]
	if !ok {
		log.Printf("failed to get ttl value for on token: <%s>\n", token)
		return false
	}
	ttl, err := time.ParseDuration(v.(string))
	if err != nil {
		log.Printf("failed to get parsed duration time (%s) value for on token: <%s>\n", v, token)
		return false
	}

	if time.Since(t) > ttl {
		log.Printf("token (%s) expired : created <%s> ttl <%s>\n", token, t, ttl)
		return false
	}
	return true
}

func queryUpdate(c *gin.Context) bool {
	if c.Request.URL.Query().Has(webappCommon.FSaltTokenID) {
		if checkSaltTokenStillValid(c) {
			log.Printf("path already has the token seed and it is valid, no need to add it...")
			return false
		}
		log.Printf("path token is invalid redirect with new token")
	}
	refererUrl, err := url.Parse(c.Request.Referer())
	if err != nil {
		log.Printf("skip redirect for url parse failed for referrer... <%s>", c.Request.Referer())
		return false
	}
	if strings.HasSuffix(refererUrl.Path, "app-worker.js") {
		log.Printf("skip redirect for referrer url <%s>", c.Request.Referer())
		return false
	}
	seedLen, tokenLen := 32, 40
	seed, token := generateSeedAndToken(c.Request.Header.Get("User-Agent"), seedLen, tokenLen)
	log.Println("/*/*/*/*/*/*/*/*/*/*/*/*/")
	log.Printf("seed: <%s> (%d)\n", seed, seedLen)
	log.Printf("token: <%s> (%d)\n", token, tokenLen)
	log.Println("/*/*/*/*/*/*/*/*/*/*/*/*/")
	isDev := c.Request.URL.Query().Has("dev")
	redirect, err := url.ParseRequestURI(c.Request.RequestURI)
	if err != nil {
		log.Printf("failed to parse request uri: %s\n", err)
		return false
	}
	redirect.Path = webappCommon.ShortPath
	redirect.RawQuery = ""
	q := redirect.Query()
	// q.Add(webappCommon.FSaltTokenID, seed)
	// q.Add(webappCommon.FSaltTokenID, fmt.Sprintf("%d", tokenLen))
	// q.Add(webappCommon.FSaltTokenID, "0")
	stidValue := fmt.Sprintf("%s$%d$%d", seed, tokenLen, 0)
	if isDev {
		stidValue += "$dev"
	}
	q.Add(webappCommon.FSaltTokenID, shortener.Base64SE.Encode([]byte(stidValue)))
	redirect.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, redirect.String())
	log.Printf("redirect with token seed (%s)\n", redirect)
	return true
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
