package frontend

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	ecies "github.com/ecies/go/v2"
	"github.com/gilwo/Sh0r7/common"
	"github.com/gilwo/Sh0r7/shortener"
	"github.com/gilwo/Sh0r7/store"
	webappCommon "github.com/gilwo/Sh0r7/webapp/common"
	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type short struct {
	app.Compo
	result                     string
	resultMap                  map[string]string
	resultReady                bool
	sessionToken               string // token used for the loaded session - life time of <SH0R7_WEBAPP_TOKEN_EXPIRATION_SHORT_LIVE>
	expireValue                string
	debug                      bool
	isPrivate                  bool   // indicate short url is private
	isPublic                   bool   // indicate short url is public
	isRemove                   bool   // indicate short url is remove
	isShortAsData              bool   // indicate whether the input should be treated as data and not auto identify - option 1
	publicData                 string // holds retrieved data of public link
	publicUrl                  string // holds retrieved url of public link
	isDataEncryptPassword      bool   // indicate whether the input should be encrypted - option 1.1
	isDataEncryptPasswordShown bool   // indicate whether the encrypt password is shown or hidden - option 1.1.1
	isExpireChecked            bool   // indicate whether the expiration feature is used - option 2
	isDescription              bool   // indicate whether the description feature is used when creating short - option 3
	isOptionPrivate            bool   // indicate whether the short private link feature is used when creating short - option 8
	isPrivatePassword          bool   // indicate whether the private password feature is used when creating short - option 4
	isPrivatePasswordShown     bool   // indicate whether the private password is shown or hidden - sub option for option 4
	isPublicPassword           bool   // indicate whether the public password feature is used when creating short - option 6
	isPublicPasswordShown      bool   // indicate whether the public password is shown or hidden - sub option for option 6
	isOptionRemove             bool   // indicate whether the short remove link feature is used when creating short - option 5
	isRemovePassword           bool   // indicate whether the remove password feature is used when creating short - option 7 (applicable after option 5 is enabled)
	isRemovePasswordShown      bool   // indicate whether the remove password is shown or hidden - sub option for option 7 (applicable after option 5 is enabled)
	isNamedPublic              bool   // indicate whether the named public feature is used when creating short - option 9
	isResultLocked             bool   // indicate that the requested short is password locked
	isResultUnlockFailed       bool   // indicate that the requested short unlock failed (wrong password)
	privatePassSalt            string // salt used for password token - for private link
	publicPassSalt             string // salt used for password token - for public link
	removePassSalt             string // salt used for password token - for remove link
	passToken                  string // the password token used to lock and unlock the short private
	updateAvailable            bool   // new version available

	accountCreateValidContact string // email or phone #
	accountCreatePass         string // password
	accountCreateStore        bool   // should the valid Contact details (email / phone #) be saved in server

	isDev          bool
	isDebug        bool
	isExperimental bool
	isDebugWindow  bool
	isLoaded       bool
}

const (
	// NOTEMESSAGE = "Sh0r7 service is still in alpha!"
	NOTEMESSAGE = "Sh0r7 service is in beta!"
)

var (
	ImgSource = "/web/logo.jpg"
	// imgSource: "logoL.png",
	// BuildVer          string = "dev"
	// BuildTime         string = "now"
	// ExternalTimeBuild string = "now"
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
		h.isResultUnlockFailed = err != nil
	}
	return app.Div().
		Class("container").
		Body(
			h.getTitleHeader(),
			app.Div().
				Class("row").
				Body(
					app.Div().
						Class("col-sm-8", "col-sm-offset-2").
						Body(
							app.H2().
								Body(app.Text("private details")),
						),
				),
			app.Div().
				Class("row").
				Body(
					app.Div().
						Class("col-sm-6", "col-sm-offset-3").
						Body(
							app.H3().
								ID("privateTitle").
								Body(
									// app.Text("using private for "+app.Window().URL().String()),
									app.Text(app.Window().URL().Query().Get(webappCommon.FShortKey)),
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
								app.Div().
									Class().
									Body(app.Text("123"+common.BuildTime)),
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
													Placeholder("Password").
													OnKeyDown(preventEnter).
													OnKeyPress(preventEnter).
													OnKeyUp(preventEnter),
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
												app.Logf("pass text: <%s>\n", v)
												app.Logf("pass salt: <%s>\n", h.privatePassSalt)
												app.Logf("pass token (calc): <%s>\n", h.passToken)
												h.isResultLocked = false
												h.Update()
											}),
									),
							),
					).ElseIf(h.isResultUnlockFailed,
						h.showRetry(),
					),
					app.Div().
						Class("col-sm-8", "col-sm-offset-1").
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
	app.Logf("publicData: <%s>\n", h.publicData)
	if r := h.RenderPublicWithPassword(); r != nil {
		app.Logf("rending publicwith password")
		return r
	}
	if !h.isResultLocked {
		if len(h.publicData+h.publicUrl) == 0 {
			app.Logf("triggering getPublicShort with passtoekn : <%s>\n", h.passToken)
			err := h.getPublicShort(h.passToken)
			// app.Logf("getpublic short result: <%v>\n", out)
			if err != nil {
				app.Logf("error getting public data (%s)\n", err)
			}
		}
		if len(h.publicUrl) > 0 {
			app.Window().Get("location").Set("href", h.publicUrl)
			return app.Main().Body(app.Div().Class().Body(app.Text("...")))
		} else {
			if h.isDataEncryptPassword {
				return h.RenderPublicWithPassword()
			}
			return app.Div().
				Body(
					app.Pre().ID("publicData").
						ContentEditable(false).
						OnContextMenu(func(ctx app.Context, e app.Event) {
							app.Logf("context menu triggered\n")
						}).
						Body(
							app.Text(h.publicData),
						),
					// app.If(h.isDataEncryptPassword,
					// 	app.Div().
					// 		Class().
					// 		Body(
					// 			app.Text("data is encrypted"),
					// 			app.Br(),
					// 			h.passwordOption("encrypt").
					// 				OnChange(func(ctx app.Context, e app.Event) {
					// 					h.skipRender = true
					// 					fmt.Printf("something changed\n")
					// 					dataLoc := app.Window().GetElementByID("publicData")
					// 					if h.isDataEncryptPassword {
					// 						data := dataLoc.Get("innerText").String()
					// 						if data == "error" {
					// 							return
					// 						}
					// 						dataBuf, err := hex.DecodeString(data)
					// 						if err == nil {
					// 							dec, isDec := h.decryptPayload(dataBuf)
					// 							if isDec {
					// 								app.Logf("dec data: <%s>\n", string(dec))
					// 								dataLoc.Set("innerText", string(dec))
					// 							} else {
					// 								app.Logf("failed to decrypt data")
					// 								dataLoc.Set("innerText", "error")
					// 							}
					// 						} else {
					// 							app.Logf("problem with decode string, %s", err)
					// 						}
					// 					} else {
					// 						dataLoc.Set("innerText", out[store.FieldDATA])
					// 					}
					// 				}),
					// 		),
					// 	// Body(
					// 	// 	app.Text("data is encrypted"),
					// 	// 	app.Br(),
					// 	// 	h.passwordOption("encrypt").OnChange(func(ctx app.Context, e app.Event) {
					// 	// 		fmt.Printf("something changed\n")
					// 	// 		dataLoc := app.Window().GetElementByID("publicData")
					// 	// 		data := dataLoc.Get("innerText").String()
					// 	// 		dataBuf, err := hex.DecodeString(data)
					// 	// 		if err != nil {
					// 	// 			app.Logf("problem with decode string, %s", err)
					// 	// 			return
					// 	// 		}
					// 	// 		// if encKey := encryptPass.Get("value").String(); encKey != "" {
					// 	// 		if h.isDataEncryptPassword {
					// 	// 			dec, isDec := h.decryptPayload(dataBuf)
					// 	// 			if !isDec {
					// 	// 				app.Logf("failed to decrypt data")
					// 	// 				return
					// 	// 			}
					// 	// 			app.Logf("dec data: <%s>\n")
					// 	// 			dataLoc.Set("innerText", string(dec))
					// 	// 		}
					// 	// 	}),
					// 	// ),
					// ),
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
						app.Div().
							Class().
							Body(app.Text("##"+common.BuildTime)),
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
											Placeholder("Password").
											OnKeyDown(preventEnter).
											OnKeyPress(preventEnter).
											OnKeyUp(preventEnter),
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
										app.Logf("pass text: <%s>\n", v)
										app.Logf("pass salt: <%s>\n", h.publicPassSalt)
										app.Logf("pass token (calc): <%s>\n", h.passToken)
										h.isResultLocked = false
										h.Update()
									}),
							),
					),
			).Else( // conider use elseif with h.isResultUnlockFailed
				h.showRetry(),
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
											Placeholder("Password").
											OnKeyDown(preventEnter).
											OnKeyPress(preventEnter).
											OnKeyUp(preventEnter),
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
										app.Logf("pass text: <%s>\n", v)
										app.Logf("pass salt: <%s>\n", h.removePassSalt)
										app.Logf("pass token (calc): <%s>\n", h.passToken)
										h.isResultLocked = false
										h.Update()
									}),
							),
					),
			).Else( // conider use elseif with h.isResultUnlockFailed
				h.showRetry(),
			),
		)
}

