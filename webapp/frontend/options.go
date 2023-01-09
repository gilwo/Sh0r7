package frontend

import "github.com/maxence-charriere/go-app/v9/pkg/app"

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

func (h *short) OptionPrivate() []app.UI {
	return []app.UI{
		app.Div().
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
								if h.isOption8 {
									return "has-success"
								}
								return "has-warning"
							}()).
							Title("Enable short private link").
							ID("option8ID").
							Body(
								app.Label().
									Class("input-group-addon").
									Body(
										app.Input().
											ID("checkboxOption8").
											Type("checkbox").
											Value("").
											OnClick(func(ctx app.Context, e app.Event) {
												h.isOption8 = ctx.JSSrc().Get("checked").Bool()
											}),
									),
								app.If(h.isOption8,
									app.Input().
										Class("form-control").
										Class("syncTextStyle").
										Class("onlyText").
										ID("option8True").
										ReadOnly(true).Value("Enable short private link").
										OnClick(func(ctx app.Context, e app.Event) {
											elem := app.Window().GetElementByID("checkboxOption8")
											elem.Set("checked", false)
											h.isOption8 = false
										}),
								).Else(
									app.Input().Class("form-control").ReadOnly(true).Value("No short private link").
										OnClick(func(ctx app.Context, e app.Event) {
											elem := app.Window().GetElementByID("checkboxOption8")
											elem.Set("checked", true)
											h.isOption8 = true
										}),
								),
							),
					),
			),
		app.If(h.isOption8,
			app.Div().
				ID("shortOption4Wrapper").
				Class("row").
				Body(
					h.passwordOption(&h.isPrivatePassword, &h.isPrivatePasswordShown, "private"),
				),
		),
	}
}

func (h *short) OptionRemove() []app.UI {
	return []app.UI{
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
												h.isOption5 = ctx.JSSrc().Get("checked").Bool()
											}),
									),
								app.If(h.isOption5,
									app.Input().
										Class("form-control").
										Class("syncTextStyle").
										Class("onlyText").
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
		app.If(h.isOption5,
			app.Div().
				ID("shortOption7Wrapper").
				Class("row").
				Body(
					h.passwordOption(&h.isRemovePassword, &h.isRemovePasswordShown, "remove"),
				),
		),
	}
}
