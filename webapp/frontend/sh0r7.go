package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gilwo/Sh0r7/shortener"
	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type short struct {
	app.Compo
	result          string
	resultMap       map[string]string
	resultReady     bool
	token           string
	isExpireChecked bool
	expireValue     string
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
		Body(
			app.Div().
				Class("marker").
				ID("headerNote"),
			app.Div().
				Class("note").
				// Style("", "").
				Body(
					app.H4().
						// Style("text-align", "center").
						Body(
							app.Text("under construction - not yet ready for live ...."),
							// app.Text("Resize the browser window to see the responsive effect."),
						),
				),
			app.Div().
				Class("marker").
				ID("headerTitle"),
			app.Div().
				Class("header", "row").
				Body(
					app.Div().
						Class("logo").
						Body(
							app.Img().
								Class("logo-img").
								Src("web/short-giraffe-0.jpg").
								Width(200),
						),
					app.Div().
						Class("text").
						Body(
							app.H1().
								Body(
									app.Text("Sh0r7"),
								),
							app.H2().
								// Style("text-align", "center").
								Body(
									app.B().
										Body(
											app.Text("Not"),
										),
									app.Text(" only URLs!"),
								),
						),
				),
			// app.Div().
			// 	Class("marker").
			// 	ID("navBarWide"),
			// app.Div().
			// 	Class("navbar", "show-in-wide").
			// 	// Style("", "").
			// 	Body(
			// 		app.A().
			// 			Href("#").
			// 			Body(
			// 				app.Text("Link"),
			// 			),
			// 		app.A().
			// 			Href("#").
			// 			Body(
			// 				app.Text("Link"),
			// 			),
			// 		app.A().
			// 			Href("#").
			// 			Body(
			// 				app.Text("Link"),
			// 			),
			// 		app.A().
			// 			Href("#").
			// 			Body(
			// 				app.Text("Link"),
			// 			),
			// 	),

			// app.Div().
			// 	Class("marker").
			// 	ID("navBarNarrow"),
			// app.Div().
			// 	Class("navbar", "show-in-narrow").
			// 	// Style("", "").
			// 	Body(
			// 		app.A().
			// 			Href("#").
			// 			Body(
			// 				app.Text("link narrow"),
			// 			),
			// 	),

			app.Div().
				Class("marker").
				ID("mainDo"),
			app.Div().
				Class("row", "shortDo").
				Body(
					app.Div().
						Class("shortInputWrapper").
						Body(
							app.If(!h.resultReady,
								app.Div().
									Class("shortInput").
									Body(
										app.Textarea().
											ID("shortInputText").
											Class("form-control").
											Rows(5).
											Cols(50).
											Wrap("off").
											Placeholder("long url or data to shorten..."),
									),
							).Else(
								app.Div().
									Class("shortOutput").
									Body(
										app.Div().
											Class("").
											Body(
												app.Div().
													Class("input-group").
													Body(
														app.Span().
															Class("input-group-addon", "fld-title").
															Styles(map[string]string{
																"float": "left",
																"width": "12%"}).
															Body(
																app.Text("sh0r7 public"),
															),
														app.Input().
															ID("short-public").
															Type("text").
															Class("form-control").
															ReadOnly(true).
															Styles(map[string]string{
																"float": "center",
																"width": "30%"}).
															Value(h.shortLink("short")),
														app.Span().
															Class("input-group-btn").
															Styles(map[string]string{
																"float": "center",
																"width": "10%"}).
															Body(
																app.Button().
																	Title("Copy to clipboard...").
																	ID("copy-public").
																	Class("btn", "btn-warning", "btn-copy").
																	Type("button").
																	Body(
																		app.Text("Copy"),
																	).
																	OnClick(func(ctx app.Context, e app.Event) {
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
											Class("").
											Body(
												app.Div().
													Class("input-group").
													Body(
														app.Span().
															Class("input-group-addon", "fld-title").
															Styles(map[string]string{
																"float": "left",
																"width": "12%"}).
															Body(
																app.Text("sh0r7 private"),
															),
														app.Input().
															ID("short-private").
															Type("text").
															Class("form-control").
															ReadOnly(true).
															Styles(map[string]string{
																"float": "center",
																"width": "30%"}).
															Value(h.shortLink("private")),

														app.Span().
															Class("input-group-btn").
															Styles(map[string]string{
																"float": "center",
																"width": "10%"}).
															Body(
																app.Button().
																	Title("Copy to clipboard...").
																	ID("copy-private").
																	Class("btn", "btn-warning", "btn-copy").
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
											Class("").
											Body(
												app.Div().
													Class("input-group").
													Body(
														app.Span().
															Class("input-group-addon", "fld-title").
															Styles(map[string]string{
																"float": "left",
																"width": "12%"}).
															Body(
																app.Text("sh0r7 delete"),
															),
														app.Input().
															ID("short-delete").
															Type("text").
															Class("form-control").
															ReadOnly(true).
															Styles(map[string]string{
																"float": "center",
																"width": "30%"}).
															Value(h.shortLink("delete")),
														app.Span().
															Class("input-group-btn").
															Styles(map[string]string{
																"float": "center",
																"width": "10%"}).
															Body(
																app.Button().
																	Title("Copy to clipboard...").
																	ID("copy-delete").
																	Class("btn", "btn-warning", "btn-copy").
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
						Class("shortButtonWrapper").
						Body(
							app.Div().
								Class("shortButtonPre"),
							app.Div().
								Class("shortButton").
								Body(
									app.If(!h.resultReady,
										app.Button().
											ID("shortInputButton").
											Class("btn", "btn-primary", "btn-lg", "btn-block").
											Body(
												app.Text("short it"),
											).
											OnClick(func(ctx app.Context, e app.Event) {
												elem := app.Window().GetElementByID("shortInputText")
												v := elem.Get("value").String()
												fmt.Printf("shortInputText value: %v\n", v)
												h.expireValue = ""
												if h.isExpireChecked {
													h.expireValue = app.Window().GetElementByID("expireSelect").Get("value").String()
													fmt.Printf("expire value: %v\n", h.expireValue)
												}
												if v != "" {
													ctx.Async(h.createShort)
												}
											}),
									).Else(
										app.Button().
											Class("btn", "btn-success", "btn-lg", "btn-block").
											Text("New").
											OnClick(func(ctx app.Context, e app.Event) {
												h.result = ""
												h.resultReady = false
												h.Update()
											}),
								),
							app.Div().
								Class("shortButtonPost"),
						),
				),
			app.Div().
				Class("marker").
				ID("mainOptions"),
			app.Div().
				Class("shortOptionsWrapper").
				Body(
					app.Div().
						Class("shortOptions").
						Body(
							app.H3().
								Body(
									app.Text("Options"),
								),
							app.Div().
								ID("shortOption2").
								// Style("", "").
								Body(
									app.Div().
										Class("form-group").
										Body(
											app.Div().
												Class("input-group").
												Body(
													app.Label().
														Class("input-group-addon").
														ID("expireCheckBox").
														// Style("", "").
														Body(
															app.Input().
																Type("checkbox").
																OnClick(func(ctx app.Context, e app.Event) {
																	h.isExpireChecked = ctx.JSSrc().Get("checked").Bool()
																	fmt.Printf("useExpire: %v\n", h.isExpireChecked)
																}),
															app.Text("Expiration"),
														),
													app.If(h.isExpireChecked,
														app.Span().
															Dir("ltr").
															Style("margin", "10px").
															Body(
																// app.Text("  "),
																app.Select().
																	Class("form-control", "shortSelect").
																	ID("expireSelect").
																	Body(
																		app.Option().
																			Value("10m").
																			Body(
																				app.Text("10 minutes"),
																			),
																		app.Option().
																			Value("12h").
																			Selected(true).
																			Body(
																				app.Text("12 hours"),
																			),
																		app.Option().
																			Value("2d").
																			Body(
																				app.Text("2 days"),
																			),
																		app.Option().
																			Value("2w").
																			Body(
																				app.Text("2 weeks"),
																			),
																		app.Option().
																			Value("2mo").
																			Body(
																				app.Text("2 months"),
																			),
																		app.Option().
																			Value("2y").
																			Body(
																				app.Text("year"),
																			),
																		app.Option().
																			Value("n").
																			Body(
																				app.Text("never"),
																			),
																	).
																	OnChange(func(ctx app.Context, e app.Event) {
																		h.expireValue = ctx.JSSrc().Get("value").String()
																		fmt.Printf("select change value: %v\n", h.expireValue)
																	}),
															),
													),
												),
										),
								),
						),
				),
			app.Div().
				Class("marker").
				ID("footer"),
			app.Div().
				Class("footer").
				Body(
					app.Textarea().
						ID("footerText"),
				),
		)
}

func (h *short) Render3() app.UI {
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
												Class("input-group-addon", "fld-title").
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
														Class("btn", "btn-warning", "btn-copy").
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
												Class("input-group-addon", "fld-title").
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
														Class("btn", "btn-warning", "btn-copy").
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
												Class("input-group-addon", "fld-title").
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
														Class("btn", "btn-warning", "btn-copy").
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
							Class("btn", "btn-primary", "btn-lg", "btn-block").
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
							Class("btn", "btn-success", "btn-lg", "btn-block").
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
	h.getStID()
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
	elem := app.Window().GetElementByID("shortInputText")
	data := elem.Get("value").String()
	errElem := app.Window().GetElementByID("footerText")
	destCreate := "http://" + host
	payload := []byte(data)

	if url, ok := urlCheck(data); ok {
		destCreate += "/create-short-url"
		payload, err = json.Marshal(map[string]string{
			"url": url,
		})
		if err != nil {
			errElem.Set("value", fmt.Sprintf("url problem: error occurred: %s", err))
			return
		}
	} else {
		destCreate += "/create-short-data"
	}

	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	// fmt.Printf("app %#v\n", app.)
	req, err := http.NewRequest(http.MethodPost, destCreate, bytes.NewBuffer(payload))
	if err != nil {
		errElem.Set("value", fmt.Sprintf("new request: error occurred: %s", err))
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("TID", h.token)
	resp, err := client.Do(req)
	if err != nil {
		errElem.Set("value", fmt.Sprintf("request invoke: error occurred: %s", err))
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errElem.Set("value", fmt.Sprintf("response read: error occurred: %s", err))
		return
	}
	if resp.StatusCode != http.StatusOK {
		errElem.Set("value", fmt.Sprintf("response status: : %v", resp.StatusCode))
		return
	}

	// elem := app.Window().GetElementByID("in-out")
	err = json.Unmarshal(body, &h.resultMap)
	if err != nil {
		errElem.Set("value", fmt.Sprintf("response read: error occurred: %s", err))
		return
	}

	r, err := json.MarshalIndent(h.resultMap, "", "\t")
	if err != nil {
		errElem.Set("value", fmt.Sprintf("response read: error occurred: %s", err))
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
	x := app.Window().URL()
	x.Path = "/"
	fmt.Printf("!# path: %#+v\n", x)
	host := x.String()
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
	h.getUserAgent()
}

func (h *short) getUserAgent() {
	ua := app.Window().Get("navigator").Get("userAgent")
	uaData := app.Window().Get("navigator").Get("userAgentData")
	fmt.Printf("*** UA: %s\n", ua)
	fmt.Printf("*** UAData platform: %s\n", uaData.Get("platform"))
	fmt.Printf("*** UAData mobile: %s\n", uaData.Get("mobile"))
	fmt.Printf("*** UAData brands brand: %s\n", uaData.Get("brands").Index(0).Get("brand"))
	fmt.Printf("*** UAData brands version: %s\n", uaData.Get("brands").Index(0).Get("version"))
}

func (h *short) getStID() {

	var err error
	urlApp := app.Window().URL().String()

	client := http.Client{
		Timeout: time.Duration(2 * time.Second),
	}
	req, err := http.NewRequest(http.MethodGet, urlApp, nil)
	if err != nil {
		app.Logf("failed to create new request: %s\n", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("RTS", shortener.GenerateToken2(uuid.NewString(), 0, -1))
	resp, err := client.Do(req)
	if err != nil {
		app.Logf("failed to invoke request: %s\n", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		app.Logf("response not ok: %s\n", resp.StatusCode)
		return
	}
	_stid := resp.Header.Get("stid")
	stid := strings.Split(_stid, ", ")
	if len(stid) != 3 {
		app.Logf("problem with stid: %#v\n", stid)
		return
	}
	seed := stid[0]
	tokenLen, err := strconv.Atoi(stid[1])
	if err != nil {
		app.Logf("problem with number convertion: %s\n", err)
		return
	}
	tokenStartPos, err := strconv.Atoi(stid[2])
	if err != nil {
		app.Logf("problem with number convertion: %s\n", err)
		return
	}

	fmt.Printf("******************************* stid from header: %+#v\n", stid)
	ua := app.Window().Get("navigator").Get("userAgent").String()

	token := shortener.GenerateToken2(ua+seed, tokenLen, tokenStartPos)
	fmt.Printf("******************************* calculated token: %s\n", token)
	if token == "" {
		app.Logf("problem with token generation\n")
		return
	}
	h.token = token
}
