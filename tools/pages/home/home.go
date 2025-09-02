package home

import (
	"tools/icon"

	page "tools/pages"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

// Page holds the state for a page demonstrating the features of
// the AppBar component.
type Page struct {
	*page.Router
}

// New constructs a Page with the provided router.
func New(router *page.Router) *Page {
	return &Page{
		Router: router,
	}
}

var _ page.Page = &Page{}

func (p *Page) Actions() []component.AppBarAction {
	return []component.AppBarAction{}
}

func (p *Page) Overflow() []component.OverflowAction {
	return []component.OverflowAction{}
}

func (p *Page) NavItem() component.NavItem {
	return component.NavItem{
		Name: "Home",
		Icon: icon.HomeIcon,
	}
}

func (p *Page) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Body1(th, "Welcom!").Layout(gtx)
		}),
	)
}
