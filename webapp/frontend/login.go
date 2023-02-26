package frontend

import "github.com/maxence-charriere/go-app/v9/pkg/app"

func (h *short) renderAccount() app.UI {
	return app.Div().Body(
		h.renderSignUp(),
	)
}

func (h *short) renderSignUp() app.UI {

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
						Name("email").Required(true),
					app.Label().For("psw").Body(app.B().Body(app.Text("Password"))),
					app.Input().Type("password").Placeholder("Enter Password").
						Name("psw").Required(true),
					app.Label().For("psw-repeat").Body(app.B().Body(app.Text("Repeat Password"))),
					app.Input().Type("password").Placeholder("Repeat Password").
						Name("psw-repeat").Required(true),

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
							app.Button().Type("submit").Class("signupbtn").Body(
								app.Text("Sign Up"),
							).OnClick(func(ctx app.Context, e app.Event) {
								app.Logf("sign up clicked with:")
								app.Logf("sign up clicked with:")
							}),
						),
				),
			),
		),
	)
}
