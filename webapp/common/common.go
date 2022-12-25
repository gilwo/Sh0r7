package common

var (
	WebappFront    func()
	WebappBack     func()
	ShortPath      = "/"
	PrivatePath    = ShortPath + "private"
	DevShortPath   = "/testapp"
	DevPrivatePath = DevShortPath + "/private"
	devBuild       bool
)

func init() {
	if devBuild {
		ShortPath = DevShortPath
		PrivatePath = DevPrivatePath
	}
}
