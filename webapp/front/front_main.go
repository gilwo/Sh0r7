package main

import (
	"github.com/gilwo/Sh0r7/webapp/common"
	_ "github.com/gilwo/Sh0r7/webapp/frontend"
)

func main() {
	if common.WebappFront == nil {
		panic("webapp front code is missing")
	}
	common.WebappFront()
}
