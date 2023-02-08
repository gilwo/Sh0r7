package common

var (
	IsDevEnv      bool
	MainServer    func()
	WebappInit    func()
	WebappGenFunc func(args ...interface{}) interface{}
)