func (h *short) showRetry() app.UI {
	return app.Div().
		Class("passwordError").
		Class("container-fluid").
		Body(
			app.Div().Class("row").Body(
				app.Text("Unlock failed"),
			),
			app.Div().Class("row").Body(
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
						retryUrl.Path = retryUrl.Query().Get(webappCommon.FShortKey)
						retryUrl.RawQuery = ""
						app.Logf("retry navigate to %s\n", retryUrl.String())
						app.Window().Get("location").Set("href", retryUrl.String())
					}),
			),
		)
}

func (h *short) Render() app.UI {
	// if h.updateAvailable {
	// 	return h.RenderUpdate()
	// }
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
						Class("col-sm-8", "col-sm-offset-2").
						Body(
							app.H4().
								Styles(
									map[string]string{
										"background": "yellow",
										"text-align": "left",
										"width":      "fit-content"}).
								Body(
									app.Text(func() string {
										r := NOTEMESSAGE
										if h.isDev {
											if common.BuildVersion != "" {
												r += " (" + common.BuildVersion + ")"
											}
											t := common.SourceTime
											n, e := strconv.ParseInt(strings.TrimSuffix(t, "*"), 10, 64)
											if e == nil {
												t = time.Unix(n, 0).String()
											}
											r += " " + t
										}
										return r
									}()),
								),
						),
					app.If(h.debug || h.isDev,
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
			// h.navBar2(),
			app.Div().
				Class("row").
				Class("marker").
				ID("navBar"),
			h.navBar3(),
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
						Class("col-sm-10", "col-sm-offset-1", "shortInputWrapper").
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
						Class("col-sm-8", "col-sm-offset-2").
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
													h.createShort(ctx)
													// ctx.Async(h.createShort)
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
												h.isDataEncryptPassword = false
												h.isDataEncryptPasswordShown = false
												h.isExpireChecked = false
												h.isDescription = false
												h.isNamedPublic = false
												h.isPublicPassword = false
												h.isPublicPasswordShown = false
												// app.Window().GetElementByID("checkboxShortAsData").Set("checked", false)
												// app.Window().GetElementByID("checkboxExpire").Set("checked", false)
												// app.Window().GetElementByID("checkboxDescription").Set("checked", false)
												// app.Window().GetElementByID("checkboxNamedPublicShort").Set("checked", false)
												if h.isOptionPrivate {
													// app.Window().GetElementByID("checkboxPrivatePassword").Set("checked", false)
													// app.Window().GetElementByID("checkboxOptionPrivate").Set("checked", false)
													h.isOptionPrivate = false
													h.isPrivatePassword = false
													h.isPrivatePasswordShown = false
												}
												// app.Window().GetElementByID("checkboxPublicPassword").Set("checked", false)
												if h.isOptionRemove {
													// app.Window().GetElementByID("checkboxRemovePassword").Set("checked", false)
													// app.Window().GetElementByID("checkboxOptionRemove").Set("checked", false)
													h.isOptionRemove = false
													h.isRemovePassword = false
													h.isRemovePasswordShown = false
												}
												// h.Update()
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
			app.If(!h.resultReady,
				app.Div().
					Class("row").
					Class("shortOptionsWrapper").
					Body(
						app.Div().
							Class("container-fluid").
							Class("shortOptions").
							Body(
								// func() []app.UI {
								// 	ret := []app.UI{h.OptionsTitle()}
								// 	ret = append(ret, h.OptionShortAsData())
								// 	ret = append(ret, h.OptionExpire())
								// 	ret = append(ret, h.OptionDescription())
								// 	ret = append(ret, h.OptionPublic())
								// 	ret = append(ret, h.OptionPrivate())
								// 	ret = append(ret, h.OptionRemove())
								// 	return ret
								// }()...,
								h.OptionsTitle(),
								h.OptionShortAsData(),
								h.OptionExpire(),
								h.OptionDescription(),
								h.OptionPublic(),
								h.OptionPrivate(),
								h.OptionRemove(),
								app.If(h.isExperimental,
									h.OptionNamedPublicShort(),
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
				Class("container").
				Body(
					app.Div().
						Class("row").
						Body(
							app.Div().
								Class("col-sm-6 col-sm-offset-3").
								Class("text-center").
								Body(
									app.P().
										Body(
											app.Strong().Text("For more information "),
											app.Br(),
											app.A().
												Style("color", "lightgreen").
												Href("mailto:info@sh0r7.me").
												Body(
													app.Text("Contact us"),
												),
										),
								),
						),
					app.If(h.debug || h.isDev,
						app.Div().
							Class("col-sm-8 col-sm-offset-2").
							Body(
								app.Textarea().
									Class("syncTextStyle").
									ID("footerText"),
							),
					),
				),
			app.If(h.isDev,
				func() app.UI {
					h.isDebugWindow = true
					return h.DebugWindow()
				}(),
			),
		))
}

func newShort() *short {
	return &short{}
}

func (h *short) parseSTID(ctx app.Context, stid string) {
	x, err := shortener.Base64SE.Decode(stid)
	if err != nil {
		app.Logf("problem with stid : %s\n", err)
		return
	}
	stidArr := strings.Split(string(x), "$$")
	seed := stidArr[0]
	tokenLen, err := strconv.Atoi(stidArr[1])
	if err != nil {
		app.Logf("problem with number convertion: %s\n", err)
		return
	}
	tokenStartPos, err := strconv.Atoi(stidArr[2])
	if err != nil {
		app.Logf("problem with number convertion: %s\n", err)
		return
	}
	if webappCommon.SliceContains(stidArr, "##dbg##") {
		h.isDebug = true
		if ctx != nil {
			ctx.SessionStorage().Set("dbg", true)
		}
	}
	if webappCommon.SliceContains(stidArr, "##dev##") {
		h.isDev = true
		if ctx != nil {
			ctx.SessionStorage().Set("dev", true)
		}
	}
	if webappCommon.SliceContains(stidArr, "##exp##") {
		h.isExperimental = true
		if ctx != nil {
			ctx.SessionStorage().Set("exp", true)
		}
	}

	// app.Logf("******************************* stid from reload: %+#v\n", stid)
	ua := app.Window().Get("navigator").Get("userAgent").String()

	token := shortener.GenerateTokenTweaked(ua+seed, tokenStartPos, tokenLen, 0)
	app.Logf("******************************* calculated token: %s\n", token)
	if token == "" {
		app.Logf("problem with token generation\n")
		return
	}
	h.sessionToken = token
	if ctx != nil {
		app.Logf("setting token on local storage\n")
		ctx.SessionStorage().Set("token", token)
	}
}

func (h *short) reload(ctx app.Context) {
	var err error
	lurl := app.Window().URL()
	lurl.Path = "/"
	lurl.RawQuery = "reload"
	// var redirectLoc *url.URL

	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	req, err := http.NewRequest(http.MethodGet, lurl.String(), nil)
	if err != nil {
		app.Logf("failed to create new request: %s\n", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set(webappCommon.FRequestTokenSeed, uuid.NewString()+"#*$$"+uuid.NewString())
	// req.Header.Set("Content-Type", "text/plain")
	x, err := httputil.DumpRequest(req, true)
	app.Logf("invoking request: %+#v, err: %v\n", string(x), err)
	resp, err := client.Do(req)
	if err != nil {
		app.Logf("failed to invoke request: %s\n", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		app.Logf("response not ok: %v\n", resp.StatusCode)
		return
	}
	// x, err = httputil.DumpResponse(resp, true)
	// app.Logf("getting response location: %#+v, err: %v\n", string(x), err)
	// loc, err := resp.Location()
	// app.Logf("response location: [%#+v], error [%v]\n", loc, err)

	// defer resp.Body.Close()
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	h.handleError("response reads", err)
	// 	return
	// }
	// app.Logf("response: %+#v\n", body)
	// app.Logf("redirect Location: %+#v\n", redirectLoc)
	x, err = httputil.DumpResponse(resp, true)
	app.Logf("getting response : %#+v, err: %v\n", string(x), err)
	app.Logf("header %s : %#+v\n", webappCommon.FSaltTokenID, resp.Header.Values(webappCommon.FSaltTokenID))
	stid := resp.Header.Get(webappCommon.FSaltTokenID)
	d, err := shortener.Base64SE.Decode(stid)
	app.Logf("stid: %v -> %v, error: %v", stid, string(d), err)
	// for i, e := range strings.Split(stid, ", ") {
	// 	// stid := resp.Header.Values(webappCommon.FSaltTokenID)
	// 	// for i, e := range stid {
	// 	d, err := shortener.Base64SE.Decode(e)
	// 	app.Logf("stid: %d) %v -> %v, error: %v", i, string(e), string(d), err)
	// }
	h.parseSTID(ctx, stid)
}

func (h *short) load2(ctx app.Context) {
	if h.isLoaded {
		return
	}
	defer func() { h.isLoaded = true }()
	h.logInit()
	lurl := app.Window().URL()
	app.Logf("url: %#+v\n", lurl)
	if strings.Contains(lurl.Path, webappCommon.PrivatePath) && lurl.Query().Has(webappCommon.FShortKey) {
		h.isPrivate = true
		if lurl.Query().Has(webappCommon.PasswordProtected) {
			h.privatePassSalt = lurl.Query().Get(webappCommon.PasswordProtected)
			h.isResultLocked = true
		}
	} else if strings.Contains(lurl.Path, webappCommon.RemovePath) && lurl.Query().Has(webappCommon.FShortKey) {
		h.isRemove = true
		if lurl.Query().Has(webappCommon.PasswordProtected) {
			h.removePassSalt = lurl.Query().Get(webappCommon.PasswordProtected)
			h.isResultLocked = true
		}
	} else if strings.Contains(lurl.Path, webappCommon.PublicPath) && lurl.Query().Has(webappCommon.FShortKey) {
		h.isPublic = true
		if lurl.Query().Has(webappCommon.PasswordProtected) {
			h.publicPassSalt = lurl.Query().Get(webappCommon.PasswordProtected)
			h.isResultLocked = true
		}
	} else {
		if lurl.Query().Has(webappCommon.FSaltTokenID) {
			stid, ok := lurl.Query()[webappCommon.FSaltTokenID]
			if !ok {
				app.Logf("problem with stid: %#v\n", stid)
				return
			}
			h.parseSTID(ctx, stid[0])
		} else {
			if ctx != nil && h.sessionToken == "" {
				err := ctx.SessionStorage().Get("token", &h.sessionToken)
				app.Logf("getting token from local storage: <%s>, err: %v\n", h.sessionToken, err)

			}
		}
	}
	app.Logf("load2....\n")
	kpair, _ := ecies.GenerateKey()
	app.Logf("ecies\nprv: %#+v\npub: %#+v\n", kpair, kpair.PublicKey)
}

func (h *short) logInit() {
	lurl := app.Window().URL()
	for k, v := range lurl.Query() {
		var decQuery []byte
		if len(v) > 0 {
			if len(v) > 1 {
				app.Logf("skipping non first values for key %s\n", k)
			}
			decQuery, _ = shortener.Base64SE.Decode(v[0])
		} else {
			decQuery, _ = shortener.Base64SE.Decode(k)
		}
		decQueryFields := strings.Split(string(decQuery), "$$")
		if webappCommon.SliceContains(decQueryFields, "##dev##") {
			h.isDev = true
		}
		if webappCommon.SliceContains(decQueryFields, "##dbg##") {
			h.isDebug = true
		}
	}
	app.Logf("lurl: <%s>, debug: %v, dev: %v\n", lurl, h.isDebug, h.isDev)

	h.isDebug = true
	if h.isDebug || h.isDev {
		orgLog := app.DefaultLogger
		app.DefaultLogger = func(format string, v ...any) {
			orgLog("[%s]: "+format,
				append([]any{time.Now().Format(time.RFC3339)}, v...)...)
		}
		app.Logf("webapp run with isDebug: %v and isDev: %v\n", h.isDebug, h.isDev)
	} else {
		// app.Logf("webapp run with isDebug: %v and isDev: %v\n", h.isDebug, h.isDev)
		app.Logf("disabling any log messages")
		app.DefaultLogger = func(format string, v ...any) {}
	}
}

func (h *short) OnInit() {
	h.load2(nil)
	app.Logf("******************************* init - build version :<%s>, time: <%s>\n", common.BuildVersion, common.BuildTime)
}
func (h *short) OnPreRender(ctx app.Context) {
	app.Logf("******************************* prerender")
}
func (h *short) OnDisMount() {
	app.Logf("******************************* dismount")
}
func (h *short) OnMount(ctx app.Context) {
	h.load2(ctx)
	app.Logf("******************************* mount")
}
func (h *short) OnNav(ctx app.Context) {
	h.load2(ctx)
	ctx.SessionStorage().Get("exp", &h.isExperimental)
	ctx.SessionStorage().Get("dev", &h.isDev)
	ctx.SessionStorage().Get("dbg", &h.isDebug)
	ctx.SessionStorage().Get("token", &h.sessionToken)
	ctx.SessionStorage().Clear()
	app.Logf("******************************* nav")
}
func (h *short) OnResize(ctx app.Context) {
	h.ResizeContent()
	app.Logf("******************************* resize")
}
func (h *short) OnUpdate(ctx app.Context) {
	app.Logf("******************************* update")
	app.Logf("******************************* update123")
}
func (h *short) OnAppUpdate(ctx app.Context) {
	h.updateAvailable = ctx.AppUpdateAvailable()
	app.Logf("11111111111111111111111111111111\n")
	app.Logf("******************************* app update: %v\n", h.updateAvailable)
	// ctx.Async(
	// 	func() {
	// 	app.Log("3")
	// 	<-time.After(time.Second)
	// 	app.Log("2")
	// 	<-time.After(time.Second)
	// 	app.Log("1")
	// 	<-time.After(time.Second)
	app.Log("!!! reloading ...")
	ctx.Reload() // TODO:  maybe do it async .. ?, maybe dont force update .. ?
	// })
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

func (h *short) handleError(msg string, err error) {
	errElem := app.Window().GetElementByID("movText")
	app.Logf("handle error: %s : %v\n", msg, err)
	if !errElem.IsNull() {
		errElem.Set("value", fmt.Sprintf("%s: error occurred: %v", msg, err))
	}
}

func (h *short) createShort(ctx app.Context) {
	var err error
	app.Logf("!!URL: %+#v\n", app.Window().URL())
	appUrl := app.Window().URL()
	dest := url.URL{
		Scheme: appUrl.Scheme,
		Host:   appUrl.Host,
	}
	elem := app.Window().GetElementByID("shortInputText")
	data := elem.Get("value").String()
	destCreate := dest.String()
	app.Logf("!!! new dest: %s\n", destCreate)
	payload := []byte(data)
	isEnc := false

	if url, ok := urlCheck(data); ok && !h.isShortAsData {

		destCreate += "/create-short-url"
		payload, err = json.Marshal(map[string]string{
			"url": url,
		})
		if err != nil {
			h.handleError("url problem", err)
			return
		}
	} else {
		destCreate += "/create-short-data"
		payload, isEnc = h.encryptPayload(payload)
	}

	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	// app.Logf("app %#v\n", app.)
	req, err := http.NewRequest(http.MethodPost, destCreate, bytes.NewBuffer(payload))
	if err != nil {
		h.handleError("new request", err)
		return
	}
	if isEnc {
		req.Header.Set(webappCommon.FDataEncrypted, uuid.NewString()) // the value doesnt really matter
	}
	if eDesc := app.Window().GetElementByID("shortDescription"); !eDesc.IsNull() {
		if desc := eDesc.Get("value").String(); desc != "" {
			req.Header.Set(webappCommon.FShortDesc, desc)
		}
	}
	if eNamed := app.Window().GetElementByID("shortNamedPublicShort"); !eNamed.IsNull() {
		if name := eNamed.Get("value").String(); name != "" {
			nameUnescaped, err := url.PathUnescape(name)
			if err == nil {
				name = nameUnescaped
			}
			app.Logf("namedpublic: final (%s), unescape(%s), origingal(%s) err (%v)\n",
				name, nameUnescaped, eNamed.Get("value").String(), err)
			req.Header.Set(webappCommon.FNamedPublic, shortener.Base64SE.Encode([]byte(name)))
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
		h.handleError("request invokes", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.handleError("response reads", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		h.handleError(fmt.Sprintf("response status: : %v", resp.StatusCode), nil)
		h.reload(ctx)
		return
	}

	// elem := app.Window().GetElementByID("in-out")
	err = json.Unmarshal(body, &h.resultMap)
	if err != nil {
		h.handleError("response read", err)
		return
	}

	r, err := json.MarshalIndent(h.resultMap, "", "\t")
	if err != nil {
		h.handleError("response read", err)
		return
	}
	h.result = string(r)
	h.resultReady = true

	// app.Logf("******************************* create short result: %s\n", string(body))
	// elem.Set("value", string(body))
	// app.Logf("******************************* create shoty: %#v\n", r)
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
		return host + from[which]
	}
	app.Logf("field <%s> not handled\n", which)
	h.handleError("create short link failed", fmt.Errorf("invalid field <%s> to create short link", which))
	return ""
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

func (h *short) getPublicShort(passToken string) error {

	var err error
	url := app.Window().URL()
	url.Path = "/" + url.Query().Get(webappCommon.FShortKey)
	url.RawQuery = ""

	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		app.Logf("failed to create new request: %s\n", err)
		return err
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
		return err
	}

	if resp.StatusCode != http.StatusOK {
		app.Logf("response not ok: %v\n", resp.StatusCode)
		return fmt.Errorf("status: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Logf("failed to read response body: %s\n", err)
		return err
	}
	encInd := resp.Header.Get(webappCommon.FDataEncrypted)
	h.isDataEncryptPassword = (len(encInd) == len(uuid.NewString()))
	app.Logf("resp body: <%s>\n", string(body))
	tup, err := store.NewTupleFromString(string(body))
	if err != nil {
		app.Logf("failed to parse body: %s\n", err)
		// return map[string]string{
		// 	store.FieldDATA: string(body),
		// }, []string{store.FieldDATA}, nil
		h.publicData = string(body)
	} else {
		h.publicUrl = tup.Get(store.FieldURL)
	}

	// return map[string]string{
	// 	store.FieldURL: tup.Get(store.FieldURL),
	// }, []string{store.FieldURL}, nil
	return nil
}

func (h *short) getRemoveShort(passToken string) (map[string]string, []string, error) {

	var err error
	url := app.Window().URL()
	url.Path = "/" + url.Query().Get(webappCommon.FShortKey)
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

	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	app.Logf("failed to read response body: %s\n", err)
	// 	return nil, nil, err
	// }
	// app.Logf("resp body: <%s>\n", string(body))
	// tup, err := store.NewTupleFromString(string(body))
	// if err != nil {
	// 	app.Logf("failed to parse body: %s\n", err)
	// 	return map[string]string{
	// 		store.FieldDATA: string(body),
	// 	}, []string{store.FieldDATA}, nil
	// }
	// return map[string]string{
	// 	store.FieldURL: tup.Get(store.FieldURL),
	// }, []string{store.FieldURL}, nil
	return nil, nil, nil
}

func (h *short) getPrivateInfo(passToken string) (map[string]string, []string, error) {

	var err error
	url := app.Window().URL()
	url.Path = "/" + url.Query().Get(webappCommon.FShortKey) + "/info"
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
	app.Logf("resp body: <%s>\n", string(body))

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
	knock1 := 0
	knock2 := 0
	if !h.isExperimental {
		knock1 = 3
		knock2 = 3
	}
	return app.Div().
		Class("row").
		Class("header").
		ID("logoTitle").
		Body(
			app.Div().
				// Class("col-md-4", "col-md-offset-2", "col-sm-4", "col-sm-offset-2", "col-xs-4", "col-xs-offset-3").
				Class("col-md-4", "col-md-offset-2", "col-sm-4", "col-sm-offset-2", "col-xs-4", "col-xs-offset-1").
				// Class("col-xs-4", "col-xs-offset-1").
				Class("logo").
				Body(
					app.Img().
						ID("logo").
						Class("logo-img").
						Class("img-responsive").
						Class("clickLinkAble").
						Src(ImgSource).
						Alt("Sh0r7 Logo").
						Width(200).
						OnClick(func(ctx app.Context, e app.Event) {
							app.Logf("knock 1 (%d), 2 (%d)\n", knock1, knock2)
							knock1 -= 1
							if (knock1 > 0 || knock2 > 0) || h.isExperimental {
								return
							}
							lurl := app.Window().URL()
							lurl.Path = webappCommon.ShortPath
							lurl.RawQuery = lurl.RawQuery + "&exp"
							app.Window().Get("location").Set("href", lurl.String())
							// ctx.NavigateTo(lurl)
						}),
				),
			app.Div().
				// Class("col-md-6", "col-md-offset-0", "col-sm-4", "col-sm-offset-0", "col-xs-6", "col-xs-offset-3").
				Class("col-md-6", "col-md-offset-0", "col-sm-4", "col-sm-offset-0", "col-xs-6", "col-xs-offset-1").
				// Class("col-xs-6", "col-xs-offset-1").
				Class("text").
				Body(
					app.H1().
						Body(
							app.Text("Sh0r7"),
						).
						OnMouseOver(func(ctx app.Context, e app.Event) {
							elem := app.Window().GetElementByID("titleH2")
							elem.Set("innerText", "the world!!")
						}).
						OnMouseOut(func(ctx app.Context, e app.Event) {
							elem := app.Window().GetElementByID("titleH2")
							elem.Set("innerText", "the earth562!")
						}).
						Class("clickLinkAble").
						OnClick(func(ctx app.Context, e app.Event) {
							app.Logf("knock 1 (%d), 2 (%d)\n", knock1, knock2)
							knock2 -= 1
							if (knock2 > 0 || knock1 > 0) || h.isExperimental {
								return
							}
							lurl := app.Window().URL()
							lurl.Path = webappCommon.ShortPath
							// url.RawQuery = ""
							app.Window().Get("location").Set("href", lurl.String())
							// ctx.NavigateTo(lurl)
						}),
					app.H2().ID("titleH2").
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

// app.Logf("entered password option for %s\n", which)
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

func (h *short) DebugWindow() app.UI {
	panerr := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	var removeMoveFunc func()
	mouseDownFunc := func(ctx app.Context, e app.Event) {
		moveable := app.Window().GetElementByID("movable")
		left := strings.TrimSuffix(app.Window().Call("getComputedStyle", moveable.JSValue()).Call("getPropertyValue", "left").String(), "px")
		top := strings.TrimSuffix(app.Window().Call("getComputedStyle", moveable.JSValue()).Call("getPropertyValue", "top").String(), "px")
		mX, mY := app.Window().CursorPosition()
		leftNum, err := strconv.Atoi(left)
		panerr(err)
		topNum, err := strconv.Atoi(top)
		panerr(err)
		mouseMoveFunc := func(ctx app.Context, e app.Event) {
			cX, cY := app.Window().CursorPosition()
			dx := mX - cX
			dy := mY - cY
			leftValue := leftNum - dx
			topValue := topNum - dy
			moveable.Get("style").Set("left", fmt.Sprintf("%dpx", leftValue))
			moveable.Get("style").Set("top", fmt.Sprintf("%dpx", topValue))
		}
		removeMoveFunc = app.Window().AddEventListener("mousemove", mouseMoveFunc)
	}
	mouseUpFunc := func(ctx app.Context, e app.Event) { removeMoveFunc() }
	r := app.Div().
		ID("movable").
		Body(
			app.Div().
				ID("grabHere").
				Body(app.Text("Debug Window")).
				OnMouseDown(mouseDownFunc).OnMouseUp(mouseUpFunc).
				OnDrag(func(ctx app.Context, e app.Event) { app.Logf("drag") }).
				OnDragEnter(func(ctx app.Context, e app.Event) { app.Logf("drag enter") }).
				OnDragLeave(func(ctx app.Context, e app.Event) { app.Logf("drag leave") }).
				OnDragStart(func(ctx app.Context, e app.Event) { app.Logf("drag start ") }).
				OnDragEnd(func(ctx app.Context, e app.Event) { app.Logf("drag end") }).
				OnDragOver(func(ctx app.Context, e app.Event) { app.Logf("drag over") }),
			app.Textarea().
				ID("movText").
				Body(app.Text("messages goes here...")),
		)
	return r
}

func preventEnter(ctx app.Context, e app.Event) {
	// key := e.Get("key").String()
	keyCode := e.Get("keyCode").Int()
	// app.Logf("event key down: [%s]\n", key)
	// app.Logf("event key down: [%d]\n", keyCode)
	if keyCode == 13 { // preventing enter
		// app.Logf("preventing enter...\n")
		e.PreventDefault()
	}
}

func (h *short) encryptPayload(data []byte) (ret []byte, isEnc bool) {
	ret = data
	if encryptPass := app.Window().GetElementByID("encryptPasswordText"); h.isDataEncryptPassword && !encryptPass.IsNull() {
		if encKey := encryptPass.Get("value").String(); encKey != "" {
			ret = shortener.EncryptData([]byte(data), encKey)
			isEnc = true
		}
	}
	return
}

func (h *short) decryptPayload(data []byte) (ret []byte, isDec bool) {
	ret = data
	if encryptPass := app.Window().GetElementByID("encryptPasswordText"); h.isDataEncryptPassword && !encryptPass.IsNull() {
		if encKey := encryptPass.Get("value").String(); encKey != "" {
			ret = shortener.DecryptData([]byte(data), encKey)
			if len(ret) > 0 {
				isDec = true
			}
		}
	}
	return
}

func (h *short) RenderPublicWithPassword() (ret app.UI) {
	if len(h.publicData) > 0 {
		h.isDataEncryptPassword = true
		ret = app.Div().
			Body(
				app.Pre().ID("publicData").
					ContentEditable(false).
					OnContextMenu(func(ctx app.Context, e app.Event) {
						app.Logf("context menu triggered\n")
					}).
					Body(
						app.Text(h.publicData),
					),
				app.Div().
					Class().
					Body(
						app.Text("data is encrypted"),
						app.Br(),
						h.passwordOption("encrypt").
							OnChange(func(ctx app.Context, e app.Event) {
								fmt.Printf("something changed\n")
								dataLoc := app.Window().GetElementByID("publicData")
								if h.isDataEncryptPassword {
									// data := dataLoc.Get("innerText").String()
									// if data != "error" {
									dataBuf, err := hex.DecodeString(h.publicData)
									if err == nil {
										dec, isDec := h.decryptPayload(dataBuf)
										if isDec {
											app.Logf("dec data: <%s>\n", string(dec))
											dataLoc.Set("innerText", string(dec))
										} else {
											app.Logf("failed to decrypt data")
											dataLoc.Set("innerText", "error")
										}
									} else {
										app.Logf("problem with decode string, %s", err)
										dataLoc.Set("innerText", "error")
									}
									// }
								} else {
									dataLoc.Set("innerText", h.publicData)
								}
							}),
					),
				// Body(
				// 	app.Text("data is encrypted"),
				// 	app.Br(),
				// 	h.passwordOption("encrypt").OnChange(func(ctx app.Context, e app.Event) {
				// 		fmt.Printf("something changed\n")
				// 		dataLoc := app.Window().GetElementByID("publicData")
				// 		data := dataLoc.Get("innerText").String()
				// 		dataBuf, err := hex.DecodeString(data)
				// 		if err != nil {
				// 			app.Logf("problem with decode string, %s", err)
				// 			return
				// 		}
				// 		// if encKey := encryptPass.Get("value").String(); encKey != "" {
				// 		if h.isDataEncryptPassword {
				// 			dec, isDec := h.decryptPayload(dataBuf)
				// 			if !isDec {
				// 				app.Logf("failed to decrypt data")
				// 				return
				// 			}
				// 			app.Logf("dec data: <%s>\n")
				// 			dataLoc.Set("innerText", string(dec))
				// 		}
				// 	}),
				// ),
			)
	}
	return ret
}

func (h *short) navBar3() app.UI {
	return app.Div().
		Class("row").
		Class("nav").
		Body(
			app.Div().
				Class("col-sm-8", "col-sm-offset-2").
				Class("col-xs-8", "col-xs-offset-4").
				Body(
					app.Ul().
						Class("nav nav-pills navbar-right _nav-justified").
						Body(
							// app.Li().
							// 	Aria("role", "presentation").
							// 	Class("active").
							// 	Body(
							// 		app.A().
							// 			Href("#").
							// 			Body(
							// 				app.Text("Home"),
							// 			),
							// 	),
							app.Li().
								Class().
								Role("presentation").
								Body(
									app.A().
										Href("#").
										Body(
											app.Text("Sign In"),
										).
										OnClick(func(ctx app.Context, e app.Event) {
											jui := app.Window().GetElementByID("id02")
											style := jui.Get("style")
											style.Set("display", "block")
											// jui2 := app.Window().Get("style")
											// jui2.Set("overflow", "hidden")
											// prevent main view interaction (scroll) when modal is open
											html := app.Window().Get("document").Get("children").Index(0)
											html.Get("style").Set("overflow", "hidden")
										}),
									h.renderSignIn2(),
									// h.renderSignIn2().(app.HTMLDiv).
									// 	OnFocus(func(ctx app.Context, e app.Event) {
									// 		app.Logf("modal is on - change html and body style...")
									// 	}).
									// 	OnBlur(func(ctx app.Context, e app.Event) {
									// 		app.Logf("modal is off - change html and body style...")
									// 	}),
								).
								OnMouseOut(func(ctx app.Context, e app.Event) {
									ctx.JSSrc().Get("attributes").Get("class").Set("value", "")
								}).
								OnMouseOver(func(ctx app.Context, e app.Event) {
									ctx.JSSrc().Get("attributes").Get("class").Set("value", "active")
								}),
							app.Li().
								Class().
								Role("presentation").
								Body(
									app.A().
										Href("#").
										Body(
											app.Text("Sign Up"),
										).
										OnClick(func(ctx app.Context, e app.Event) {
											jui := app.Window().GetElementByID("id01")
											style := jui.Get("style")
											style.Set("display", "block")
											// prevent main view interaction (scroll) when modal is open
											html := app.Window().Get("document").Get("children").Index(0)
											html.Get("style").Set("overflow", "hidden")
										}),
									h.renderSignUp2(),
								).
								OnMouseOut(func(ctx app.Context, e app.Event) {
									ctx.JSSrc().Get("attributes").Get("class").Set("value", "")
								}).
								OnMouseOver(func(ctx app.Context, e app.Event) {
									ctx.JSSrc().Get("attributes").Get("class").Set("value", "active")
								}),
						),
				))
}
func (h *short) navBar2() app.UI {
	return app.Div().
		Class("row").
		Class("nav").
		Body(
			app.Div().
				Class("col-sm-10", "col-sm-offset-1").
				Body(
					app.Nav().
						Class("navbar navbar-default").
						Body(
							app.Div().
								Class("container-fluid").
								Body(
									app.Div().
										Class("navbar-header").
										Body(
											app.A().
												Class("navbar-brand").
												Href("#").
												Body(
													app.Img().
														Alt("Brand").
														Src(ImgSource).
														Height(25).Width(25),
												),
										),
									// ---- start ---
									app.Div().
										Class("navbar-collapse collapse in").
										ID("bs-example-navbar-collapse-1").
										Aria("expanded", true).
										Style("", "").
										Body(
											app.Ul().
												Class("nav navbar-nav").
												Body(
													app.Li().
														Class("").
														Body(
															app.A().
																Href("#").
																Body(
																	app.Text("Link1"), app.Span().
																		Class("sr-only").
																		Body(
																			app.Text("(current)"),
																		),
																),
														).
														OnMouseOut(func(ctx app.Context, e app.Event) {
															ctx.JSSrc().Get("attributes").Get("class").Set("value", "")
														}).
														OnMouseOver(func(ctx app.Context, e app.Event) {
															ctx.JSSrc().Get("attributes").Get("class").Set("value", "active")
														}),
													app.Li().
														Class("").
														Body(
															app.A().
																Href("#").
																Body(
																	app.Text("Link2"),
																),
														).
														OnMouseOut(func(ctx app.Context, e app.Event) {
															ctx.JSSrc().Get("attributes").Get("class").Set("value", "")
														}).
														OnMouseOver(func(ctx app.Context, e app.Event) {
															ctx.JSSrc().Get("attributes").Get("class").Set("value", "active")
														}),
												),
											app.Ul().
												Class("nav navbar-nav navbar-right").
												Body(
													app.Li().
														Class("").
														Body(
															app.A().
																Href("#").
																Body(
																	app.Text("Sign In"),
																).
																OnClick(func(ctx app.Context, e app.Event) {
																	jui := app.Window().GetElementByID("id02")
																	style := jui.Get("style")
																	style.Set("display", "block")
																}),
															h.renderSignIn2(),
														).
														OnMouseOut(func(ctx app.Context, e app.Event) {
															ctx.JSSrc().Get("attributes").Get("class").Set("value", "")
														}).
														OnMouseOver(func(ctx app.Context, e app.Event) {
															ctx.JSSrc().Get("attributes").Get("class").Set("value", "active")
														}),
													app.Li().
														Class("").
														Body(
															app.A().
																Href("#").
																Body(
																	app.Text("Sign Up"),
																).
																OnClick(func(ctx app.Context, e app.Event) {
																	jui := app.Window().GetElementByID("id01")
																	style := jui.Get("style")
																	style.Set("display", "block")
																}),
															h.renderSignUp2(),
														).
														OnMouseOut(func(ctx app.Context, e app.Event) {
															ctx.JSSrc().Get("attributes").Get("class").Set("value", "")
														}).
														OnMouseOver(func(ctx app.Context, e app.Event) {
															ctx.JSSrc().Get("attributes").Get("class").Set("value", "active")
														}),
												),
										),
								// ---- end ----
								),
						),
				),
		)
}

func (h *short) navBar() app.UI {
	// return h.newNav()
	return app.Nav().
		Class("navbar", "navbar-default").
		Body(
			app.Div().
				Class("container-fluid").
				Body(
					// <!-- Brand and toggle get grouped for better mobile display -->
					app.Div().
						Class("navbar-header").
						Body(
							app.Button().
								Type("button").
								Class("navbar-toggle collapsed").
								DataSet("toggle", "collapse").
								DataSet("target", "#bs-example-navbar-collapse-1").
								Aria("expanded", "false").
								Body(
									app.Span().Class("sr-only").
										Body(app.Text("Toggle navigation")),
									app.Span().Class("icon-bar"),
									app.Span().Class("icon-bar"),
									app.Span().Class("icon-bar"),
								),
							app.A().
								Class("navbar-brand").
								Href("#").
								Body(
									app.Text("Brand"),
								),
						),
					// <!-- Collect the nav links, forms, and other content for toggling -->
					app.Div().
						Class("collapse navbar-collapse").
						ID("bs-example-navbar-collapse-1").
						Body(
							app.Ul().
								Class("nav navbar-nav").
								Body(
									app.Li().
										Class("active").
										Body(
											app.A().
												Href("#").
												Body(
													app.Text("Link"), app.Span().
														Class("sr-only").
														Body(app.Text("(current)")),
												),
										),
									app.Li().
										Body(app.A().Href("#").Body(app.Text("Link"))),
									app.Li().
										Class("dropdown").
										Body(
											app.A().
												Href("#").
												Class("dropdown-toggle").
												DataSet("toggle", "dropdown").
												Aria("role", "button").
												Aria("haspopup", true).
												Aria("expanded", "false").
												Body(
													app.Text("Dropdown"),
													app.Span().Class("caret"),
												),
											app.Ul().
												Class("dropdown-menu").
												Body(
													app.Li().
														Body(app.A().Href("#").Body(app.Text("Action"))),
													app.Li().
														Body(app.A().Href("#").Body(app.Text("Another action"))),
													app.Li().
														Body(app.A().Href("#").Body(app.Text("Something else here"))),
													app.Li().
														Aria("role", "separator").
														Class("divider"),
													app.Li().
														Body(app.A().Href("#").Body(app.Text("Separated link"))),
													app.Li().
														Aria("role", "separator").
														Class("divider"),
													app.Li().
														Body(app.A().Href("#").Body(app.Text("One more separated link"))),
												),
										),
								),
							app.Form().
								Class("navbar-form navbar-left").
								Body(
									app.Div().
										Class("form-group").
										Body(app.Input().Type("text").Class("form-control").Placeholder("Search")),
									app.Button().
										Type("submit").
										Class("btn btn-default").
										Body(app.Text("Submit")),
								),
							app.Ul().
								Class("nav navbar-nav navbar-right").
								Body(
									app.Li().Body(app.A().Href("#").Body(app.Text("Link"))),
									app.Li().
										Class("dropdown").
										Body(
											app.A().
												Href("#").
												Class("dropdown-toggle").
												DataSet("toggle", "dropdown").
												Aria("role", "button").
												Aria("haspopup", true).
												Aria("expanded", "false").
												Body(
													app.Text("Dropdown"), app.Span().
														Class("caret"),
												),
											app.Ul().
												Class("dropdown-menu").
												Body(
													app.Li().
														Body(app.A().Href("#").Body(app.Text("Action"))),
													app.Li().
														Body(app.A().Href("#").Body(app.Text("Another action"))),
													app.Li().
														Body(app.A().Href("#").Body(app.Text("Something else here"))),
													app.Li().
														Aria("role", "separator").
														Class("divider"),
													app.Li().Body(app.A().Href("#").Body(app.Text("Separated link"))),
												),
										),
								),
						),
				),
		)

	/*
			<div class="navbar">
		  <a href="#home">Home</a>
		  <a href="#news">News</a>
		  <div class="dropdown">
		    <button class="dropbtn">Dropdown
		      <i class="fa fa-caret-down"></i>
		    </button>
		    <div class="dropdown-content">
		      <a href="#">Link 1</a>
		      <a href="#">Link 2</a>
		      <a href="#">Link 3</a>
		    </div>
		  </div>
		</div>
	*/
	return app.Div().
		Class("navbar").
		Body(
			app.A().
				Href("#home").
				Body(
					app.Text("Home"),
				),
			app.A().
				Href("#news").
				Body(
					app.Text("News"),
				),
			app.Div().
				Class("dropdown").
				Body(
					app.Button().
						Class("dropbtn").
						Body(
							app.Text("Dropdown"), app.I().
								Class("fa fa-caret-down"),
						),
					app.Div().
						Class("dropdown-content").
						Body(
							app.A().
								Href("#").
								Body(
									app.Text("Link 1"),
								),
							app.A().
								Href("#").
								Body(
									app.Text("Link 2"),
								),
							app.A().
								Href("#").
								Body(
									app.Text("Link 3"),
								),
						),
				),
		)
	return app.Div().
		Class("row").
		Class("navbar").
		Body(
			app.A().Href("#home").Body(app.Text("Home")),
			app.A().Href("#news").Body(app.Text("News")),
			app.Div().Class("dropdown").Body(
				app.Button().Class("dropbtn").Body(
					app.Text("Dropdown"),
					app.I().Class("fa", "fa-caret-down"),
				),
				app.Div().Class("sdropdown-content").Body(
					app.A().Href("#").Body(app.Text("Link 1")),
					app.A().Href("#").Body(app.Text("Link 2")),
					app.A().Href("#").Body(app.Text("Link 3")),
				),
			),
		)
	// return app.If(h.isDev,
	// 	app.Div().
	// 		ID("signup").
	// 		Class("col-sm-1").
	// 		Body(
	// 			h.renderAccount(),
	// 		))
}

func (h *short) newNav() app.UI {
	return app.Div().
		ID("container").
		Body(
			app.Nav().
				Body(
					app.Div().
						ID("logo").
						Body(
							app.Text("Company"),
						),
					app.Ul().
						Body(
							app.Li().
								Body(
									app.A().
										Href("#").
										Body(
											app.Text("Home"),
										),
								),
							app.Li().
								Body(
									app.A().
										Href("#").
										Body(
											app.Text("About"),
										),
								),
							app.Li().
								Class("dropdown").
								OnMouseOver(func(ctx app.Context, e app.Event) {
									// app.Window().Call("hover", "this")
									// ctx.JSSrc()
								}).
								OnMouseOut(func(ctx app.Context, e app.Event) {
									// app.Window().Call("unhover", "this")
								}).
								// OnMouseOver("hover(this);").
								// OnMouseOut("out(this);").
								Body(
									app.A().
										ID("gallerid").
										Href("#").
										Body(
											app.Text("Gallery"), app.I().
												Class("fa fa-caret-down"),
										),
									app.Div().
										Class("dd").
										Body(
											app.Div().
												ID("up_arrow"),
											app.Ul().
												Body(
													app.Li().
														Body(
															app.A().
																Href("#").
																Body(
																	app.Text("2019"),
																),
														),
													app.Li().
														Body(
															app.A().
																Href("#").
																Body(
																	app.Text("2018"),
																),
														),
													app.Li().
														Body(
															app.A().
																Href("#").
																Body(
																	app.Text("2017"),
																),
														),
												),
										),
								),
							app.Li().
								Body(
									app.A().
										Href("#").
										Body(
											app.Text("Contact"),
										),
								),
							app.Li().
								Class("dropdown").
								Body(
									app.A().
										Href("#").
										Body(
											app.Text("Others"), app.I().
												Class("fa fa-caret-down"),
										),
									app.Div().
										Class("dd").
										Body(
											app.Div().
												ID("u_a_c").
												Body(
													app.Div().
														ID("up_arrow"),
												),
											app.Ul().
												Body(
													app.Li().
														Body(
															app.A().
																Href("#").
																Body(
																	app.Text("DOCS"),
																),
														),
													app.Li().
														Body(
															app.A().
																Href("#").
																Body(
																	app.Text("API"),
																),
														),
													app.Li().
														Body(
															app.A().
																Href("#").
																Body(
																	app.Text("PROJECTS"),
																),
														),
												),
										),
								),
						),
				),
		)
}
