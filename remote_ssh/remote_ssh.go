package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/crypto/ssh"
)

type Page struct {
	remoteIpInput widget.Editor
	usernameInput widget.Editor
	passwordInput widget.Editor
	cmdInput      widget.Editor
	execButton    widget.Clickable
	modalButton   widget.Clickable
	theme         *material.Theme
	showDialog    bool
	confirmMsg    string
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
	mainPage := layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Middle,
	}
	mainPage.Layout(gtx,
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
									p.passwordInput.Mask = '*'
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
			// 点击按钮逻辑
			if p.execButton.Clicked(gtx) {
				p.checkInput()
				if !p.showDialog {
					p.executeCmd()
				}
			}
			return Button(gtx, 80, p.theme, &p.execButton, "execute")
		}),
	)

	// 弹出对话框
	if p.showDialog {
		p.drawConfirmDialog(gtx)
	}

	return mainPage.Layout(gtx)
}

func (p *Page) checkInput() {
	itemName := ""
	switch {
	case p.remoteIpInput.Text() == "":
		itemName = "ip address"
	case p.usernameInput.Text() == "":
		itemName = "login user name"
	case p.passwordInput.Text() == "":
		itemName = "login user password"
	case p.cmdInput.Text() == "":
		itemName = "command"
	}
	if len(itemName) != 0 {
		p.confirmMsg = fmt.Sprintf("%s is required", itemName)
		p.showDialog = true
	}
}

func (p *Page) drawConfirmDialog(gtx layout.Context) {
	// 全屏
	full := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, color.NRGBA{A: 150})
	full.Pop()

	// 窗口大小和位置（居中）
	boxW := min(gtx.Constraints.Max.X-80, 420)
	boxH := 150
	cx := gtx.Constraints.Max.X / 2
	cy := gtx.Constraints.Max.Y / 2
	rect := image.Rect(cx-boxW/2, cy-boxH/2, cx+boxW/2, cy+boxH/2)

	// 窗口背景（白色矩形）
	box := clip.Rect{Min: rect.Min, Max: rect.Max}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	box.Pop()

	// 将坐标系偏移道对话框左上角，然后在内部做正常布局
	offset := op.Offset(image.Pt(rect.Min.X, rect.Min.Y)).Push(gtx.Ops)
	inner := gtx
	inner.Constraints.Min = image.Pt(boxW, 0)
	inner.Constraints.Max = image.Pt(boxW, boxH)

	layout.UniformInset(unit.Dp(16)).Layout(inner, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(material.Body1(p.theme, p.confirmMsg).Layout),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(p.theme, &p.modalButton, "confirm")
				// 注意：Clicked() 无参
				if p.modalButton.Clicked(gtx) {
					p.showDialog = false
				}
				return btn.Layout(gtx)
			}),
		)
	})
	offset.Pop()
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

func (p *Page) executeCmd() {

	config := &ssh.ClientConfig{
		User: p.usernameInput.Text(),
		Auth: []ssh.AuthMethod{
			ssh.Password(p.passwordInput.Text()),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	host := p.remoteIpInput.Text()
	if len(strings.Split(host, ":")) == 1 {
		host += ":22"
	}
	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		p.confirmMsg = fmt.Sprintf("dail %s failed, %v", host, err)
		p.showDialog = true
		return
	}
	defer conn.Close()
	
	session, err := conn.NewSession()
	if err != nil {
		p.confirmMsg = fmt.Sprintf("create session failed, %v", err)
		p.showDialog = true
		return
	}
	defer session.Close()

	output, err := session.CombinedOutput(p.cmdInput.Text())
	if err != nil {
		p.confirmMsg = fmt.Sprintf("execute command failed, %v", err)
		p.showDialog = true
		return
	}
	fmt.Println(string(output))
}
