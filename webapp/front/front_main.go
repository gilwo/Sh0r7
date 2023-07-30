package main

import (
	"fmt"

	"github.com/gilwo/Sh0r7/common"
	webappCommon "github.com/gilwo/Sh0r7/webapp/common"
	_ "github.com/gilwo/Sh0r7/webapp/frontend"
)

var BB string = "dev"

func main() {
	if webappCommon.WebappFront == nil {
		panic("webapp front code is missing")
	}
	fmt.Printf("build version : %s\n", common.BuildVersion)
	fmt.Printf("source time : %s\n", common.SourceTime)
	fmt.Printf("build time : %s\n", common.BuildTime)
	// pkgPath := reflect.TypeOf(frontend.BuildVer) //.PkgPath()
	// fmt.Printf("package path: %+#v\n", pkgPath)
	webappCommon.WebappFront()
}
