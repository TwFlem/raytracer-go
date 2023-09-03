package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {
	fname := "out/hello-gradient.ppm"
	f, err := os.Open(fname)
	if err != nil {
		f, err = os.Create(fname)
		if err != nil {
			panic(err)
		}
	}
	w := 256
	h := 256
	ppm := []string{
		"P3",
		strconv.Itoa(w) + " " + strconv.Itoa(h),
		"255",
	}
	for j := 0; j < h; j++ {
		fmt.Printf("computing %d of %d\n", j, h-1)
		for i := 0; i < w; i++ {
			r := math.Floor(float64(i) / float64((w - 1)) * 256)
			g := math.Floor(float64(j) / float64((h - 1)) * 256)
			b := 0
			ppm = append(ppm, fmt.Sprintf("%s %s %s", strconv.Itoa(int(r)), strconv.Itoa(int(g)), strconv.Itoa(int(b))))
		}
	}

	io.WriteString(f, strings.Join(ppm, "\n"))
}
