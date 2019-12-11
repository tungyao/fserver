package test

import (
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"
)
import "../../file"

func TestFile(t *testing.T) {
	file.Start()
}
func TestSp(t *testing.T) {
	fs, _ := os.Open("abcd.jpg")
	data, _ := ioutil.ReadAll(fs)
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
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
