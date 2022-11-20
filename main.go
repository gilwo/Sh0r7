package main

import (
	"github.com/gilwo/Sh0r7/common"
	_ "github.com/gilwo/Sh0r7/server"
)

func main() {
	if common.MainServer == nil {
		panic("server start failed")
	}
	common.MainServer()
}
