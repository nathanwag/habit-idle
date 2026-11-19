// Gera os ícones do app (web/assets/icons): o Rasko na porta da torre, em pixel
// art 36×36 ampliado sem suavizar. Rode da raiz: go run ./scripts/icon
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

const n = 36

func hex(s string) color.RGBA {
	var r, g, b uint8
	fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
	return color.RGBA{r, g, b, 255}
}

var (
	k2   = hex("#10141f")
	s2   = hex("#202e37")
	s3   = hex("#394a50")
	s4   = hex("#577277")
	s5   = hex("#819796")
	gold = hex("#e8c170")
	g2   = hex("#de9e41")
	g3   = hex("#be772b")
	red  = hex("#cf573c")
	br   = hex("#884b2b")
	warm = []color.RGBA{hex("#4a3524"), hex("#2e2420"), k2}
)

func main() {
	f, err := os.Open("web/assets/sprites/rasko.png")
	if err != nil {
		log.Fatal(err)
	}
	sprite, err := png.Decode(f)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	art := draw(sprite)
	for _, s := range []int{180, 192, 512} {
		if err := save(fmt.Sprintf("web/assets/icons/icon-%d.png", s), scale(art, s)); err != nil {
			log.Fatal(err)
		}
	}
}

func draw(sprite image.Image) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	rect := func(x, y, w, h int, c color.RGBA) {
		for j := y; j < y+h; j++ {
			for i := x; i < x+w; i++ {
				if i >= 0 && j >= 0 && i < n && j < n {
					img.SetRGBA(i, j, c)
				}
			}
		}
	}

	// Parede de tijolos.
	rect(0, 0, n, n, s3)
	for y := 0; y < n; y += 5 {
		rect(0, y, n, 1, s2)
		start := 8
		if (y/5)%2 == 1 {
			start = 2
		}
		for x := start; x < n; x += 12 {
			rect(x, y, 1, 5, s2)
		}
	}

	// Porta em arco: larga o bastante para o Rasko (16 px) ter folga dos lados.
	const cx, cy = 18.0, 15.0
	inside := func(x, y int, rx, ry float64) bool {
		dx, dy := float64(x)+.5-cx, float64(y)+.5-cy
		if float64(y) >= cy {
			return math.Abs(dx) <= rx
		}
		return (dx/rx)*(dx/rx)+(dy/ry)*(dy/ry) <= 1
	}
	for y := 0; y < 33; y++ {
		for x := 0; x < n; x++ {
			dx := float64(x) + .5 - cx
			switch {
			case inside(x, y, 11, 9):
				g := math.Hypot(dx, float64(y-31)*1.2) / 18
				c := warm[2]
				if g < .45 {
					c = warm[0]
				} else if g < .7 {
					c = warm[1]
				}
				img.SetRGBA(x, y, c)
			case inside(x, y, 12.5, 10.5):
				c := s4
				if dx < 0 {
					c = s5
				}
				img.SetRGBA(x, y, c)
			}
		}
	}
	rect(7, 33, 22, 3, g3)

	// Tochas dos lados.
	for _, t := range []struct {
		cx float64
		bx int
	}{{3, 2}, {33, 32}} {
		flame(img, t.cx, 9, 6)
		rect(t.bx, 15, 2, 5, br)
	}

	// Rasko parado (quadro 4 da fita 16×28).
	b := sprite.Bounds()
	for y := 0; y < 28; y++ {
		for x := 0; x < 16; x++ {
			c := color.RGBAModel.Convert(sprite.At(b.Min.X+64+x, b.Min.Y+y)).(color.RGBA)
			if c.A > 0 {
				img.SetRGBA(10+x, 5+y, c)
			}
		}
	}
	return img
}

func flame(img *image.RGBA, cx float64, top, h int) {
	for y := top; y < top+h; y++ {
		t := float64(y-top) / float64(h-1)
		w := 2.2 * t / .7
		if t >= .7 {
			w = 2.2 * (1 - (t-.7)/.3*.4)
		}
		for x := int(math.Floor(cx - w)); x <= int(math.Ceil(cx+w)); x++ {
			rel := math.Abs(float64(x)+.5-cx) / math.Max(w, .5)
			if rel > 1 || x < 0 || x >= n {
				continue
			}
			c := g2
			if rel < .45 && t > .4 {
				c = gold
			} else if t < .3 || rel > .8 {
				c = red
			}
			img.SetRGBA(x, y, c)
		}
	}
}

func scale(src *image.RGBA, s int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, s, s))
	for y := 0; y < s; y++ {
		for x := 0; x < s; x++ {
			dst.SetRGBA(x, y, src.RGBAAt(x*n/s, y*n/s))
		}
	}
	return dst
}

func save(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
