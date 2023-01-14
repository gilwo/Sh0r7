package frontend

import (
	"time"

	"github.com/gilwo/Sh0r7/store"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (h *short) shortCreationOutput() app.UI {
	return app.Div().
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
			app.If(h.isOptionPrivate,
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
			),
			app.If(h.isOptionRemove,
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
		)
}

func (h *short) OptionsTitle() app.UI {
	return app.Div().
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
		)
}

func (h *short) OptionShortAsData() app.UI {
	return app.Div().
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
										Checked(false).
										OnClick(func(ctx app.Context, e app.Event) {
											h.isShortAsData = ctx.JSSrc().Get("checked").Bool()
										}),
								),
							app.If(h.isShortAsData,
								app.Input().
									Class("form-control").
									Class("syncTextStyle").
									Class("onlyText").
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
		)
}

func (h *short) OptionExpire() app.UI {
	return app.Div().
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
										Checked(false).
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
		)
}

func (h *short) OptionNamedPublicShort() app.UI {
	return app.Div().
		ID("shortOption9Wrapper").
		Class("row").
		Body(
			app.Div().
				Class("form-group").
				Class("col-md-offset-2", "col-md-6", "col-sm-offset-2", "col-sm-6", "col-xs-offset-1", "col-xs-10").
				Body(
					app.Div().
						Class("input-group").
						Class(func() string {
							if h.isNamedPublic {
								return "has-success"
							}
							return "has-warning"
						}()).
						Title("Use own name for this short (public only)").
						ID("shortNamedPublicShortWrapper").
						Body(
							app.Label().
								Class("input-group-addon").
								Body(
									app.Input().
										Type("checkbox").
										ID("checkboxNamedPublicShort").
										Value("").
										Checked(false).
										OnClick(func(ctx app.Context, e app.Event) {
											h.isNamedPublic = ctx.JSSrc().Get("checked").Bool()
										}),
								),
							app.If(h.isNamedPublic,
								app.Div().Class("input-group-addon").Body(
									app.Label().Body(
										app.Text("Named public"),
									).OnClick(func(ctx app.Context, e app.Event) {
										elem := app.Window().GetElementByID("checkboxNamedPublicShort")
										elem.Set("checked", false)
										h.isNamedPublic = false
									}),
								),
								app.Input().
									Class("form-control").
									Class("syncTextStyle").
									ID("shortNamedPublicShort").
									ReadOnly(false).Placeholder("my named short..."),
							).Else(
								app.Input().Class("form-control").ReadOnly(true).Value("Random public").OnClick(func(ctx app.Context, e app.Event) {
									elem := app.Window().GetElementByID("checkboxNamedPublicShort")
									elem.Set("checked", true)
									h.isNamedPublic = true
								}),
							),
						),
				),
		)
}

func (h *short) OptionDescription() app.UI {
	return app.Div().
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
										Checked(false).
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
		)
}

func (h *short) OptionPublic() app.UI {
	return app.Div().
		ID("shortOption6Wrapper").
		Class("row").
		Body(
			h.passwordOption(&h.isPublicPassword, &h.isPublicPasswordShown, "public"),
		)
}

func (h *short) OptionPrivate() app.UI {
	return app.Div().
		ID("shortOption8Wrapper").
		Class("row").
		Body(
			app.Div().
				Class("form-group").
				Class("col-md-offset-2", "col-md-6", "col-sm-offset-2", "col-sm-6", "col-xs-offset-1", "col-xs-10").
				Body(
					app.Div().
						Class("input-group").
						Class(func() string {
							if h.isOptionPrivate {
								return "has-success"
							}
							return "has-warning"
						}()).
						Title("Enable short private link").
						ID("optionPrivateID").
						Body(
							app.Label().
								Class("input-group-addon").
								Body(
									app.Input().
										ID("checkboxOptionPrivate").
										Type("checkbox").
										Value("").
										Checked(false).
										OnClick(func(ctx app.Context, e app.Event) {
											app.Logf("checkbox ID click: %s\n", ctx.JSSrc().Get("id").String())
											h.isOptionPrivate = ctx.JSSrc().Get("checked").Bool()
											if !h.isOptionPrivate {
												h.isPrivatePassword = false
												h.isPrivatePasswordShown = false
											}
										}),
								),
							app.If(h.isOptionPrivate,
								app.Input().
									Class("form-control").
									Class("syncTextStyle").
									Class("onlyText").
									ID("optionPrivateTrue").
									ReadOnly(true).Value("Enable short private link").
									OnClick(func(ctx app.Context, e app.Event) {
										elem := app.Window().GetElementByID("checkboxOptionPrivate")
										elem.Set("checked", false)
										h.isOptionPrivate = false
										h.isPrivatePassword = false
										h.isPrivatePasswordShown = false
									}),
							).Else(
								app.Input().Class("form-control").ReadOnly(true).Value("No short private link").
									OnClick(func(ctx app.Context, e app.Event) {
										elem := app.Window().GetElementByID("checkboxOptionPrivate")
										elem.Set("checked", true)
										h.isOptionPrivate = true
									}),
							),
						),
				),
			app.If(h.isOptionPrivate,
				h.passwordOption(&h.isPrivatePassword, &h.isPrivatePasswordShown, "private"),
			),
		)
}

func (h *short) OptionRemove() app.UI {
	return app.Div().
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
							if h.isOptionRemove {
								return "has-success"
							}
							return "has-warning"
						}()).
						Title("Enable short removal link").
						ID("optionRemoveID").
						Body(
							app.Label().
								Class("input-group-addon").
								Body(
									app.Input().
										ID("checkboxOptionRemove").
										Type("checkbox").
										Value("").
										Checked(false).
										OnClick(func(ctx app.Context, e app.Event) {
											h.isOptionRemove = ctx.JSSrc().Get("checked").Bool()
											if !h.isOptionRemove {
												h.isRemovePassword = false
												h.isRemovePasswordShown = false
											}
										}),
								),
							app.If(h.isOptionRemove,
								app.Input().
									Class("form-control").
									Class("syncTextStyle").
									Class("onlyText").
									ID("optionRemoveTrue").
									ReadOnly(true).Value("Enable short removal link").
									OnClick(func(ctx app.Context, e app.Event) {
										elem := app.Window().GetElementByID("checkboxOptionRemove")
										elem.Set("checked", false)
										h.isOptionRemove = false
										h.isRemovePassword = false
										h.isRemovePasswordShown = false
									}),
							).Else(
								app.Input().Class("form-control").ReadOnly(true).Value("No short removal link").
									OnClick(func(ctx app.Context, e app.Event) {
										elem := app.Window().GetElementByID("checkboxOptionRemove")
										elem.Set("checked", true)
										h.isOptionRemove = true
									}),
							),
						),
				),
			app.If(h.isOptionRemove,
				h.passwordOption(&h.isRemovePassword, &h.isRemovePasswordShown, "remove"),
			),
		)
}
