// Command genicon writes placeholder PWA icons (a beaker on blue) to the given directory.
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: genicon <outdir>")
	}
	dir := os.Args[1]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	for _, size := range []int{192, 512} {
		path := filepath.Join(dir, "icon-"+strconv.Itoa(size)+".png")
		if err := write(path, size); err != nil {
			log.Fatal(err)
		}
	}
}

func write(path string, s int) error {
	img := image.NewRGBA(image.Rect(0, 0, s, s))
	bg := color.RGBA{0x30, 0x60, 0xd0, 0xff}
	glass := color.RGBA{0xff, 0xff, 0xff, 0xff}
	slime := color.RGBA{0x7c, 0xf0, 0x8c, 0xff}
	f := float64(s)
	for y := 0; y < s; y++ {
		for x := 0; x < s; x++ {
			fx, fy := float64(x)/f, float64(y)/f
			c := bg
			// beaker neck
			if fx > 0.42 && fx < 0.58 && fy > 0.2 && fy < 0.45 {
				c = glass
			}
			// beaker body: triangle widening downward
			if fy >= 0.45 && fy < 0.8 {
				half := 0.08 + (fy-0.45)*0.6
				if fx > 0.5-half && fx < 0.5+half {
					c = glass
					if fy > 0.6 {
						c = slime
					}
				}
			}
			img.Set(x, y, c)
		}
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, img)
}
