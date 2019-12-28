package test

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"testing"
)
import "../../fserver"

func TestFile(t *testing.T) {
	fserver.Start("http://localhost:7777/")
}
func sha(data []byte) string {
	t := sha1.New()
	t.Write(data)
	return fmt.Sprintf("%x", t.Sum(nil))
}
func TestOther(t *testing.T) {
}
func TestImage(t *testing.T) {
	fs, err := os.Open("./ff7d0fbc486ac031f9da55c9607cc55b77ca77a1.png")
	if err != nil {
		log.Println(err)
	}
	img, _, err := image.Decode(fs)
	if err != nil {
		log.Println(err)
	}
	bounds := img.Bounds()
	fmt.Println(bounds.Size())
}
func TestSp(t *testing.T) {
	fs, _ := os.Open("abcd.png")
	data, _ := ioutil.ReadAll(fs)
	conn, err := net.Dial("tcp", "file.yaop.ink:443")
	if err != nil {
		log.Println(err)
	}
	d := make([]byte, 0)
	for _, v := range []byte("abcdd.png") {
		d = append(d, v)
	}
	for i := len([]byte("abcdd.png")); i < 128; i++ {
		d = append(d, 0)
	}
	for _, v := range data {
		d = append(d, v)
	}
	conn.Write(d)
	conn.Close()
}
func TestGetFormData(t *testing.T) {
	fs, err := os.Open("./upload")
	if err != nil {
		log.Println(err)
	}
	data, err := ioutil.ReadAll(fs)
	if err != nil {
		log.Println(err)
	}
	// boundary name
	var boundary []byte
	var boundaryPoint = false
	// file size
	var fileSize int
	var fileSizeSilce = make([]byte, 0)
	var fileSizePoint bool
	var seek = make([]int, 0)
	var trueData int = 0
	for k, v := range data {
		// find boundary name, out => boundary name
		if !boundaryPoint {
			if v == 'b' {
				if data[k+1] == 'o' && data[k+2] == 'u' && data[k+3] == 'n' {
					for _, c := range data[k+9:] {
						if c == '\r' {
							break
						}
						boundary = append(boundary, c)
					}
					boundaryPoint = true
				}
			}
		}
		// find file size,out => fileSize
		if !fileSizePoint {
			if v == 'C' {
				if data[k+8] == 'L' && data[k+9] == 'e' && data[k+10] == 'n' {
					for _, c := range data[k+16:] {
						if c == '\r' {
							break
						}
						fileSizeSilce = append(fileSizeSilce, c)
					}
					fileSizePoint = true
					fileSize, _ = strconv.Atoi(string(fileSizeSilce))
				}
			}
		}
		if boundaryPoint {
			if bytes.Equal(boundary, data[k:k+len(boundary)]) {
				seek = append(seek, k)
			}
			//		if len(seek) == 2 {
			//			for j, c := range data[k+len(boundary):] {
			//				if c == '\r' {
			//					d := data[k+len(boundary):]
			//					if d[j+1] == 'r' {
			//						fmt.Println(d)
			//						return
			//					}
			//				}
			//			}
			//		}
			//		seek = append(seek, k)
			//	}
		}
	}
	for i := 0; i < len(data[seek[1]:seek[2]]); i++ {
		if data[seek[1]:seek[2]][i] == 13 && data[seek[1]:seek[2]][i+1] == 13 {
			trueData = i + 2
			break
		}
	}
	fmt.Println(string(boundary))
	fmt.Println(seek)
	fmt.Println(trueData)
	fmt.Println(fileSize)
	// 239 191 189 239 191
	fmt.Println(data[seek[1]:seek[2]][trueData : seek[2]-len(boundary)])
	fs, _ = os.Create("a.jpg")
	fs.Write(data[seek[1]:seek[2]][trueData : seek[2]-len(boundary)])
	fs.Close()
}
