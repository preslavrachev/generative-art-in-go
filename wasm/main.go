package main

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"syscall/js"

	"github.com/preslavrachev/generative-art-in-go/sketch"
)

type app struct {
	console            js.Value
	loadImageFunc      js.Func
	startRenderingFunc js.Func
	inputBuffer        []uint8
	sourceImg          image.Image
	done               chan struct{}
}

func newApp() *app {
	app := app{}
	app.setup()

	return &app
}

func (app *app) setup() {
	app.console = js.Global().Get("console")
	app.loadImageFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		array := args[0]
		app.inputBuffer = make([]uint8, array.Get("byteLength").Int())
		js.CopyBytesToGo(app.inputBuffer, array)

		app.console.Call("log", array.Get("byteLength").Int())

		reader := bytes.NewReader(app.inputBuffer)
		var err error
		app.sourceImg, _, err = image.Decode(reader)
		if err != nil {
			app.console.Call("log", err.Error())
			return nil
		}
		return nil
	})

	app.startRenderingFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		app.console.Call("log", "creating context")
		destWidth := app.sourceImg.Bounds().Dx()
		s := sketch.NewSketch(app.sourceImg, sketch.UserParams{
			StrokeRatio:              0.75,
			DestWidth:                destWidth,
			DestHeight:               app.sourceImg.Bounds().Dy(),
			InitialAlpha:             0.1,
			StrokeReduction:          0.002,
			AlphaIncrease:            0.06,
			StrokeInversionThreshold: 0.05,
			StrokeJitter:             int(0.1 * float64(destWidth)),
			MinEdgeCount:             3,
			MaxEdgeCount:             4,
		})
		app.console.Call("log", "drawing starts")
		for i := 0; i < 5000; i++ {
			s.Update()
			if i%500 == 0 {
				app.updateImage(s.Output())
			}
			if i%100 == 0 {
				app.trackProgress(int(float64(i) / 5000 * 100))
			}
		}

		app.console.Call("log", "drawing done")
		app.trackProgress(100)
		return nil
	})
}

func (app *app) trackProgress(progress int) {
	js.Global().Call("trackProgress", progress)
}

func (app *app) updateImage(img image.Image) {
	buf := bytes.NewBuffer(make([]byte, 0))
	png.Encode(buf, img)
	dst := js.Global().Get("Uint8Array").New(len(buf.Bytes()))
	js.CopyBytesToJS(dst, buf.Bytes())
	js.Global().Call("displayImage", dst)
}

func main() {
	app := newApp()
	js.Global().Set("loadImage", app.loadImageFunc)
	js.Global().Set("startRendering", app.startRenderingFunc)
	<-app.done
}
