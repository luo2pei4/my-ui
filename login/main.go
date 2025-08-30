package main

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type LoginPage struct {
	UsernameEditor widget.Editor
	PasswordEditor widget.Editor
	LoginBtn       widget.Clickable
	Username       string
	Password       string
	msg            string
	firstFrame     bool
}

func newLoginPage() *LoginPage {
	loginPage := &LoginPage{
		UsernameEditor: widget.Editor{},
		PasswordEditor: widget.Editor{},
		LoginBtn:       widget.Clickable{},
		firstFrame:     true,
	}
	loginPage.UsernameEditor.SingleLine = true
	loginPage.PasswordEditor.SingleLine = true
	return loginPage
}

func checkInput(username, password string) string {
	if len(username) == 0 {
		return "user name is required"
	}
	if len(password) == 0 {
		return "password is required"
	}
	return ""
}

func (lp *LoginPage) Layout(gtx layout.Context, win *app.Window, th *material.Theme) layout.Dimensions {
	// 初次加载的时候将光标设置到用户名输入框中
	if lp.firstFrame {
		gtx.Execute(key.FocusCmd{Tag: &lp.UsernameEditor})
		lp.firstFrame = false
	}

	// 当前页面的全局键盘事件监听
	for {
		event, ok := gtx.Event(
			key.Filter{Name: key.NameEnter},
			key.Filter{Name: key.NameUpArrow},
			key.Filter{Name: key.NameDownArrow},
		)
		if !ok {
			break
		}
		switch evn := event.(type) {
		case key.Event:
			switch evn.Name {
			case key.NameEnter:
				if evn.State == key.Release {
					fmt.Println("press key 'enter'")
				}
			case key.NameUpArrow:
				if evn.State == key.Release {
					fmt.Println("press key 'Up arrow'")
				}
			case key.NameDownArrow:
				if evn.State == key.Release {
					fmt.Println("press key 'Down arrow'")
				}
			}
		}
	}

	// 定义组件&布局
	return layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Middle,
	}.Layout(gtx,
		// 异常信息提示
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(30),
				Bottom: unit.Dp(5),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.H6(th, lp.msg)
				lbl.Alignment = text.Middle
				return lbl.Layout(gtx)
			})
		}),
		// 用户名输入框
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if len([]rune(lp.UsernameEditor.Text())) > 20 {
				runes := []rune(lp.UsernameEditor.Text())
				lp.UsernameEditor.SetText(string(runes[:20]))
			}
			lp.Username = lp.UsernameEditor.Text()
			return layout.Inset{
				Top:    unit.Dp(30),
				Bottom: unit.Dp(5),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return widget.Border{
					Color: color.NRGBA{R: 204, G: 204, B: 204, A: 255},
					Width: unit.Dp(1),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top:    unit.Dp(5),
						Bottom: unit.Dp(5),
						Left:   unit.Dp(5),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = gtx.Dp(300)
						gtx.Constraints.Max.X = gtx.Dp(300)
						return material.Editor(th, &lp.UsernameEditor, "user name").Layout(gtx)
					})
				})
			})
		}),
		// 密码输入框
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if len([]rune(lp.PasswordEditor.Text())) > 20 {
				runes := []rune(lp.PasswordEditor.Text())
				lp.PasswordEditor.SetText(string(runes[:20]))
			}
			lp.Password = lp.PasswordEditor.Text()
			lp.PasswordEditor.Mask = '*'
			return layout.Inset{
				Bottom: unit.Dp(30),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return widget.Border{
					Color: color.NRGBA{R: 204, G: 204, B: 204, A: 255},
					Width: unit.Dp(1),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top:    unit.Dp(5),
						Bottom: unit.Dp(5),
						Left:   unit.Dp(5),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = gtx.Dp(300)
						gtx.Constraints.Max.X = gtx.Dp(300)
						return material.Editor(th, &lp.PasswordEditor, "password").Layout(gtx)
					})
				})
			})
		}),
		// Login按钮
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if lp.LoginBtn.Clicked(gtx) {
				lp.msg = checkInput(lp.Username, lp.Password)
			}
			gtx.Constraints.Min.X = gtx.Dp(300)
			gtx.Constraints.Max.X = gtx.Dp(300)
			return material.Button(th, &lp.LoginBtn, "Login").Layout(gtx)
		}),
	)
}

func main() {
	go func() {
		win := new(app.Window)
		win.Option(app.Title("tools"))
		if err := loop(win); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(win *app.Window) error {
	logingPage := newLoginPage()
	th := material.NewTheme()
	var ops op.Ops
	for {
		switch e := win.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			logingPage.Layout(gtx, win, th)
			e.Frame(gtx.Ops)
		}
	}
}
