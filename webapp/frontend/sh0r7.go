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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type short struct {
	app.Compo
	result                 string
	resultMap              map[string]string
	resultReady            bool
	sessionToken           string // token used for the loaded session - life time of <SH0R7_WEBAPP_TOKEN_EXPIRATION_SHORT_LIVE>
	expireValue            string
	debug                  bool
	isPrivate              bool   // indicate short url is private
	isPublic               bool   // indicate short url is public
	isRemove               bool   // indicate short url is remove
	isShortAsData          bool   // indicate whether the input should be treated as data and not auto identify - option 1
	isExpireChecked        bool   // indicate whether the expiration feature is used - option 2
	isDescription          bool   // indicate whether the description feature is used when creating short - option 3
	isOptionPrivate        bool   // indicate whether the short private link feature is used when creating short - option 8
	isPrivatePassword      bool   // indicate whether the private password feature is used when creating short - option 4
	isPrivatePasswordShown bool   // indicate whether the private password is shown or hidden - sub option for option 4
	isPublicPassword       bool   // indicate whether the public password feature is used when creating short - option 6
	isPublicPasswordShown  bool   // indicate whether the public password is shown or hidden - sub option for option 6
	isOptionRemove         bool   // indicate whether the short remove link feature is used when creating short - option 5
	isRemovePassword       bool   // indicate whether the remove password feature is used when creating short - option 7 (applicable after option 5 is enabled)
	isRemovePasswordShown  bool   // indicate whether the remove password is shown or hidden - sub option for option 7 (applicable after option 5 is enabled)
	isResultLocked         bool   // indicate that the requested short is password locked
	privatePassSalt        string // salt used for password token - for private link
	publicPassSalt         string // salt used for password token - for public link
	removePassSalt         string // salt used for password token - for remove link
	passToken              string // the password token used to lock and unlock the short private
	updateAvailable        bool   // new version available
}

const (
	NOTEMESSAGE = "Sh0r7 service is still in alpha!"
)

var (
	ImgSource = "/web/logo.jpg"
	// imgSource: "logoL.png",
	BuildVer  string = "dev"
	BuildTime string = "now"
)

