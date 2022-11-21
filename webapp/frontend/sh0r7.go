package frontend

import (
	"fmt"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type short struct {
	app.Compo
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

				app.Textarea().
					Class("form-control").
					Rows(5).
					Cols(20).
					Wrap("off").
					Placeholder("long url or data..."),
			),
			app.Div().
				Class("v2_23"),
			app.Span().
				Class("v2_24").
				Body(
					app.Button().
						Class("btn btn-primary").
						Body(
							app.Text("short it"),
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
