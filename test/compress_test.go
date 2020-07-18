package test

import (
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"testing"
)

func compressImageResource(reader io.Reader, writer io.Writer) {
	img, _, err := image.Decode(reader)
	if err != nil {
		log.Panicln(err)
	}
	err = jpeg.Encode(writer, img, &jpeg.Options{Quality: 40})
	if err != nil {
		log.Panicln(err)
	}
}
func TestCompress(t *testing.T) {
	fs, err := os.Open("../mount/d41d8cd98f00b204e9800998ecf8427e.jpg")
	if err != nil {
		log.Panicln(err)
	}
	fw, _ := os.OpenFile("test.jpg", os.O_CREATE, 666)
	compressImageResource(fs, fw)
}
