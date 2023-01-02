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
	result              string
	resultMap           map[string]string
	resultReady         bool
	token               string
	expireValue         string
	debug               bool
	isPrivate           bool
	isShortAsData       bool   // indicate whether the input should be treated as data and not auto identify - option 1
	isExpireChecked     bool   // indicate whether the expiration feature is used - option 2
	isDescription       bool   // indicate whether the description feature is used when creating short - option 3
	isPrivatePassword   bool   // indicate whether the private password feature is used when creating short - option 4
	isPasswordNotHidden bool   // indicate whether the password is shown or hidden - sub option for option 4
	isOption5           bool   // indicate whether the short remove link feature is used when creating short - option 5
	isResultLocked      bool   // indicate that the result info is locked
	privatePassSalt     string // salt like used to create the passwork token
	passToken           string // the password token used to lock and unlock the short private
}

const (
	NOTEMESSAGE = "Sh0r7 service is still in alpha!"
)

var (
	ImgSource = "/web/logo.jpg"
	// imgSource: "logoL.png",
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
															Value(h.shortLink(store.FieldPublic, h.resultMap)),
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
																		app.Logf("current value: %v\n", elem.Get("body"))
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
															Value(h.shortLink(store.FieldPrivate, h.resultMap)),

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
																	app.Logf("current value: %v\n", elem.Get("body"))
																	elem.Set("textContent", "Copied")
																	ctx.After(400*time.Millisecond, func(ctx app.Context) {
																		elem.Set("textContent", "Copy")
																	})
																}),
															),
													),
											),
										app.If(h.isOption5,
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
																Value(h.shortLink(store.FieldRemove, h.resultMap)),
															app.Span().
																Class("input-group-btn").
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
																		app.Logf("current value: %v\n", elem.Get("body"))
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
												h.isPrivatePassword = false
												h.isPasswordNotHidden = false
												h.isOption5 = false
												app.Window().GetElementByID("checkboxShortAsData").Set("checked", false)
												app.Window().GetElementByID("checkboxExpire").Set("checked", false)
												app.Window().GetElementByID("checkboxDescription").Set("checked", false)
												app.Window().GetElementByID("checkboxPrivatePassword").Set("checked", false)
												app.Window().GetElementByID("checkboxOption5").Set("checked", false)
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
										Class("col-md-offset-2", "col-md-6", "col-sm-offset-2", "col-sm-6", "col-xs-offset-1", "col-xs-10").
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
													app.Label().
														Class("input-group-addon").
														Body(
															app.Input().
																ID("checkboxShortAsData").
																Type("checkbox").
																Value("").
																OnClick(func(ctx app.Context, e app.Event) {
																	h.isShortAsData = ctx.JSSrc().Get("checked").Bool()
																}),
														),
													app.If(h.isShortAsData,
														app.Input().
															Class("form-control").
															Class("syncTextStyle").
															ID("treatData").
															ReadOnly(true).Value("Input treated as data").
															OnClick(func(ctx app.Context, e app.Event) {
																elem := app.Window().GetElementByID("checkboxShortAsData")
																elem.Set("checked", false)
																h.isShortAsData = false
															}),
													).Else(
														app.Input().Class("form-control").ReadOnly(true).Value("Automatic treat input as data or Url").
															OnClick(func(ctx app.Context, e app.Event) {
																elem := app.Window().GetElementByID("checkboxShortAsData")
																elem.Set("checked", true)
																h.isShortAsData = true
															}),
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
										Class("col-md-offset-2", "col-md-6", "col-sm-offset-2", "col-sm-6", "col-xs-offset-1", "col-xs-10").
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
													app.Label().
														Class("input-group-addon").
														Body(
															app.Input().
																Type("checkbox").
																ID("checkboxExpire").
																Value("").
																OnClick(func(ctx app.Context, e app.Event) {
																	h.isExpireChecked = ctx.JSSrc().Get("checked").Bool()
																}),
														),
													app.If(h.isExpireChecked,
														app.Div().Class("input-group-addon").Body(
															app.Label().Body(
																app.Text("Expiration"),
															).OnClick(func(ctx app.Context, e app.Event) {
																elem := app.Window().GetElementByID("checkboxExpire")
																elem.Set("checked", false)
																h.isExpireChecked = false
															}),
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
															).OnChange(func(ctx app.Context, e app.Event) {
																h.expireValue = ctx.JSSrc().Get("value").String()
																app.Logf("select change value: %v\n", h.expireValue)
															}),
														),
														app.Div().Class("input-group-addon"),
													).Else(
														app.Input().Class("form-control").ReadOnly(true).Value("Default expiration (12 hours)").
															OnClick(func(ctx app.Context, e app.Event) {
																elem := app.Window().GetElementByID("checkboxExpire")
																elem.Set("checked", true)
																h.isExpireChecked = true
															}),
													),
												),
										),
								),
							app.Div().
								ID("shortOption3Wrapper").
								Class("row").
								Body(
									app.Div().
										Class("form-group").
										Class("col-md-offset-2", "col-md-6", "col-sm-offset-2", "col-sm-6", "col-xs-offset-1", "col-xs-10").
										Body(
											app.Div().
												Class("input-group").
												Class(func() string {
													if h.isDescription {
														return "has-success"
													}
													return "has-warning"
												}()).
												Title("short description for this short").
												ID("shortDescriptionWrapper").
												Body(
													app.Label().
														Class("input-group-addon").
														Body(
															app.Input().
																Type("checkbox").
																ID("checkboxDescription").
																Value("").
																OnClick(func(ctx app.Context, e app.Event) {
																	h.isDescription = ctx.JSSrc().Get("checked").Bool()
																}),
														),
													app.If(h.isDescription,
														app.Div().Class("input-group-addon").Body(
															app.Label().Body(
																app.Text("Description"),
															).OnClick(func(ctx app.Context, e app.Event) {
																elem := app.Window().GetElementByID("checkboxDescription")
																elem.Set("checked", false)
																h.isDescription = false
															}),
														),
														app.Input().
															Class("form-control").
															Class("syncTextStyle").
															ID("shortDescription").
															ReadOnly(false).Placeholder("my short description..."),
													).Else(
														app.Input().Class("form-control").ReadOnly(true).Value("No description").OnClick(func(ctx app.Context, e app.Event) {
															elem := app.Window().GetElementByID("checkboxDescription")
															elem.Set("checked", true)
															h.isDescription = true
														}),
													),
												),
										),
								),
							app.Div().
								ID("shortOption4Wrapper").
								Class("row").
								Body(
									app.Div().
										Class("form-group").
										Class("col-md-offset-2", "col-md-6", "col-sm-offset-2", "col-sm-6", "col-xs-offset-1", "col-xs-10").
										Body(
											app.Div().
												Class("input-group").
												Class(func() string {
													if h.isPrivatePassword {
														return "has-success"
													}
													return "has-warning"
												}()).
												Title("limit access to private link with password").
												ID("privateAccessPassword").
												Body(
													app.Label().
														Class("input-group-addon").
														Body(
															app.Input().
																Type("checkbox").
																ID("checkboxPrivatePassword").
																Value("").
																OnClick(func(ctx app.Context, e app.Event) {
																	h.isPrivatePassword = ctx.JSSrc().Get("checked").Bool()
																}),
														),
													app.If(h.isPrivatePassword,
														app.Div().Class("input-group-addon").Body(
															app.Label().Body(
																app.Text("Private password"),
															).OnClick(func(ctx app.Context, e app.Event) {
																elem := app.Window().GetElementByID("checkboxPrivatePassword")
																elem.Set("checked", false)
																h.isPrivatePassword = false
															}),
														),
														app.Input().
															Class("form-control").
															Class("syncTextStyle").
															ID("privatePasswordText").
															Value("").
															ReadOnly(false).Type("password"),
														func() app.UI {
															classIcon := "glyphicon glyphicon-eye-close"
															return app.Label().Class("input-group-addon").
																Body(
																	app.Span().
																		ID("passwordReveal").
																		Class(classIcon),
																).
																OnClick(func(ctx app.Context, e app.Event) {
																	h.isPasswordNotHidden = !h.isPasswordNotHidden
																	inputType := "password"
																	if h.isPasswordNotHidden {
																		inputType = "text"
																		classIcon = "glyphicon glyphicon-eye-open"
																	}
																	el := app.Window().GetElementByID("privatePasswordText")
																	el.Set("type", inputType)
																	jui := app.Window().GetElementByID("passwordReveal")
																	attrs := jui.Get("attributes")
																	attrs.Set("class", classIcon)
																})
														}(),
													).Else(
														app.Input().Class("form-control").ReadOnly(true).Value("No password on private link").
															OnClick(func(ctx app.Context, e app.Event) {
																elem := app.Window().GetElementByID("checkboxPrivatePassword")
																elem.Set("checked", true)
																h.isPrivatePassword = true
															}),
													),
												),
										),
								),
							app.Div().
								ID("shortOption5Wrapper").
								Class("row").
								Body(
									app.Div().
										Class("form-group").
										Class("col-md-offset-2", "col-md-6", "col-sm-offset-2", "col-sm-6", "col-xs-offset-1", "col-xs-10").
										Body(
											app.Div().
												Class("input-group").
												Class(func() string {
													if h.isOption5 {
														return "has-success"
													}
													return "has-warning"
												}()).
												Title("Enable short removal link").
												ID("option5ID").
												Body(
													app.Label().
														Class("input-group-addon").
														Body(
															app.Input().
																ID("checkboxOption5").
																Type("checkbox").
																Value("").
																OnClick(func(ctx app.Context, e app.Event) {
																	h.isShortAsData = ctx.JSSrc().Get("checked").Bool()
																}),
														),
													app.If(h.isOption5,
														app.Input().
															Class("form-control").
															Class("syncTextStyle").
															ID("option5True").
															ReadOnly(true).Value("Enable short removal link").
															OnClick(func(ctx app.Context, e app.Event) {
																elem := app.Window().GetElementByID("checkboxOption5")
																elem.Set("checked", false)
																h.isOption5 = false
															}),
													).Else(
														app.Input().Class("form-control").ReadOnly(true).Value("No short removal link").
															OnClick(func(ctx app.Context, e app.Event) {
																elem := app.Window().GetElementByID("checkboxOption5")
																elem.Set("checked", true)
																h.isOption5 = true
															}),
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
	lurl := app.Window().URL()
	app.Logf("url: %#+v\n", lurl)
	if strings.Contains(lurl.Path, webappCommon.PrivatePath) && lurl.Query().Has("key") {
		h.isPrivate = true
		if lurl.Query().Has(webappCommon.PasswordProtected) {
			h.privatePassSalt = lurl.Query().Get(webappCommon.PasswordProtected)
			h.isResultLocked = true
		}
	} else {
		h.getStID()
	}
	app.Logf("******************************* init")
}
func (h *short) OnPreRender() {
	app.Logf("******************************* prerender")
}
func (h *short) OnDisMount() {
	app.Logf("******************************* dismount")
}
func (h *short) OnMount() {
	app.Logf("******************************* mount")
}
func (h *short) OnNav() {
	app.Logf("******************************* nav")
}
func (h *short) OnUpdate() {
	app.Logf("******************************* update")
}
func (h *short) OnAppUpdate() {
	app.Logf("******************************* app update")
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
			req.Header.Set(webappCommon.FPrvPassToken, shortener.GenerateTokenTweaked(prvPass+h.token, 0, 30, 10))
		}
	}
	if h.expireValue != "" {
		req.Header.Set(webappCommon.FExpiration, h.expireValue)
	}
	if !h.isOption5 {
		req.Header.Set(webappCommon.FRemove, "false")
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set(webappCommon.FTokenID, h.token)
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
