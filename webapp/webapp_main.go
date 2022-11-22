//go:build webapp

package webapp

import (
	"fmt"

	"github.com/gilwo/Sh0r7/common"
	_ "github.com/gilwo/Sh0r7/webapp/backend"
	webappCommon "github.com/gilwo/Sh0r7/webapp/common"
	_ "github.com/gilwo/Sh0r7/webapp/frontend"

	"github.com/gin-gonic/gin"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	// "github.com/maxence-charriere/go-app/v9/pkg/cli"
	// "github.com/maxence-charriere/go-app/v9/pkg/errors"
)

func init() {
	fmt.Println("webapp init invoked")
	common.WebappInit = webappInit
	common.WebappGenFunc = webappgenfunc
}

var (
	helloH *app.Handler = &app.Handler{
		Name:        "Sh0r7",
		Description: "Sh0r7 url and data shortener",
		// Resources:   app.CustomProvider(".", helloPath),
		Icon: app.Icon{
			Default: "/web/sh0r7-website-favicon-color.png",
			Large:   "/web/sh0r7-logo-color-on-transparent-background.png",
		},
		// Resources: app.LocalDir("/web"),
		// Resources: app.LocalDir(""),H1
		Styles: []string{
			"/web/sh0r7-main.css",
			"https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css",
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
		path := c.Request.RequestURI
		fmt.Printf("webappgenfunc: handling path: %s\n", path)
		if _, ok := webappServedPaths[path]; !ok {
			return false
		}
		helloH.ServeHTTP(c.Writer, c.Request)
		return true
	}
	return false
}
