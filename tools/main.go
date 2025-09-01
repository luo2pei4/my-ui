package main

import (
	"log"
	"os"

	page "tools/pages"
	"tools/pages/home"
	remotessh "tools/pages/remote_ssh"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
)

func main() {
	go func() {
		win := new(app.Window)
		win.Option(app.Title("tools"), app.Maximized.Option())
		if err := loop(win); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(win *app.Window) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	var ops op.Ops

	router := page.NewRouter()
	router.Register("home", home.New(&router))
	router.Register("remote", remotessh.New(&router))

	for {
		switch e := win.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			router.Layout(gtx, th)
			e.Frame(gtx.Ops)
		}
	}
}
