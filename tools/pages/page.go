package pages

import (
	"fmt"
	"time"
	"tools/icon"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

type Page interface {
	Actions() []component.AppBarAction
	Overflow() []component.OverflowAction
	Layout(gtx layout.Context, th *material.Theme) layout.Dimensions
	NavItem() component.NavItem
}

type Router struct {
	pages          map[any]Page
	current        any
	NavAnim        component.VisibilityAnimation
	NonModalDrawer bool
	*component.AppBar
	*component.ModalNavDrawer
}

func NewRouter() Router {
	modal := component.NewModal()

	nav := component.NewNav("Tools", "")
	modalNav := component.ModalNavFrom(&nav, modal)

	bar := component.NewAppBar(modal)
	bar.NavigationIcon = icon.MenuIcon

	na := component.VisibilityAnimation{
		State:    component.Invisible,
		Duration: time.Millisecond * 250,
	}
	return Router{
		pages:          make(map[any]Page),
		AppBar:         bar,
		NavAnim:        na,
		ModalNavDrawer: modalNav,
	}
}

func (r *Router) Register(tag any, p Page) {
	r.pages[tag] = p
	navItem := p.NavItem()
	navItem.Tag = tag
	if r.current == any(nil) {
		r.current = tag
		r.AppBar.Title = navItem.Name
		r.AppBar.SetActions(p.Actions(), p.Overflow())
	}
	r.ModalNavDrawer.AddNavItem(navItem)
}

func (r *Router) SwitchTo(tag any) {
	p, ok := r.pages[tag]
	if !ok {
		return
	}
	navItem := p.NavItem()
	r.current = tag
	r.AppBar.Title = navItem.Name
	r.AppBar.SetActions(p.Actions(), p.Overflow())
}

func (r *Router) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	for _, event := range r.AppBar.Events(gtx) {
		switch event.(type) {
		case component.AppBarNavigationClicked:
			if r.NonModalDrawer {
				r.NonModalDrawer = false
			} else {
				r.NonModalDrawer = true
			}
			if r.NonModalDrawer {
				r.NavAnim.ToggleVisibility(gtx.Now)
			} else {
				r.ModalNavDrawer.Appear(gtx.Now)
				r.NavAnim.Disappear(gtx.Now)
			}
		}
	}
	if r.ModalNavDrawer.NavDestinationChanged() {
		fmt.Printf("tag: %v\n", r.ModalNavDrawer.CurrentNavDestination())
		r.SwitchTo(r.ModalNavDrawer.CurrentNavDestination())
	}
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return r.AppBar.Layout(gtx, th, "test", "home")
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if r.NonModalDrawer {
						r.NavAnim.Disappear(gtx.Now)
					} else {
						r.NavAnim.Appear(gtx.Now)
					}
					gtx.Constraints.Max.X = 240 // 菜单宽度
					return r.NavDrawer.Layout(gtx, th, &r.NavAnim)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return r.pages[r.current].Layout(gtx, th)
					})
				}),
			)
		}),
	)
}
