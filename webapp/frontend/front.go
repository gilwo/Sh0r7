/*
//go:build webapp
*/

package frontend

import (
	"github.com/gilwo/Sh0r7/webapp/common"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func init() {
	common.WebappFront = mainfrtonend
}

func mainfrtonend() {
	app.Route(common.ShortPath, newShort())

	app.RunWhenOnBrowser()
}
