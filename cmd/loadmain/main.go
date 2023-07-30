// You can edit this code!
// Click here and start typing.
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func init() {
	log.Default().SetFlags(log.Flags() | log.Llongfile)
}
func main() {
	fmt.Println("Hello, 世界")
	loadMain()
}

func loadMain() {
	var err error
	lurl := "https://54f4-85-65-160-71.eu.ngrok.io"
	var redirectLoc *url.URL

	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirectLoc, err = req.Response.Location()
			log.Printf("redirectLoc : %+#v, err: %v\n", redirectLoc, err)
			// if redirectLoc != nil {
			// return errors.New("redirect location")
			// }
			return nil
		},
	}
	req, err := http.NewRequest(http.MethodGet, lurl, nil)
	if err != nil {
		log.Printf("failed to create new request: %s\n", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	// req.Header.Set("Content-Type", "text/plain")
	x, err := httputil.DumpRequest(req, true)
	log.Printf("invoking request: %+#v, err: %v\n", string(x), err)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to invoke request: %s\n", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("response not ok: %v\n", resp.StatusCode)
		return
	}
	// x, err = httputil.DumpResponse(resp, true)
	// app.Logf("getting response location: %#+v, err: %v\n", string(x), err)
	loc, err := resp.Location()
	log.Printf("repsonse location: [%#+v], error [%v]\n", loc, err)

	// defer resp.Body.Close()
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	h.handleError("response reads", err)
	// 	return
	// }
	// app.Logf("response: %+#v\n", body)
	log.Printf("redirect Location: %+#v\n", redirectLoc)
}

