package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type E2Image struct {
	file   *os.File
	image  image.Image
	bounds image.Rectangle
}

var images map[string]E2Image

func getImage(name string) (E2Image, bool) {
	img, ok := images[name]
	if ok {
		return img, true
	} else {
		img := E2Image{}
		file, err := os.Open(name)
		if err != nil {
			return img, false
		}
		img.file = file

		m, _, err := image.Decode(file)
		if err != nil {
			file.Close()
			return img, false
		}
		img.image = m
		img.bounds = m.Bounds()

		return img, true
	}
}

func convertPixel(r uint32, g uint32, b uint32) (uint32, uint32, uint32) {
	return uint32(float32(r) / 65535.0 * 255.0), uint32(float32(g) / 65535.0 * 255.0), uint32(float32(b) / 65535.0 * 255.0)
}

func handleImage(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	img, ok := getImage(name)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error")
	}

	action := r.URL.Query().Get("action")
	switch action {
	case "resolution":
		fmt.Fprintf(w, "%d,%d", img.bounds.Max.X, img.bounds.Max.Y)
	case "pixels":
		x, _ := strconv.Atoi(r.URL.Query().Get("x"))
		y, _ := strconv.Atoi(r.URL.Query().Get("y"))
		count, _ := strconv.Atoi(r.URL.Query().Get("count"))
		width := img.bounds.Max.X

		mode, _ := strconv.Atoi(r.URL.Query().Get("mode"))
		pixels := make([]string, count)
		for i := 0; i < count; i++ {
			xPos := (x + i) % width
			yPos := y
			if x+i >= width {
				yPos = y + 1
			}

			r32, g32, b32, _ := img.image.At(xPos, yPos).RGBA()
			r, g, b := convertPixel(r32, g32, b32)
			switch mode {
			case 2:
				pixels[i] = strconv.FormatUint((uint64)((r*65536)+(g*256)+b), 10)
			case 3:
				pixels[i] = fmt.Sprintf("%d%d%d", r, g, b)
			default:
				pixels[i] = strconv.FormatUint((uint64)((r*65536)+(g*256)+b), 10)
			}
		}

		fmt.Fprint(w, strings.Join(pixels, ","))
	}
}

func main() {
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "RemoteAddr: %s\n", r.RemoteAddr)
		for name, headers := range r.Header {
			for _, value := range headers {
				fmt.Fprintf(w, "%s: %s\n", name, value)
			}
		}
	})

	http.HandleFunc("/buffer", func(w http.ResponseWriter, r *http.Request) {
		size, err := strconv.Atoi(r.URL.Query().Get("size"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid parameter")
		}

		buffer := make([]byte, size)
		for i := 0; i < size; i++ {
			buffer[i] = 'A'
		}

		w.Write(buffer)
	})

	http.HandleFunc("/image", handleImage)

	http.ListenAndServe(":1337", nil)
}
