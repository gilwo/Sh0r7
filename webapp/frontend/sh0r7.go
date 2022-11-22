package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type short struct {
	app.Compo

	result      string
	resultMap   map[string]string
	resultReady bool
}

func (h *short) Render2() app.UI {
	return app.Div().Body(
		app.Div().Class("image-title").Body(
			app.Img().
				Alt("sh0r7-logo").
				Src("/web/sh0r7-logo-color-on-transparent-background.png").
				Width(400).Height(300),
		),
		app.Div().Body(
			app.H1().Class("hello-title").Text("Hello beutiful World! 2"),
		))
}

func (h *short) Render() app.UI {
	return app.Div().
		Class("v2_3").
		Body(
			app.Div().
				Class("v9_30"),
			app.Div().
				Class("v2_12"),
			app.Div().
				Class("v2_13"),
			app.Div().
				Class("v2_15"),
			app.Div().
				Class("v2_16"),
			app.Span().
				Class("v2_19").
				Body(
					app.Text("Sh0r7"),
				),
			app.Div().
				Class("v2_20"),
			app.Span().
				Class("v2_21").
				Body(
					app.Text("not only urls"),
				),
			app.Div().
				Class("v2_22").Body(
				app.If(!h.resultReady,
					app.Textarea().
						ID("in-out").
						Class("form-control").
						Rows(5).
						Cols(20).
						Wrap("off").
						Placeholder("long url or data..."),
				).Else(
					app.Div().
						Class("row").
						Body(
							app.Div().
								Class("col-lg-6").
								Body(
									app.Div().
										Class("input-group").
										Body(
											app.Span().
												Class("input-group-addon fld-title").
												Body(
													app.Text("sh0r7 public"),
												),
											app.Input().
												ID("short-public").
												Type("text").
												Class("form-control").
												ReadOnly(true).
												Value(h.shortLink("short")),
											app.Span().
												Class("input-group-btn").
												Body(
													app.Button().
														ID("copy-public").
														Class("btn btn-warning btn-copy").
														Type("button").
														Body(
															app.Text("Copy"),
														).OnClick(func(ctx app.Context, e app.Event) {
														h.copyToClipboard("short-public")
														elem := app.Window().GetElementByID("copy-public")
														fmt.Printf("current value: %v\n", elem.Get("body"))
														elem.Set("textContent", "Copied")
														ctx.After(400*time.Millisecond, func(ctx app.Context) {
															elem.Set("textContent", "Copy")
														})
													}),
												),
										),
								),
							app.Div().
								Class("col-lg-6").
								Body(
									app.Div().
										Class("input-group").
										Body(
											app.Span().
												Class("input-group-addon fld-title").
												Body(
													app.Text("sh0r7 private"),
												),
											app.Input().
												ID("short-private").
												Type("text").
												Class("form-control").
												ReadOnly(true).
												Value(h.shortLink("private")),

											app.Span().
												Class("input-group-btn").
												Body(
													app.Button().
														ID("copy-private").
														Class("btn btn-warning btn-copy").
														Type("button").
														Body(
															app.Text("Copy"),
														).OnClick(func(ctx app.Context, e app.Event) {
														h.copyToClipboard("short-private")
														elem := app.Window().GetElementByID("copy-private")
														fmt.Printf("current value: %v\n", elem.Get("body"))
														elem.Set("textContent", "Copied")
														ctx.After(400*time.Millisecond, func(ctx app.Context) {
															elem.Set("textContent", "Copy")
														})
													}),
												),
										),
								),
							app.Div().
								Class("col-lg-6").
								Body(
									app.Div().
										Class("input-group").
										Body(
											app.Span().
												Class("input-group-addon fld-title").
												Body(
													app.Text("sh0r7 delete"),
												),
											app.Input().
												ID("short-delete").
												Type("text").
												Class("form-control").
												ReadOnly(true).
												Value(h.shortLink("delete")),
											app.Span().
												Class("input-group-btn").
												Body(
													app.Button().
														ID("copy-delete").
														Class("btn btn-warning btn-copy").
														Type("button").
														Body(
															app.Text("Copy"),
														).OnClick(func(ctx app.Context, e app.Event) {
														h.copyToClipboard("short-delete")
														elem := app.Window().GetElementByID("copy-delete")
														fmt.Printf("current value: %v\n", elem.Get("body"))
														elem.Set("textContent", "Copied")
														ctx.After(400*time.Millisecond, func(ctx app.Context) {
															elem.Set("textContent", "Copy")
														})
													}),
												),
										),
								),
						),
				),
			),
			app.Div().
				Class("v2_23"),
			app.Span().
				Class("v2_24").
				Body(
					app.If(!h.resultReady,
						app.Button().
							Class("btn btn-primary btn-lg btn-block").
							Text("short it").
							OnClick(func(ctx app.Context, e app.Event) {
								elem := app.Window().GetElementByID("in-out")
								v := elem.Get("value")
								fmt.Printf("in-out value: %v\n", v)
								if v.String() != "" {
									ctx.Async(h.createShort)
								}
							}),
					).Else(
						app.Button().
							Class("btn btn-success btn-lg btn-block").
							Text("New").
							OnClick(func(ctx app.Context, e app.Event) {
								h.result = ""
								h.resultReady = false
								h.Update()
							}),
					),
				),
			app.Div().
				Class("v6_26"),
			app.Div().
				Class("v9_29"),
			app.Span().
				Class("v6_27").
				Body(
					app.Text("option"),
				),
			app.Div().
				Class("v7_28"),
		)
}

