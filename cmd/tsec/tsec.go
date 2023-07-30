package main

import (
	"fmt"
	"time"
)

func main() {
	t0 := time.Now().Truncate(time.Hour * 24)
	t1 := time.Since(t0).Seconds()
	tnsec := int64(t1)
	fmt.Println(tnsec)
}
