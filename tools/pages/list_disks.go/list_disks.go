package listdisks

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
	"tools/icon"
	page "tools/pages"
	"tools/utils/command"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"golang.org/x/crypto/ssh"
)

type Page struct {
	remoteIpInput widget.Editor
	usernameInput widget.Editor
	passwordInput widget.Editor
	execButton    widget.Clickable
	modalButton   widget.Clickable
	showDialog    bool
	confirmMsg    string
	resultEditor  widget.Editor
	devices       []command.BlockDevice
	*page.Router
}

func New(router *page.Router) *Page {
	page := &Page{
		Router: router,
	}
	page.remoteIpInput.SingleLine = true
	page.usernameInput.SingleLine = true
	page.passwordInput.SingleLine = true
	page.resultEditor.ReadOnly = true
	page.resultEditor.WrapPolicy = text.WrapGraphemes
	return page
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
		Name: "List Disks",
		Icon: icon.RemoteIcon,
	}
}

func (p *Page) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
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
									return material.Editor(th, &p.remoteIpInput, "remote ip address").Layout(gtx)
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
									return material.Editor(th, &p.usernameInput, "user name").Layout(gtx)
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
									return material.Editor(th, &p.passwordInput, "password").Layout(gtx)
								})
							})
						}),
					)
				})
			},
		),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// 点击按钮逻辑
			if p.execButton.Clicked(gtx) {
				p.checkInput()
				if !p.showDialog {
					p.executeCmd()
				}
			}
			return Button(gtx, 80, th, &p.execButton, "execute")
		}),
		// 结果显示区域（占满剩余空间）
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			in := layout.UniformInset(unit.Dp(8))
			return in.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.drawTable(gtx, th)
			})
		}),
	)

	// 弹出对话框
	if p.showDialog {
		p.drawConfirmDialog(gtx, th)
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
	}
	if len(itemName) != 0 {
		p.confirmMsg = fmt.Sprintf("%s is required", itemName)
		p.showDialog = true
	}
}

func (p *Page) drawConfirmDialog(gtx layout.Context, th *material.Theme) {
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
			layout.Rigid(material.Body1(th, p.confirmMsg).Layout),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(th, &p.modalButton, "confirm")
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

	output, err := session.CombinedOutput(command.Lsblk)
	if err != nil {
		p.confirmMsg = fmt.Sprintf("execute command failed, %v", err)
		p.showDialog = true
		return
	}
	blocks, err := command.GetBlockDevices(output)
	if err != nil {
		p.confirmMsg = err.Error()
		p.showDialog = true
		return
	}
	p.devices = blocks
}

func Button(gtx layout.Context, width unit.Dp, th *material.Theme, wid *widget.Clickable, txt string) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Dp(width)
	gtx.Constraints.Max.X = gtx.Dp(width)
	return material.Button(th, wid, txt).Layout(gtx)
}

// 绘制表格
func (p *Page) drawTable(gtx layout.Context, th *material.Theme) layout.Dimensions {

	fx := layout.Flex{Axis: layout.Vertical}

	// 设置表格样式
	const (
		colWidth  = 50
		rowHeight = 30
	)
	rows := make([]layout.FlexChild, 0)
	rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutTableCell(gtx, th, "No", colWidth, rowHeight, true)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutTableCell(gtx, th, "Name", colWidth, rowHeight, true)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutTableCell(gtx, th, "Type", colWidth, rowHeight, true)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutTableCell(gtx, th, "Size", colWidth, rowHeight, true)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutTableCell(gtx, th, "Serial", colWidth, rowHeight, true)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutTableCell(gtx, th, "Vendor", colWidth, rowHeight, true)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutTableCell(gtx, th, "Model", colWidth, rowHeight, true)
			}),
		)
	}))

	// 绘制表格行
	for i, dev := range p.devices {
		rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutTableCell(gtx, th, strconv.Itoa(i), colWidth, rowHeight, i%2 == 0)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutTableCell(gtx, th, dev.Name, colWidth, rowHeight, i%2 == 0)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					devType := "HDD"
					if !dev.Rota {
						devType = "SSD"
					}
					return layoutTableCell(gtx, th, devType, colWidth, rowHeight, i%2 == 0)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutTableCell(gtx, th, strconv.Itoa(int(dev.Size)), colWidth, rowHeight, i%2 == 0)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutTableCell(gtx, th, dev.Serial, colWidth, rowHeight, i%2 == 0)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutTableCell(gtx, th, dev.Vendor, colWidth, rowHeight, i%2 == 0)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layoutTableCell(gtx, th, dev.Model, colWidth, rowHeight, i%2 == 0)
				}),
			)
		}))
	}
	return fx.Layout(gtx, rows...)
}

func layoutTableCell(gtx layout.Context, th *material.Theme, text string, width, height unit.Dp, isHeader bool) layout.Dimensions {
	// 设置单元格样式
	lbl := material.Label(th, unit.Sp(16), text)
	if isHeader {
		lbl.Font.Weight = font.Bold
	}

	// 绘制单元格背景
	if !isHeader {
		if gtx.Constraints.Min.Y%2 == 0 {
			widget.Border{Color: th.Palette.Fg, Width: unit.Dp(0.5)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: gtx.Constraints.Min}
			})
		}
	}

	// 绘制文本
	return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X = gtx.Dp(width)
		gtx.Constraints.Min.Y = gtx.Dp(height)
		return lbl.Layout(gtx)
	})
}
