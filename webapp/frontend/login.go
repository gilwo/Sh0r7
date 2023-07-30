package frontend

import (
	"github.com/gilwo/Sh0r7/shortener"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (h *short) renderAccount() app.UI {
	return app.Div().Body(
		h.renderSignUp(),
	)
}

func (h *short) renderSignUp() app.UI {
	var inEmail, inPass, inRepPass string
	// https://www.w3schools.com/howto/tryit.asp?filename=tryhow_css_signup_form_modal
	return app.Div().ID("accountWrapperID").Class().Body(
		app.Button().
			ID("accountButtonID").
			Class().
			Body(
				app.Text("account"),
			).
			Style("width", "auto").
			OnClick(func(ctx app.Context, e app.Event) {
				jui := app.Window().GetElementByID("id01")
				style := jui.Get("style")
				style.Set("display", "block")
			}),
		app.Div().ID("id01").Class("modal").Body(
			app.Span().ID("").Class("close").Title("close modal").Body(
				app.Text("✖"),
			).OnClick(func(ctx app.Context, e app.Event) {
				jui := app.Window().GetElementByID("id01")
				style := jui.Get("style")
				style.Set("display", "none")
			}),
			app.Form().Class("modal-content").Body(
				app.Div().Class("container-fluid").Body(
					app.H1().Body(
						app.Text("Sign Up"),
					),
					app.P().Body(
						app.Text("Please fill in this form to create an account."),
					),
					app.Hr(),
					app.Label().For("email").Body(
						app.B().Body(app.Text("Email")),
					),
					app.Input().Type("text").Placeholder("Enter Email").
						Name("email").Required(true).
						OnBlur(func(ctx app.Context, e app.Event) {
							inEmail = ctx.JSSrc().Get("value").String()
						}),
					app.Label().For("psw").Body(app.B().Body(app.Text("Password"))),
					app.Input().Type("password").Placeholder("Enter Password").
						Name("psw").Required(true).
						OnBlur(func(ctx app.Context, e app.Event) {
							inPass = ctx.JSSrc().Get("value").String()
						}),
					app.Label().For("psw-repeat").Body(app.B().Body(app.Text("Repeat Password"))),
					app.Input().ID("signRepPassID").Type("password").Placeholder("Repeat Password").
						Name("psw-repeat").Required(true).
						OnBlur(func(ctx app.Context, e app.Event) {
							inRepPass = ctx.JSSrc().Get("value").String()
							if inPass != inRepPass {
								ctx.JSSrc().Set("type", "text")
								ctx.JSSrc().Set("value", "not the same")
							}
						}).OnFocus(func(ctx app.Context, e app.Event) {
						ctx.JSSrc().Set("type", "password")
					}),
					app.Label().Body(
						app.Input().Type("checkbox").Checked(true).
							Name("remember").
							Style("margin-bottom", "15px"),
						app.Text("Remember me"),
					),
					app.P().Body(
						app.Text("By creating an account you agree to our **Terms & Privacy**."),
						//   <p>By creating an account you agree to our <a href="#" style="color:dodgerblue">Terms & Privacy</a>.</p>
					),

					app.Div().Class("clearfix").
						Body(
							app.Button().Type("button").Class("cancelbtn").Body(
								app.Text("Cancel"),
							).OnClick(func(ctx app.Context, e app.Event) {
								jui := app.Window().GetElementByID("id01")
								style := jui.Get("style")
								style.Set("display", "none")
							}),
							app.Button().Type("button").Class("signupbtn").Body(
								app.Text("Sign Up"),
							).OnClick(func(ctx app.Context, e app.Event) {
								if len(inEmail+inPass) == 0 {
									return
								}
								app.Logf("sign up clicked with:")
								app.Logf("email <%s>, pass <%s>\n", inEmail, inPass)
							}),
						),
				),
			),
		),
	)
}

func (h *short) renderSignUp2() app.UI {
	var inEmail, inPass, inRepPass string
	// https://www.w3schools.com/howto/tryit.asp?filename=tryhow_css_signup_form_modal
	return app.Div().ID("accountWrapperID").Class().Body(
		// app.Button().
		// 	ID("accountButtonID").
		// 	Class().
		// 	Body(
		// 		app.Text("account"),
		// 	).
		// 	Style("width", "auto").
		app.Div().ID("id01").Class("modal").Body(
			app.Span().ID("").Class("closeModal").Title("close modal").Body(
				app.Text("✖"),
			).OnClick(func(ctx app.Context, e app.Event) {
				jui := app.Window().GetElementByID("id01")
				style := jui.Get("style")
				style.Set("display", "none")
				// set back scroll on main view when the modal is closed
				html := app.Window().Get("document").Get("children").Index(0)
				html.Get("style").Set("overflow", "visible")
			}),
			app.Form().Class("modal-content").Body(
				app.Div().Class("container-fluid").Body(
					app.H1().Body(
						app.Text("Sign Up"),
					),
					app.P().Body(
						app.Text("Please fill in this form to create an account."),
					),
					app.Hr(),
					app.Label().For("email").Body(
						app.B().Body(app.Text("Email")),
					),
					app.Input().Type("text").Placeholder("Enter Email").
						Name("email").Required(true).
						OnBlur(func(ctx app.Context, e app.Event) {
							inEmail = ctx.JSSrc().Get("value").String()
						}),
					app.Label().For("psw").Body(app.B().Body(app.Text("Password"))),
					app.Input().Type("password").Placeholder("Enter Password").
						Name("psw").Required(true).
						OnBlur(func(ctx app.Context, e app.Event) {
							inPass = ctx.JSSrc().Get("value").String()
						}),
					app.Label().For("psw-repeat").Body(app.B().Body(app.Text("Repeat Password"))),
					app.Input().ID("signRepPassID").Type("password").Placeholder("Repeat Password").
						Name("psw-repeat").Required(true).
						OnBlur(func(ctx app.Context, e app.Event) {
							inRepPass = ctx.JSSrc().Get("value").String()
							if inPass != inRepPass {
								ctx.JSSrc().Set("type", "text")
								ctx.JSSrc().Set("value", "not the same")
							}
						}).OnFocus(func(ctx app.Context, e app.Event) {
						ctx.JSSrc().Set("type", "password")
					}),
					app.Label().Body(
						app.Input().Type("checkbox").Checked(true).
							Name("store").
							Style("margin-bottom", "15px"),
						app.Text("Store me (able to restore account)"),
					).Title("Store email in server, restore account will be available"),
					app.P().Body(
						app.Text("By creating an account you agree to our **Terms & Privacy**."),
						//   <p>By creating an account you agree to our <a href="#" style="color:dodgerblue">Terms & Privacy</a>.</p>
					),

					app.Div().Class("clearfix").
						Body(
							app.Button().Type("button").Class("cancelbtn").Body(
								app.Text("Cancel"),
							).OnClick(func(ctx app.Context, e app.Event) {
								jui := app.Window().GetElementByID("id01")
								style := jui.Get("style")
								style.Set("display", "none")
							}),
							app.Button().Type("button").Class("signupbtn").Body(
								app.Text("Sign Up"),
							).OnClick(func(ctx app.Context, e app.Event) {
								if len(inEmail+inPass) == 0 {
									return
								}
								app.Logf("sign up clicked with:")
								app.Logf("email <%s>, pass <%s>\n", inEmail, inPass)
							}),
						),
				),
			),
		),
	)
}

func (h *short) renderSignIn2() app.UI {
	var inEmail, inPass string
	// https://www.w3schools.com/howto/tryit.asp?filename=tryhow_css_signup_form_modal
	return app.Div().ID("accountWrapperID").Class().Body(
		// app.Button().
		// 	ID("accountButtonID").
		// 	Class().
		// 	Body(
		// 		app.Text("account"),
		// 	).
		// 	Style("width", "auto").
		app.Div().ID("id02").Class("modal").Body(
			app.Span().ID("").Class("closeModal").Title("close modal").Body(
				app.Text("✖"),
			).OnClick(func(ctx app.Context, e app.Event) {
				jui := app.Window().GetElementByID("id02")
				style := jui.Get("style")
				style.Set("display", "none")
				// set back scroll on main view when the modal is closed
				html := app.Window().Get("document").Get("children").Index(0)
				html.Get("style").Set("overflow", "visible")
			}),
			app.Form().Class("modal-content").Body(
				app.Div().Class("container-fluid").Body(
					app.H1().Body(
						app.Text("Sign In"),
					),
					app.P().Body(
						app.Text("Please fill in to log in to your account."),
					),
					app.Hr(),
					app.Label().For("email").Body(
						app.B().Body(app.Text("Email")),
					),
					app.Input().Type("text").Placeholder("Enter Email").
						Name("email").Required(true).
						OnBlur(func(ctx app.Context, e app.Event) {
							inEmail = ctx.JSSrc().Get("value").String()
						}),
					app.Label().For("psw").Body(app.B().Body(app.Text("Password"))),
					app.Input().Type("password").Placeholder("Enter Password").
						Name("psw").Required(true).
						OnBlur(func(ctx app.Context, e app.Event) {
							inPass = ctx.JSSrc().Get("value").String()
						}),
					app.Label().Body(
						app.Input().Type("checkbox").Checked(true).
							Name("remember").
							Style("margin-bottom", "15px"),
						app.Text("Remember me"),
					),
					app.P().Body(
						app.Text("some login notification message."),
					),

					app.Div().Class("clearfix").
						Body(
							app.Button().Type("button").Class("cancelbtn").Body(
								app.Text("Cancel"),
							).OnClick(func(ctx app.Context, e app.Event) {
								jui := app.Window().GetElementByID("id02")
								style := jui.Get("style")
								style.Set("display", "none")
							}),
							app.Button().Type("button").Class("signinbtn").Body(
								app.Text("Sign In"),
							).OnClick(func(ctx app.Context, e app.Event) {
								if len(inEmail+inPass) == 0 {
									return
								}
								app.Logf("sign in clicked with:")
								app.Logf("email <%s>, pass <%s>\n", inEmail, inPass)
							}),
						),
				),
			),
		),
	)
}

func (h *short) signupLogic(contact, pass string, store bool) {
	shortener.EncryptData([]byte(contact), pass)
}
