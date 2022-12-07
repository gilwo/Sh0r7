package common

var (
	WebappFront func()
	WebappBack  func()
	ShortPath   = "/testapp"
	PrivatePath = ShortPath + "/private"
)
