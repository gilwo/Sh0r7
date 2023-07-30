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
	short := newShort()
	app.Route(common.ShortPath, short)
	app.Route(common.PrivatePath, short)
	app.Route(common.PublicPath, short)
	app.Route(common.RemovePath, short)
	// app.Route(common.ClockPath, newClock())

	app.RunWhenOnBrowser()
}