func newShort() *short {
	return &short{}
}

func (h *short) OnInit() {
	fmt.Println("******************************* init")
}
func (h *short) OnPreRender() {
	fmt.Println("******************************* prerender")
}
func (h *short) OnDisMount() {
	fmt.Println("******************************* dismount")
}
func (h *short) OnMount() {
	fmt.Println("******************************* mount")
}
func (h *short) OnNav() {
	fmt.Println("******************************* nav")
}
func (h *short) OnUpdate() {
	fmt.Println("******************************* update")
}
func (h *short) OnAppUpdate() {
	fmt.Println("******************************* app update")
}

func urlCheck(s string) (string, bool) {
	s = strings.TrimRight(s, "\n")
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		s = "https://" + s
		u, err = url.Parse(s)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return "", false
		}
	}
	return u.String(), true
}
func (h *short) createShort() {
	var err error
	host := app.Window().URL().Host
	elem := app.Window().GetElementByID("in-out")
	data := elem.Get("value").String()
	destCreate := "http://" + host
	payload := []byte(data)

	if url, ok := urlCheck(data); ok {
		destCreate += "/create-short-url"
		payload, err = json.Marshal(map[string]string{
			"url": url,
		})
		if err != nil {
			elem.Set("value", fmt.Sprintf("url problem: error occurred: %s", err))
			return
		}
	} else {
		destCreate += "/create-short-data"
	}

	client := http.Client{
		Timeout: time.Duration(1 * time.Second),
	}
	// fmt.Printf("app %#v\n", app.)
	req, err := http.NewRequest(http.MethodPost, destCreate, bytes.NewBuffer(payload))
	if err != nil {
		elem.Set("value", fmt.Sprintf("new request: error occurred: %s", err))
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	resp, err := client.Do(req)
	if err != nil {
		elem.Set("value", fmt.Sprintf("request invoke: error occurred: %s", err))
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		elem.Set("value", fmt.Sprintf("response read: error occurred: %s", err))
		return
	}
	if resp.StatusCode != http.StatusOK {
		elem.Set("value", fmt.Sprintf("response status: : %v", resp.StatusCode))
		return
	}

	// elem := app.Window().GetElementByID("in-out")
	err = json.Unmarshal(body, &h.resultMap)
	if err != nil {
		elem.Set("value", fmt.Sprintf("response read: error occurred: %s", err))
		return
	}

	r, err := json.MarshalIndent(h.resultMap, "", "\t")
	if err != nil {
		elem.Set("value", fmt.Sprintf("response read: error occurred: %s", err))
		return
	}
	h.result = string(r)
	h.resultReady = true

	fmt.Printf("******************************* create short result: %s\n", string(body))
	elem.Set("value", string(body))
	fmt.Printf("******************************* create shoty: %#v\n", r)
	h.Update()
}

func (h *short) shortLink(which string) string {
	host := app.Window().URL().String()
	switch which {
	case "private", "short", "delete":
	default:
		// error
	}
	return host + h.resultMap[which]
}

func (h *short) copyToClipboard(from string) {
	elem := app.Window().GetElementByID(from)
	clipboard := app.Window().Get("navigator").Get("clipboard")
	clipboard.Call("writeText", elem.Get("value"))
}