func (h *short) RenderPrivate() app.UI {
	var err error

	out := map[string]string{
		"Lorem":       "ipsum dolor sit amet",
		"consectetur": "adipiscing elit",
		"sed":         "do eiusmod tempor incididunt ut labore et dolore magna aliqua",
		"Ut":          "enim ad minim veniam",
		"quis":        "nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat",
		"Duis":        "aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur",
		"Excepteur":   "sint occaecat cupidatat non proident",
		"sunt":        "in culpa qui officia deserunt mollit anim id est laborum",
	}
	keys := []string{"Lorem", "consectetur", "sed", "Ut", "quis", "Duis", "Excepteur", "sunt"}
	tableID := "tableInfo"
	if h.isResultLocked {
		tableID = "tableLocked"
	} else {
		out, keys, err = h.getPrivateInfo(h.passToken)
		if err != nil {
			app.Logf("error getting private info (%s)\n", err)
			out = map[string]string{"error": err.Error()}
			keys = []string{"error"}
		}
	}
	return app.Div().
		Class("container").
		Body(
			h.getTitleHeader(),
			app.Div().
				Class("row").
				Body(
					app.Div().
						Class("col-xs-8", "col-xs-offset-2").
						Body(
							app.H2().
								Body(app.Text("private details")),
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
									app.Text(app.Window().URL().Query().Get(webappCommon.FPrivateKey)),
								),
							app.Br(),
						),
				),
			app.Div().
				Class("row").
				Body(
					app.If(h.isResultLocked,
						app.Div().
							ID("lockedPassword").
							Class().
							Body(
								app.Form().
									Class("form-inline").
									Body(
										app.Div().
											Class("form-group").
											Body(
												app.Input().
													ID("resultUserPassword").
													Class("form-control").
													Type("password").
													Placeholder("Password"),
											),

										app.Button().
											Title("Unlock the private short info").
											ID("unlockButton").
											Class("btn", "btn-default").
											Type("button").
											Body(
												app.Text("Unlock"),
											).
											OnClick(func(ctx app.Context, e app.Event) {
												elem := app.Window().GetElementByID("resultUserPassword")
												v := elem.Get("value").String()
												h.passToken = shortener.GenerateTokenTweaked(v+h.privatePassSalt, 0, 30, 10)
												h.isResultLocked = false
												h.Update()
											}),
									),
							),
					),
					app.Div().
						Class("col-xs-8", "col-xs-offset-1").
						Body(
							app.Table().
								ID(tableID).
								Class("table", "table-hover").
								Body(
									app.TBody().
										Body(
											app.Range(keys).Slice(func(i int) app.UI {
												s := keys[i]
												if s == store.FieldDesc {
													go func() {
														<-time.After(50 * time.Millisecond)
														elem := app.Window().GetElementByID("privateTitle")
														elem.Set("innerText", out[s])
													}()
												}
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
																		app.If(s == "error",
																			app.Div().
																				Class("1input-group", "has-error").
																				Body(
																					app.Input().
																						Class("form-control", "errorTextStyle").
																						Value(out[s]).
																						ReadOnly(true),
																				),
																		).ElseIf(s == "data",
																			app.Div().
																				Class("1input-group", "has-success").
																				Body(
																					app.Textarea().
																						ID("").
																						Class("form-control").
																						Class("syncTextStyle").
																						Style("resize", "none").
																						Wrap("off").Body(app.Text(out[s])),
																				),
																		).Else(
																			app.Div().
																				Class("1input-group", "has-success").
																				Body(
																					app.Input().
																						Class("form-control", "syncTextStyle").
																						Value(func() any {
																							switch s {
																							case store.FieldPrivate, store.FieldPublic, store.FieldRemove:
																								return h.shortLink(s, out)
																							default:
																								return out[s]
																							}
																						}()).
																						ReadOnly(true),
																				),
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

func (h *short) RenderPublic() app.UI {
	if !h.isResultLocked {
		app.Logf("triggering getPublicShort with passtoekn : <%s>\n", h.passToken)
		out, _, err := h.getPublicShort(h.passToken)
		app.Logf("getpublic short result: <%v>\n", out)
		if err != nil {
			app.Logf("error getting public data (%s)\n", err)
		} else {
			if _, ok := out[store.FieldURL]; ok {
				app.Window().Get("location").Set("href", out[store.FieldURL])
				return app.Main().Body(app.Div().Class().Body(app.Text("...")))
			} else {
				return app.Div().
					Body(
						app.Text(out[store.FieldDATA]),
					)
			}
		}
	}
	return app.Div().
		Class("container").
		Body(
			app.If(h.isResultLocked,
				app.Div().
					ID("lockedPassword").
					Class().
					Body(
						app.Form().
							Class("form-inline").
							Body(
								app.Div().
									Class("form-group").
									Body(
										app.Input().
											ID("resultUserPassword").
											Class("form-control").
											Type("password").
											Placeholder("Password"),
									),

								app.Button().
									Title("Unlock the public short").
									ID("unlockButton").
									Class("btn", "btn-default").
									Type("button").
									Body(
										app.Text("Unlock"),
									).
									OnClick(func(ctx app.Context, e app.Event) {
										elem := app.Window().GetElementByID("resultUserPassword")
										v := elem.Get("value").String()
										h.passToken = shortener.GenerateTokenTweaked(v+h.publicPassSalt, 0, 30, 10)
										h.isResultLocked = false
										h.Update()
									}),
							),
					),
			).Else(
				app.Div().
					Class("passwordError").
					Body(
						app.Text("Unlock failed"),
						app.Button().
							Title("Retry").
							ID("").
							Class("btn", "btn-default").
							Type("button").
							Body(
								app.Text("Retry"),
							).
							OnClick(func(ctx app.Context, e app.Event) {
								retryUrl, _ := url.ParseRequestURI(app.Window().URL().String())
								retryUrl.Path = retryUrl.Query().Get(webappCommon.FPrivateKey)
								retryUrl.RawQuery = ""
								app.Window().Get("location").Set("href", retryUrl.String())
							}),
					),
			),
		)
}

func (h *short) RenderRemove() app.UI {
	if !h.isResultLocked {
		app.Logf("triggering getRemoveShort with passtoekn : <%s>\n", h.passToken)
		out, _, err := h.getRemoveShort(h.passToken)
		app.Logf("getRemoveshort result: <%v>\n", out)
		if err != nil {
			app.Logf("error getting remove data (%s)\n", err)
		} else {
			return app.Div().
				Body(
					app.Text("short removed"),
				)
		}
	}
	return app.Div().
		Class("container").
		Body(
			app.If(h.isResultLocked,
				app.Div().
					ID("lockedPassword").
					Class().
					Body(
						app.Form().
							Class("form-inline").
							Body(
								app.Div().
									Class("form-group").
									Body(
										app.Input().
											ID("resultUserPassword").
											Class("form-control").
											Type("password").
											Placeholder("Password"),
									),

								app.Button().
									Title("Unlock the remove short").
									ID("unlockButton").
									Class("btn", "btn-default").
									Type("button").
									Body(
										app.Text("Unlock"),
									).
									OnClick(func(ctx app.Context, e app.Event) {
										elem := app.Window().GetElementByID("resultUserPassword")
										v := elem.Get("value").String()
										h.passToken = shortener.GenerateTokenTweaked(v+h.removePassSalt, 0, 30, 10)
										h.isResultLocked = false
										h.Update()
									}),
							),
					),
			).Else(
				app.Div().
					Class("passwordError").
					Body(
						app.Text("Unlock failed"),
						app.Button().
							Title("Retry").
							ID("").
							Class("btn", "btn-default").
							Type("button").
							Body(
								app.Text("Retry"),
							).
							OnClick(func(ctx app.Context, e app.Event) {
								retryUrl, _ := url.ParseRequestURI(app.Window().URL().String())
								retryUrl.Path = retryUrl.Query().Get(webappCommon.FPrivateKey)
								retryUrl.RawQuery = ""
								app.Window().Get("location").Set("href", retryUrl.String())
							}),
					),
			),
		)
}

func (h *short) Render() app.UI {
	if h.updateAvailable {
		return h.RenderUpdate()
	}
	if h.isPrivate {
		return h.RenderPrivate()
	}
	if h.isRemove {
		return h.RenderRemove()
	}
	if h.isPublic {
		return h.RenderPublic()
	}
	return app.Main().Body(app.Div().
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
			h.getTitleHeader(),
			app.Div().
				Class("row").
				Class("marker").
				ID("mainDo"),
			app.Div().
				Class("row").
				Class("shortDo").
				Body(
					app.Div().
						Class("col-xs-10", "col-xs-offset-1", "shortInputWrapper").
						Class("shortInputWrapper").
						Body(
							app.If(!h.resultReady,
								app.Div().
									Class("shortInput").
									Body(
										app.Textarea().
											ID("shortInputText").
											Class("form-control").
											Class("form-group").
											Class("syncTextStyle").
											Style("resize", "none").
											Rows(5).
											Cols(50).
											Wrap("off").
											Placeholder("long url or data to shorten..."),
									),
							).Else(
								h.shortCreationOutput(),
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
												app.Logf("shortInputText value: %v\n", v)
												h.expireValue = ""
												if h.isExpireChecked {
													h.expireValue = app.Window().GetElementByID("expireSelect").Get("value").String()
													app.Logf("expire value: %v\n", h.expireValue)
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
												// reset the options
												h.isShortAsData = false
												h.isExpireChecked = false
												h.isDescription = false
												h.isPublicPassword = false
												h.isPublicPasswordShown = false
												app.Window().GetElementByID("checkboxShortAsData").Set("checked", false)
												app.Window().GetElementByID("checkboxExpire").Set("checked", false)
												app.Window().GetElementByID("checkboxDescription").Set("checked", false)
												if h.isOptionPrivate {
													app.Window().GetElementByID("checkboxPrivatePassword").Set("checked", false)
													app.Window().GetElementByID("checkboxOptionPrivate").Set("checked", false)
													h.isOptionPrivate = false
													h.isPrivatePassword = false
													h.isPrivatePasswordShown = false
												}
												app.Window().GetElementByID("checkboxPublicPassword").Set("checked", false)
												if h.isOptionRemove {
													app.Window().GetElementByID("checkboxRemovePassword").Set("checked", false)
													app.Window().GetElementByID("checkboxOptionRemove").Set("checked", false)
													h.isOptionRemove = false
													h.isRemovePassword = false
													h.isRemovePasswordShown = false
												}
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
							h.OptionsTitle(),
							h.OptionShortAsData(),
							h.OptionExpire(),
							h.OptionDescription(),
							h.OptionPublic(),
							h.OptionPrivate(),
							h.OptionRemove(),
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
		))
}

func newShort() *short {
	return &short{}
}

func (h *short) load() {
	lurl := app.Window().URL()
	app.Logf("url: %#+v\n", lurl)
	if strings.Contains(lurl.Path, webappCommon.PrivatePath) && lurl.Query().Has("key") {
		h.isPrivate = true
		if lurl.Query().Has(webappCommon.PasswordProtected) {
			h.privatePassSalt = lurl.Query().Get(webappCommon.PasswordProtected)
			h.isResultLocked = true
		}
	} else if strings.Contains(lurl.Path, webappCommon.RemovePath) && lurl.Query().Has("key") {
		h.isRemove = true
		if lurl.Query().Has(webappCommon.PasswordProtected) {
			h.removePassSalt = lurl.Query().Get(webappCommon.PasswordProtected)
			h.isResultLocked = true
		}
	} else if strings.Contains(lurl.Path, webappCommon.PublicPath) && lurl.Query().Has("key") {
		h.isPublic = true
		if lurl.Query().Has(webappCommon.PasswordProtected) {
			h.publicPassSalt = lurl.Query().Get(webappCommon.PasswordProtected)
			h.isResultLocked = true
		}
	} else {
		h.getStID()
	}
	app.Logf("******************************* init")
}
func (h *short) load2() {
	lurl := app.Window().URL()
	app.Logf("url: %#+v\n", lurl)
	if strings.Contains(lurl.Path, webappCommon.PrivatePath) && lurl.Query().Has("key") {
		h.isPrivate = true
		if lurl.Query().Has(webappCommon.PasswordProtected) {
			h.privatePassSalt = lurl.Query().Get(webappCommon.PasswordProtected)
			h.isResultLocked = true
		}
	} else if strings.Contains(lurl.Path, webappCommon.RemovePath) && lurl.Query().Has("key") {
		h.isRemove = true
		if lurl.Query().Has(webappCommon.PasswordProtected) {
			h.removePassSalt = lurl.Query().Get(webappCommon.PasswordProtected)
			h.isResultLocked = true
		}
	} else if strings.Contains(lurl.Path, webappCommon.PublicPath) && lurl.Query().Has("key") {
		h.isPublic = true
		if lurl.Query().Has(webappCommon.PasswordProtected) {
			h.publicPassSalt = lurl.Query().Get(webappCommon.PasswordProtected)
			h.isResultLocked = true
		}
	} else {
		if lurl.Query().Has(webappCommon.FSaltTokenID) {
			qVals := lurl.Query()
			stid, ok := qVals[webappCommon.FSaltTokenID]
			if !ok {
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

			ua := app.Window().Get("navigator").Get("userAgent").String()

			token := shortener.GenerateTokenTweaked(ua+seed, tokenStartPos, tokenLen, 0)
			if token == "" {
				app.Logf("problem with token generation\n")
				return
			}
			h.sessionToken = token
		}
	}
	app.Logf("load2....\n")
}
func (h *short) OnInit() {
	h.load2()
	app.Logf("******************************* init - build ver :<%s>, time: <%s>\n", BuildVer, BuildTime)
}
func (h *short) OnPreRender(ctx app.Context) {
	h.load()
	app.Logf("******************************* prerender")
}
func (h *short) OnDisMount() {
	app.Logf("******************************* dismount")
}
func (h *short) OnMount() {
	app.Logf("******************************* mount")
}
func (h *short) OnNav() {
	h.load()
	app.Logf("******************************* nav")
}
func (h *short) OnResize() {
	h.ResizeContent()
	app.Logf("******************************* update")
}
func (h *short) OnUpdate() {
	app.Logf("******************************* update")
}
func (h *short) OnAppUpdate(ctx app.Context) {
	h.updateAvailable = ctx.AppUpdateAvailable()
	app.Logf("******************************* app update: %v\n", h.updateAvailable)
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
	app.Logf("!!URL: %+#v\n", app.Window().URL())
	appUrl := app.Window().URL()
	dest := url.URL{
		Scheme: appUrl.Scheme,
		Host:   appUrl.Host,
	}
	elem := app.Window().GetElementByID("shortInputText")
	data := elem.Get("value").String()
	errElem := app.Window().GetElementByID("footerText")
	destCreate := dest.String()
	app.Logf("!!! new dest: %s\n", destCreate)
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
	// app.Logf("app %#v\n", app.)
	req, err := http.NewRequest(http.MethodPost, destCreate, bytes.NewBuffer(payload))
	if err != nil {
		errElem.Set("value", fmt.Sprintf("new request: error occurred: %s", err))
		return
	}
	if eDesc := app.Window().GetElementByID("shortDescription"); !eDesc.IsNull() {
		if desc := eDesc.Get("value").String(); desc != "" {
			req.Header.Set(webappCommon.FShortDesc, desc)
		}
	}
	if ePrvPass := app.Window().GetElementByID("privatePasswordText"); !ePrvPass.IsNull() {
		if prvPass := ePrvPass.Get("value").String(); prvPass != "" {
			req.Header.Set(webappCommon.FPrvPassToken, shortener.GenerateTokenTweaked(prvPass+h.sessionToken, 0, 30, 10))
		}
	}
	if ePubPass := app.Window().GetElementByID("publicPasswordText"); !ePubPass.IsNull() {
		if pubPass := ePubPass.Get("value").String(); pubPass != "" {
			req.Header.Set(webappCommon.FPubPassToken, shortener.GenerateTokenTweaked(pubPass+h.sessionToken, 0, 30, 10))
		}
	}
	if eRemPass := app.Window().GetElementByID("removePasswordText"); !eRemPass.IsNull() {
		if remPass := eRemPass.Get("value").String(); remPass != "" {
			req.Header.Set(webappCommon.FRemPassToken, shortener.GenerateTokenTweaked(remPass+h.sessionToken, 0, 30, 10))
		}
	}
	if h.expireValue != "" {
		req.Header.Set(webappCommon.FExpiration, h.expireValue)
	}
	if !h.isOptionPrivate {
		req.Header.Set(webappCommon.FPrivate, "false")
	}
	if !h.isOptionRemove {
		req.Header.Set(webappCommon.FRemove, "false")
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set(webappCommon.FTokenID, h.sessionToken)
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

	app.Logf("******************************* create short result: %s\n", string(body))
	elem.Set("value", string(body))
	app.Logf("******************************* create shoty: %#v\n", r)
	h.Update()
}

func (h *short) shortLink(which string, from map[string]string) string {
	x := app.Window().URL()
	newURL := url.URL{
		Scheme: x.Scheme,
		Host:   x.Host,
		Path:   "/",
	}
	// app.Logf("!# path: %#+v\n", x)
	host := newURL.String()
	switch which {
	case store.FieldPrivate, store.FieldPublic, store.FieldRemove:
	default:
		app.Logf("field <%s> not handled\n", which)
		// error
	}
	return host + from[which]
}

func (h *short) copyToClipboard(from string) {
	elem := app.Window().GetElementByID(from)
	if !app.Window().Get("window").Get("isSecureContext").Bool() {
		// https://stackoverflow.com/questions/51805395/navigator-clipboard-is-undefined
		app.Logf("!! cant copy to clipboard using navigator on non secure origin, use execCommand")

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
	req.Header.Set(webappCommon.FRequestTokenSeed, preSeed)
	resp, err := client.Do(req)
	if err != nil {
		app.Logf("failed to invoke request: %s\n", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		app.Logf("response not ok: %v\n", resp.StatusCode)
		return
	}
	_stid := resp.Header.Get(webappCommon.FSaltTokenID)
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

	app.Logf("******************************* stid from header: %+#v\n", stid)
	ua := app.Window().Get("navigator").Get("userAgent").String()

	token := shortener.GenerateTokenTweaked(ua+seed, tokenStartPos, tokenLen, 0)
	app.Logf("******************************* calculated token: %s\n", token)
	if token == "" {
		app.Logf("problem with token generation\n")
		return
	}
	h.sessionToken = token

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

func (h *short) getPublicShort(passToken string) (map[string]string, []string, error) {

	var err error
	url := app.Window().URL()
	url.Path = "/" + url.Query().Get(webappCommon.FPrivateKey)
	url.RawQuery = ""

	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		app.Logf("failed to create new request: %s\n", err)
		return nil, nil, err
	}
	if passToken != "" {
		req.Header.Set(webappCommon.FPubPassToken, passToken)
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("xRedirect", "no")
	app.Logf("invoking request: %+#v\n", req)
	resp, err := client.Do(req)
	if err != nil {
		app.Logf("failed to invoke request: %s\n", err)
		return nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		app.Logf("response not ok: %v\n", resp.StatusCode)
		return nil, nil, fmt.Errorf("status: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Logf("failed to read response body: %s\n", err)
		return nil, nil, err
	}
	app.Logf("resp body: <%s>\n", string(body))
	tup, err := store.NewTupleFromString(string(body))
	if err != nil {
		app.Logf("failed to parse body: %s\n", err)
		return map[string]string{
			store.FieldDATA: string(body),
		}, []string{store.FieldDATA}, nil
	}
	return map[string]string{
		store.FieldURL: tup.Get(store.FieldURL),
	}, []string{store.FieldURL}, nil
}
func (h *short) getRemoveShort(passToken string) (map[string]string, []string, error) {

	var err error
	url := app.Window().URL()
	url.Path = "/" + url.Query().Get(webappCommon.FPrivateKey)
	url.RawQuery = ""

	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		app.Logf("failed to create new request: %s\n", err)
		return nil, nil, err
	}
	if passToken != "" {
		req.Header.Set(webappCommon.FRemPassToken, passToken)
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("xRedirect", "no")
	app.Logf("invoking request: %+#v\n", req)
	resp, err := client.Do(req)
	if err != nil {
		app.Logf("failed to invoke request: %s\n", err)
		return nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		app.Logf("response not ok: %v\n", resp.StatusCode)
		return nil, nil, fmt.Errorf("status: %v", resp.StatusCode)
	}

	return nil, nil, nil
}

func (h *short) getPrivateInfo(passToken string) (map[string]string, []string, error) {

	var err error
	url := app.Window().URL()
	url.Path = "/" + url.Query().Get(webappCommon.FPrivateKey) + "/info"
	url.RawQuery = ""

	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		app.Logf("failed to create new request: %s\n", err)
		return nil, nil, err
	}
	if passToken != "" {
		req.Header.Set(webappCommon.FPrvPassToken, passToken)
	}
	req.Header.Set("Content-Type", "text/plain")
	resp, err := client.Do(req)
	if err != nil {
		app.Logf("failed to invoke request: %s\n", err)
		return nil, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		app.Logf("response not ok: %v\n", resp.StatusCode)
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
		// case "p":
		// 	// drop it
		// 	delete(r, k)
		// case "s":
		// 	r["short"] = r[k]
		// 	delete(r, k)
		// case "d":
		// 	r["delete"] = r[k]
		// 	delete(r, k)
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
	order = append(order, store.FieldDATA)
	return r, order, nil
}

func (h *short) getTitleHeader() app.UI {
	return app.Div().
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
						Width(200).
						OnClick(func(ctx app.Context, e app.Event) {
							url := app.Window().URL()
							url.Path = webappCommon.ShortPath
							url.RawQuery = ""
							app.Window().Get("location").Set("href", url.String())
						}),
				),
			app.Div().
				Class("col-md-6", "col-md-offset-0", "col-sm-4", "col-sm-offset-0", "col-xs-6", "col-xs-offset-3").
				Class("text").
				Body(
					app.H1().
						Body(
							app.Text("Sh0r7"),
						).
						OnClick(func(ctx app.Context, e app.Event) {
							url := app.Window().URL()
							url.Path = webappCommon.ShortPath
							url.RawQuery = ""
							app.Window().Get("location").Set("href", url.String())
						}),
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
		)
}

func (h *short) passwordOption(isPassword, isPasswordShown *bool, which string) app.HTMLDiv {
	whichTitle := cases.Title(language.English).String(which)

	switch which {
	case "public":
		isPassword = &h.isPublicPassword
		isPasswordShown = &h.isPublicPasswordShown
	case "private":
		isPassword = &h.isPrivatePassword
		isPasswordShown = &h.isPrivatePasswordShown
	case "remove":
		isPassword = &h.isRemovePassword
		isPasswordShown = &h.isRemovePasswordShown
	}
	return app.Div().
		Class("form-group").
		Class("col-md-offset-2", "col-md-6", "col-sm-offset-2", "col-sm-6", "col-xs-offset-1", "col-xs-10").
		Body(
			app.Div().
				Class("input-group").
				Class(func() string {
					if *isPassword {
						return "has-success"
					}
					return "has-warning"
				}()).
				Title("limit access to "+which+" link with password").
				ID(which+"AccessPassword").
				Body(
					app.Label().
						Class("input-group-addon").
						Body(
							app.Input().
								Type("checkbox").
								ID("checkbox"+whichTitle+"Password").
								Value("").
								OnClick(func(ctx app.Context, e app.Event) {
									elem := ctx.JSSrc()
									app.Logf("checkbox element: <%s>\n", elem.Get("id").String())
									*isPassword = ctx.JSSrc().Get("checked").Bool()
									app.Logf("chkbox: setting %s to %v\n", which, *isPassword)
								}),
						),
					app.If(*isPassword,
						app.Div().Class("input-group-addon").Body(
							app.Label().Body(
								app.Text(whichTitle+" password"),
							).OnClick(func(ctx app.Context, e app.Event) {
								elem := app.Window().GetElementByID("checkbox" + whichTitle + "Password")
								elem.Set("checked", false)
								*isPassword = false
								app.Logf("password addon: setting %s to %v\n", which, *isPassword)
							}),
						),
						app.Input().
							Class("form-control").
							Class("syncTextStyle").
							ID(which+"PasswordText").
							Value("").
							ReadOnly(false).Type("password"),
						func() app.UI {
							classIcon := "glyphicon glyphicon-eye-close"
							return app.Label().Class("input-group-addon").
								Body(
									app.Span().
										ID(which + "PasswordReveal").
										Class(classIcon),
								).
								OnClick(func(ctx app.Context, e app.Event) {
									*isPasswordShown = !*isPasswordShown
									inputType := "password"
									if *isPasswordShown {
										inputType = "text"
										classIcon = "glyphicon glyphicon-eye-open"
									}
									el := app.Window().GetElementByID(which + "PasswordText")
									el.Set("type", inputType)
									jui := app.Window().GetElementByID(which + "PasswordReveal")
									attrs := jui.Get("attributes")
									attrs.Set("class", classIcon)
								})
						}(),
					).Else(
						app.Input().Class("form-control").ReadOnly(true).Value("No password on "+which+" link").
							OnClick(func(ctx app.Context, e app.Event) {
								elem := app.Window().GetElementByID("checkbox" + whichTitle + "Password")
								elem.Set("checked", true)
								*isPassword = true
								app.Logf("no password: setting %s to %v\n", which, *isPassword)
							}),
					),
				),
		)
}

func (h *short) RenderUpdate() app.UI {
	return app.Div().
		Class("container").
		Body(
			app.Div().
				ID("updateApp").
				Class().
				Body(
					app.Form().
						Class("form-inline").
						Body(
							app.Button().
								Title("Update webapp").
								ID("updateAppBtn").
								Class("btn", "btn-default").
								Type("button").
								Body(
									app.Text("Update"),
								).
								OnClick(func(ctx app.Context, e app.Event) {
									// Reloads the page to display the modifications.
									ctx.Reload()
								}),
						),
				),
		)
}
