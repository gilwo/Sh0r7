package frontend

import (
	"fmt"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type short struct {
	app.Compo
}

func (h *short) Render() app.UI {
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
