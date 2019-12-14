package test

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
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
	fmt.Println(sha([]byte("123456")))
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
