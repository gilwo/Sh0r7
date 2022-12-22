package common

import (
	"os"
	"strings"
)

var (
	WebappFront    func()
	WebappBack     func()
	ShortPath      = "/"
	PrivatePath    = ShortPath + "private"
	DevShortPath   = "/testapp"
	DevPrivatePath = ShortPath + "/private"
)

func init() {
	envProd := os.Getenv("SH0R7_PRODUCTION")
	if envProd == "" || strings.ToLower(envProd) != "true" {
		ShortPath = DevShortPath
		PrivatePath = DevPrivatePath
	}
}
