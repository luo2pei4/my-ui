package main

import (
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type Page struct {
	remoteIpInput widget.Editor
	usernameInput widget.Editor
	passwordInput widget.Editor
	cmdInput      widget.Editor
	execButton    widget.Clickable
	theme         *material.Theme
}

func NewPage() *Page {
	page := &Page{
		theme: material.NewTheme(),
	}
	page.remoteIpInput.SingleLine = true
	page.usernameInput.SingleLine = true
	page.passwordInput.SingleLine = true
	return page
}

func (p *Page) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(
			func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Alignment: layout.Middle,
						Spacing:   layout.SpaceSides,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return widget.Border{
								Color: color.NRGBA{R: 204, G: 204, B: 204, A: 255},
								Width: unit.Dp(1),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{
									Top:    unit.Dp(5),
									Bottom: unit.Dp(5),
									Left:   unit.Dp(5),
								}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.X = gtx.Dp(280)
									gtx.Constraints.Max.X = gtx.Dp(280)
									return material.Editor(p.theme, &p.remoteIpInput, "remote ip address").Layout(gtx)
								})
							})
						}),
						layout.Rigid(layout.Spacer{Width: 10}.Layout),
						// 用户名输入框
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return widget.Border{
								Color: color.NRGBA{R: 204, G: 204, B: 204, A: 255},
								Width: unit.Dp(1),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{
									Top:    unit.Dp(5),
									Bottom: unit.Dp(5),
									Left:   unit.Dp(5),
								}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.X = gtx.Dp(200)
									gtx.Constraints.Max.X = gtx.Dp(200)
									return material.Editor(p.theme, &p.usernameInput, "user name").Layout(gtx)
								})
							})
						}),
						layout.Rigid(layout.Spacer{Width: 10}.Layout),
						// 密码输入框
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return widget.Border{
								Color: color.NRGBA{R: 204, G: 204, B: 204, A: 255},
								Width: unit.Dp(1),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{
									Top:    unit.Dp(5),
									Bottom: unit.Dp(5),
									Left:   unit.Dp(5),
								}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.X = gtx.Dp(200)
									gtx.Constraints.Max.X = gtx.Dp(200)
									return material.Editor(p.theme, &p.passwordInput, "password").Layout(gtx)
								})
							})
						}),
					)
				})
			},
		),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Left:   unit.Dp(10),
				Right:  unit.Dp(10),
				Bottom: unit.Dp(5),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal,
					Alignment: layout.Middle,
					Spacing:   layout.SpaceSides,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return widget.Border{
							Color: color.NRGBA{R: 204, G: 204, B: 204, A: 255},
							Width: unit.Dp(1),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top:    unit.Dp(5),
								Bottom: unit.Dp(5),
								Left:   unit.Dp(5),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = gtx.Dp(710)
								gtx.Constraints.Max.X = gtx.Dp(710)
								return material.Editor(p.theme, &p.cmdInput, "cmd").Layout(gtx)
							})
						})
					}),
				)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return Button(gtx, 80, p.theme, &p.execButton, "execute")
		}),
	)
}

func loop(win *app.Window) error {
	page := NewPage()
	var ops op.Ops
	for {
		switch e := win.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			page.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func main() {
	go func() {
		win := new(app.Window)
		win.Option(app.Title("remote-ssh"))
		if err := loop(win); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func Button(gtx layout.Context, width unit.Dp, th *material.Theme, wid *widget.Clickable, txt string) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Dp(width)
	gtx.Constraints.Max.X = gtx.Dp(width)
	return material.Button(th, wid, txt).Layout(gtx)
}
