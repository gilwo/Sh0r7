package common

var (
	IsDevEnv      bool
	BuildVersion  string
	SourceTime    string
	BuildTime     string
	MainServer    func()
	WebappInit    func()
	WebappGenFunc func(args ...interface{}) interface{}
)
