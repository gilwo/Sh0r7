package common

var (
	MainServer    func()
	WebappInit    func()
	WebappGenFunc func(args ...interface{}) interface{}
)
