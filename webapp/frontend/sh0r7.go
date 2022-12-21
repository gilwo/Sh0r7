package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gilwo/Sh0r7/shortener"
	"github.com/gilwo/Sh0r7/store"
	webappCommon "github.com/gilwo/Sh0r7/webapp/common"
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
	isShortAsData   bool
	expireValue     string
	debug           bool
	isPrivate       bool
}

const (
	NOTEMESSAGE = "Sh0r7 service is still in alpha!"
)

var (
	ImgSource = "/web/logo.jpg"
	// imgSource: "logoL.png",
)

func (h *short) RenderPrivate() app.UI {
	out, keys, err := h.getPrivateInfo()
	if err != nil {
		out = map[string]string{"error": err.Error()}
		keys = []string{"error"}
	}
	return app.Div().
		Class("container").
		Body(
			app.Div().
				Class("row").
				Body(
					app.Div().
						Class("col-xs-8", "col-xs-offset-2").
						Body(
							app.H2().
								Body(app.Text("Sh0r7 private details")),
						),
				),
			app.Div().
				Class("row").
				Body(
					app.Div().
						Class("col-xs-6", "col-xs-offset-3").
						Body(
							app.H3().
								ID("privateTitle").
								Body(
									// app.Text("using private for "+app.Window().URL().String()),
									app.Text(app.Window().URL().Query().Get("key")+" deatils"),
								),
							app.Br(),
						),
				),
			app.Div().
				Class("row").
				Body(
					app.Div().
						Class("col-xs-8", "col-xs-offset-1").
						Body(
							app.Table().
								Class("table", "table-hover").
								Body(
									app.TBody().
										Body(
											app.Range(keys).Slice(func(i int) app.UI {
												s := keys[i]
												return app.Tr().
													Class().
													Body(
														app.Td().
															Class("result").
															// Class(s).
															Styles(map[string]string{
																// "vertical-align": "middle",
															}).
															Body(
																app.Text(s),
															),
														app.Td().
															Class("result").
															// Class(s+"Value").
															Body(
																// <div class="form-group">
																// <div class="1input-group has-success">
																// <!-- <div class="input-group-addon"></div> -->
																// <input class="form-control syncTextStyle" value="1234" readonly>
																// <!-- <div class="input-group-addon" ></div> -->
																// </div>
																// </div>
																app.Div().
																	Class("form-group").
																	Class("resultForm").
																	Body(
																		// app.Text(out[s]),
																		app.Div().
																			Class("1input-group", "has-success").
																			Body(
																				app.Input().
																					Class("form-control", "syncTextStyle").
																					Value(out[s]).
																					ReadOnly(true),
																			),
																	),
															),
													)
												// return app.Div().
												// 	Class("input-group").
												// 	Body(
												// 		app.Span().
												// 			Class("").
												// 			Styles(map[string]string{
												// 				"float": "left",
												// 				"width": "12%"}).
												// 			Body(
												// 				app.Text(s),
												// 			),
												// 		app.Input().
												// 			ID("").
												// 			Type("text").
												// 			Class("").
												// 			ReadOnly(true).
												// 			Styles(map[string]string{
												// 				"float": "center",
												// 				"width": "30%"}).
												// 			Value(out[s]),
												// 	)
											}),
										),
								),
						),
				),
			app.Br(),
		)
}

