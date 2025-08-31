package main

import (
	"left_menu/component/icon"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

type FramePage struct {
	appBar         *component.AppBar
	theme          *material.Theme
	NavAnim        component.VisibilityAnimation
	modalLayer     *component.ModalLayer
	NonModalDrawer bool
	*component.ModalNavDrawer
}

func newFramePage() *FramePage {

	modal := component.NewModal()
	nav := component.NewNav("Menu", "")

	p := &FramePage{
		appBar:     &component.AppBar{Title: "Left menu"},
		theme:      material.NewTheme(),
		modalLayer: modal,
		NavAnim: component.VisibilityAnimation{
			State:    component.Invisible,
			Duration: time.Millisecond * 250,
		},
	}
	p.ModalNavDrawer = component.ModalNavFrom(&nav, modal)
	p.appBar.NavigationIcon = icon.MenuIcon

	// 添加目录
	p.ModalNavDrawer.AddNavItem(component.NavItem{Name: "Home", Icon: icon.HomeIcon})
	p.ModalNavDrawer.AddNavItem(component.NavItem{Name: "About", Icon: icon.OtherIcon})
	p.ModalNavDrawer.AddNavItem(component.NavItem{Name: "About", Icon: icon.OtherIcon})
	p.ModalNavDrawer.AddNavItem(component.NavItem{Name: "About", Icon: icon.OtherIcon})
	p.ModalNavDrawer.AddNavItem(component.NavItem{Name: "About", Icon: icon.OtherIcon})

	return p
}

func (p *FramePage) Layout(gtx layout.Context) layout.Dimensions {
	for _, event := range p.appBar.Events(gtx) {
		switch event.(type) {
		case component.AppBarNavigationClicked:
			if p.NonModalDrawer {
				p.NonModalDrawer = false
			} else {
				p.NonModalDrawer = true
			}
			if p.NonModalDrawer {
				p.NavAnim.ToggleVisibility(gtx.Now)
			} else {
				p.ModalNavDrawer.Appear(gtx.Now)
				p.NavAnim.Disappear(gtx.Now)
			}
		}
	}
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.appBar.Layout(gtx, p.theme, "test", "home")
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if p.NonModalDrawer {
						p.NavAnim.Disappear(gtx.Now)
					} else {
						p.NavAnim.Appear(gtx.Now)
					}
					gtx.Constraints.Max.X = 240 // 菜单宽度
					return p.NavDrawer.Layout(gtx, p.theme, &p.NavAnim)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return material.Body1(p.theme, "this is context").Layout(gtx)
					})
				}),
			)
		}),
	)
}

func loop(win *app.Window) error {
	p := newFramePage()
	var ops op.Ops
	for {
		switch e := win.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			p.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func main() {
	go func() {
		win := new(app.Window)
		win.Option(app.Title("left-memu"), app.Maximized.Option())
		if err := loop(win); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
