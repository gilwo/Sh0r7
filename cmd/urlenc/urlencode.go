package main

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("not enough arguments")
		return
	}
	if os.Args[1] == "-r" { //reverse - unecape
		fmt.Println("reversing")
		for _, e := range os.Args[2:] {
			o, err := unEscapeString(e)
			fmt.Printf("%s: <%s> (%v)\n", e, o, err)
		}
	} else {
		for _, e := range os.Args[1:] {
			fmt.Printf("%s: <%s>\n", e, escapeString(e))
		}
	}
}
func unEscapeString(s string) (string, error) {
	return url.PathUnescape(s)
}

func escapeString(s string) string {
	var buf bytes.Buffer
	for _, c := range s {
		buf.WriteString(fmt.Sprintf("%%%02X", c))
	}
	return buf.String()
}