func (h *short) Render() app.UI {
	if h.isPrivate {
		return h.RenderPrivate()
	}
	return app.Div().
		ID("mainWrapper").
		Class("container").
		Body(
			app.Div().
				Class("row").
				Class("marker").
				ID("headerNote"),
			app.Div().
				Class("row").
				Class("note").
				Body(
					app.Div().
						Class("col-xs-8", "col-xs-offset-2").
						Body(
							app.H4().
								Styles(
									map[string]string{
										"background": "yellow",
										"text-align": "left",
										"width":      "fit-content"}).
								Body(
									app.Text(NOTEMESSAGE),
								),
						),
					app.If(h.debug,
						app.Div().
							Styles(map[string]string{
								"position":    "absolute",
								"margin-left": "450px",
								// "float":    "right",
							}).
							Body(
								app.P().
									ID("messages"),
							),
					),
				),
			app.Div().
				Class("row").
				Class("marker").
				ID("headerTitle"),
			app.Div().
				Class("row").
				Class("header").
				ID("logoTitle").
				Body(
					app.Div().
						Class("col-md-4", "col-md-offset-2", "col-sm-4", "col-sm-offset-2", "col-xs-4", "col-xs-offset-3").
						Class("logo").
						Body(
							app.Img().
								Class("logo-img").
								Class("img-responsive").
								Src(ImgSource).
								Alt("Sh0r7 Logo").
								Width(200),
						),
					app.Div().
						Class("col-md-6", "col-md-offset-0", "col-sm-4", "col-sm-offset-0", "col-xs-6", "col-xs-offset-3").
						Class("text").
						Body(
							app.H1().
								Body(
									app.Text("Sh0r7"),
								),
							app.H2().
								Styles(
									map[string]string{
										"margin-left": "40px",
										"text-align":  "left",
									}).
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
				Class("row").
				Class("marker").
				ID("mainDo"),
			app.Div().
				Class("row").
				Class("shortDo").
				Body(
					app.Div().
						Class("col-xs-8", "col-xs-offset-2", "shortInputWrapper").
						Class("shortInputWrapper").
						Body(
							app.If(!h.resultReady,
								app.Div().
									Class("shortInput").
									Body(
										app.Textarea().
											ID("shortInputText").
											Class("form-control").
											Class("syncTextStyle").
											Style("resize", "none").
											Rows(5).
											Cols(50).
											Wrap("off").
											Placeholder("long url or data to shorten..."),
									),
							).Else(
								app.Div().
									Class("container-fluid").
									Class("shortOutput").
									Body(
										app.Div().
											Class("row").
											Body(
												app.Div().
													Class("form-group").
													Class("input-group").
													Body(
														app.Span().
															Class("input-group-addon", "fld-title").
															// Styles(map[string]string{
															// 	"float": "left",
															// 	"width": "12%"}).
															Body(
																app.Text("public"),
															),
														app.Input().
															ID("short-public").
															Type("text").
															Class("form-control").
															Class("syncTextStyle").
															ReadOnly(true).
															// Styles(map[string]string{
															// 	"float": "center",
															// 	"width": "30%"}).
															Value(h.shortLink("short")),
														app.Span().
															Class("input-group-btn").
															// Styles(map[string]string{
															// 	"float": "center",
															// 	"width": "10%"}).
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
																	}).
																	OnMouseOver(func(ctx app.Context, e app.Event) {
																		if h.debug {
																			elem := app.Window().GetElementByID("messages")
																			elem.Set("innerText", "copy to clipboard")
																		}
																	}).
																	OnMouseOut(func(ctx app.Context, e app.Event) {
																		if h.debug {
																			elem := app.Window().GetElementByID("messages")
																			elem.Set("innerText", "")
																		}
																	}),
															),
													),
											),
										app.Div().
											Class("row").
											Body(
												app.Div().
													Class("form-group").
													Class("input-group").
													Body(
														app.Span().
															Class("input-group-addon", "fld-title").
															// Styles(map[string]string{
															// 	"float": "left",
															// 	"width": "12%"}).
															Body(
																app.Text("private"),
															),
														app.Input().
															ID("short-private").
															Type("text").
															Class("form-control").
															Class("syncTextStyle").
															ReadOnly(true).
															// Styles(map[string]string{
															// 	"float": "center",
															// 	"width": "30%"}).
															Value(h.shortLink("private")),

														app.Span().
															Class("input-group-btn").
															// Styles(map[string]string{
															// 	"float": "center",
															// 	"width": "10%"}).
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
											Class("row").
											Body(
												app.Div().
													Class("form-group").
													Class("input-group").
													Body(
														app.Span().
															Class("input-group-addon", "fld-title").
															// Styles(map[string]string{
															// 	"float": "left",
															// 	"width": "12%"}).
															Body(
																app.Text("delete"),
															),
														app.Input().
															ID("short-delete").
															Type("text").
															Class("form-control").
															Class("syncTextStyle").
															ReadOnly(true).
															// Styles(map[string]string{
															// 	"float": "center",
															// 	"width": "30%"}).
															Value(h.shortLink("delete")),
														app.Span().
															Class("input-group-btn").
															// Styles(map[string]string{
															// 	"float": "center",
															// 	"width": "10%"}).
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
						Class("col-xs-8", "col-xs-offset-2").
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
								),
							app.Div().
								Class("shortButtonPost"),
						),
				),
			app.Div().
				Class("row").
				Class("marker").
				ID("mainOptions"),
			app.Div().
				Class("row").
				Class("shortOptionsWrapper").
				Body(
					app.Div().
						Class("container-fluid").
						Class("shortOptions").
						Body(
							app.Div().
								ID("shortOptionsTitleWrapper").
								Class("row").
								Body(
									app.Div().
										Class("col-md-offset-2", "col-md-4", "col-sm-4", "col-sm-offset-2", "col-xs-4", "col-xs-offset-2").
										Body(
											app.H3().
												Body(
													app.Text("Options"),
												),
										),
								),
							app.Div().
								ID("shortOption1Wrapper").
								Class("row").
								Body(
									app.Div().
										Class("form-group").
										Class("col-md-offset-2", "col-md-6", "col-sm-offset-2", "col-sm-6", "col-xs-offset-2", "col-xs-6").
										Body(
											app.Div().
												Class("input-group").
												Class(func() string {
													if h.isShortAsData {
														return "has-success"
													}
													return "has-warning"
												}()).
												Title("use input as data").
												ID("shortAsUrl").
												Body(
													app.Div().
														Class("input-group-addon").
														Body(
															app.Label().
																Body(
																	app.Input().
																		Type("checkbox").
																		Value("").
																		OnClick(func(ctx app.Context, e app.Event) {
																			h.isShortAsData = ctx.JSSrc().Get("checked").Bool()
																		}),
																),
														),
													app.If(h.isShortAsData,
														app.Input().
															Class("form-control").
															Class("syncTextStyle").
															ReadOnly(true).Value("Input treated as data"),
													).Else(
														app.Input().Class("form-control").ReadOnly(true).Value("Automatic treat input as data or Url"),
													),
												),
										),
								),
							app.Div().
								ID("shortOption2Wrapper").
								Class("row").
								Body(
									app.Div().
										Class("form-group").
										Class("col-md-offset-2", "col-md-6", "col-sm-offset-2", "col-sm-6", "col-xs-offset-2", "col-xs-6").
										Body(
											app.Div().
												ID("shortExpire").
												Class("input-group").
												Class(func() string {
													if h.isExpireChecked {
														return "has-success"
													}
													return "has-warning"
												}()).
												Title("Set expiration for the short url").
												Body(
													app.Div().
														Class("input-group-addon").
														Body(
															app.Label().
																Body(
																	app.Input().
																		Type("checkbox").
																		Value("").
																		OnClick(func(ctx app.Context, e app.Event) {
																			h.isExpireChecked = ctx.JSSrc().Get("checked").Bool()
																		}),
																),
														),
													app.If(h.isExpireChecked,
														app.Div().Class("input-group-addon").Body(
															app.Label().Body(
																app.Text("Expiration"),
															),
														),
														app.Div().Dir("ltr").Body(
															app.Select().Class("input-group-addon").Class("form-control").ID("expireSelect").Body(
																app.Option().
																	Value("10m").
																	Body(app.Text("10 minutes")),
																app.Option().
																	Value("12h").
																	Selected(true).
																	Body(app.Text("12 hours")),
																app.Option().
																	Value("2d").
																	Body(app.Text("2 days")),
																app.Option().
																	Value("2w").
																	Body(app.Text("2 weeks")),
																app.Option().
																	Value("8w").
																	Body(app.Text("2 months")),
																app.Option().
																	Value("2y").
																	Body(app.Text("year")),
																app.Option().
																	Value("n").
																	Body(app.Text("never")),
															).OnChange(func(ctx app.Context, e app.Event) {
																h.expireValue = ctx.JSSrc().Get("value").String()
																fmt.Printf("select change value: %v\n", h.expireValue)
															}),
														),
														app.Div().Class("input-group-addon"),
													).Else(
														app.Input().Class("form-control").ReadOnly(true).Value("Default expiration (12 hours)"),
													),
												),
										),
								),
						),
				),
			app.Div().
				Class("row").
				Class("marker").
				ID("footer"),
			app.Div().
				Class("row").
				Class("footer").
				Body(
					app.Div().
						Class("col-xs-8 col-xs-offset-2").
						Body(
							app.Textarea().
								Class("syncTextStyle").
								ID("footerText"),
						),
				),
		)
}

func newShort() *short {
	return &short{}
}

func (h *short) OnInit() {
	if strings.Contains(app.Window().URL().Path, webappCommon.PrivatePath) {
		h.isPrivate = true
	} else {
		h.getStID()
	}
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
	fmt.Printf("!!URL: %+#v\n", app.Window().URL())
	appUrl := app.Window().URL()
	dest := url.URL{
		Scheme: appUrl.Scheme,
		Host:   appUrl.Host,
	}
	elem := app.Window().GetElementByID("shortInputText")
	data := elem.Get("value").String()
	errElem := app.Window().GetElementByID("footerText")
	destCreate := dest.String()
	fmt.Printf("!!! new dest: %s\n", destCreate)
	payload := []byte(data)

	if url, ok := urlCheck(data); ok && !h.isShortAsData {

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
	if h.expireValue != "" {
		req.Header.Set("exp", h.expireValue)
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
	newURL := url.URL{
		Scheme: x.Scheme,
		Host:   x.Host,
		Path:   "/",
	}
	fmt.Printf("!# path: %#+v\n", x)
	host := newURL.String()
	switch which {
	case "private", "short", "delete":
	default:
		// error
	}
	return host + h.resultMap[which]
}

func (h *short) copyToClipboard(from string) {
	elem := app.Window().GetElementByID(from)
	if !app.Window().Get("window").Get("isSecureContext").Bool() {
		// https://stackoverflow.com/questions/51805395/navigator-clipboard-is-undefined
		fmt.Printf("!! cant copy to clipboard using navigator on non secure origin, use execCommand")

		// https://web.dev/async-clipboard/
		elem.Call("select")
		app.Window().Get("document").Call("execCommand", "copy")
		return
	}
	if clipboard := app.Window().Get("navigator").Get("clipboard"); !clipboard.IsUndefined() {
		clipboard.Call("writeText", elem.Get("value"))
	}
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
	preSeed := shortener.GenerateTokenTweaked(uuid.NewString(), -1, 20, 0)
	req.Header.Set("RTS", preSeed)
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

	token := shortener.GenerateTokenTweaked(ua+seed, tokenStartPos, tokenLen, 0)
	fmt.Printf("******************************* calculated token: %s\n", token)
	if token == "" {
		app.Logf("problem with token generation\n")
		return
	}
	h.token = token

	if resp.Header.Get("debug") == "on" {
		h.debug = true
		tokData := fmt.Sprintf("preSeed: %s\nseed: %s\ntoken: %s\n", preSeed, seed, token)
		go func() {
			<-time.After(50 * time.Millisecond)
			elem := app.Window().GetElementByID("messages")
			elem.Set("innerText", tokData)
		}()
	}
}

func (h *short) getPrivateInfo() (map[string]string, []string, error) {

	var err error
	url := app.Window().URL()
	url.Path = "/" + url.Query().Get("key") + "/info"
	url.RawQuery = ""

	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		app.Logf("failed to create new request: %s\n", err)
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	resp, err := client.Do(req)
	if err != nil {
		app.Logf("failed to invoke request: %s\n", err)
		return nil, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		app.Logf("response not ok: %s\n", resp.StatusCode)
		return nil, nil, fmt.Errorf("status: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Logf("failed to read response body: %s\n", err)
		return nil, nil, err
	}
	tup, err := store.NewTupleFromString(string(body))
	if err != nil {
		app.Logf("failed to parse body: %s\n", err)
		return nil, nil, err
	}
	r := map[string]string{}
	var tc time.Time
	for _, k := range tup.Keys() {
		if strings.Contains(k, store.FieldDATA) {
			r[k] = tup.MustGet(k)
		} else {
			r[k] = tup.Get(k)
		}
		k2 := k
		if strings.HasPrefix(k2, store.FieldModTime) {
			k2 = strings.TrimSuffix(k, strings.TrimPrefix(k, store.FieldModTime))
		}
		// k2 := strings.Split(k, "_")[0]
		// r[k], err = tup.Get2(k)
		// if err != nil {
		// 	r[k] = tup.Get(k)
		// }
		switch k2 {
		case "p":
			// drop it
			delete(r, k)
		case "s":
			r["short"] = r[k]
			delete(r, k)
		case "d":
			r["delete"] = r[k]
			delete(r, k)
		case store.FieldTime, store.FieldModTime:
			tc, _ = time.Parse(time.RFC3339, r[k])
			r[k] = tc.String()
		}
	}
	if v, ok := r[store.FieldTTL]; ok {
		d, _ := time.ParseDuration(v)
		r["expire"] = tc.Add(d).String()
	}
	order := []string{}
	for k := range r {
		if k == store.FieldDATA {
			continue
		}
		order = append(order, k)
	}
	sort.Strings(order)
	return r, order, nil
}
