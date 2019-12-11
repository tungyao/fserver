package file

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

var con_type map[string]string = map[string]string{
	"png": "image/png",
	"jpg": "image/jpeg",
	"svg": "text/xml",
	"zip": "application/x-zip-compressed",
	"mp4": "video/mpeg4",
	"pdf": "application/pdf",
	"avi": "video/avi",
	"mp3": "audio/mp3"}

func Start() {
	client()
}
func client() {
	l, err := net.Listen("tcp", ":7777")
	if err != nil {
		log.Println(err)
	}
	for {
		con, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handle(con)
	}
}
func handle(conn net.Conn) {
	cache := make([]byte, 2048000)
	n, err := conn.Read(cache)
	if err != nil && err != io.EOF {
		log.Println(err, 1)
	}
	if len(cache[:n]) >= 14 {
		// HTTP protocol
		a, _ := SplitString(cache[:n], []byte("\n"))
		if Equal(a[0][len(a[0])-9:], []byte{72, 84, 84, 80, 47, 49, 46, 49, 13}) {
			url, _ := SplitString(a[0], []byte(" "))
			fmt.Println(string(url[1]))
			// http upload file
			if Equal(url[1], []byte("/upload")) {
				var filename []byte
				po := 0
				b := true
				for k, v := range a {
					if len(v) >= 12 && Equal(v[:12], []byte("Content-Type")) && b {
						m, _ := SplitString(v[15:], []byte("="))
						filename = m[1]
						b = false
						continue
					}
					if len(filename)+2 != len(v) {
						continue
					}
					if "--"+string(filename) == string(v) {
						//fmt.Println(string(v))
						m, _ := SplitString(a[k+1], []byte("filename="))
						filename = m[1]
						filename = filename[1 : len(filename)-2]
						fmt.Println("file name :" + string(filename))
						po = len(v)
						break
					}
				}
				c, p := SplitString(cache[:n], []byte("\r\n\r\n"))
				if len(c) >= 4 {
					_ = ioutil.WriteFile(string(filename), cache[:n][p[2]:n-(po+2)], 777)
					toHttpError(conn, "200 OK", "text/html")
				}
				toError(conn, "415 Unsupported Media Type", "text/html")
			}
			// http download file
			contype, _ := SplitString(url[1], []byte("."))
			if con_type[string(contype[len(contype)-1:][0])] != "" {
				tp := ""
				if con_type[string(contype[len(contype)-1:][0])] == "" {
					tp = "application/octet-stream"
				} else {
					tp = con_type[string(contype[len(contype)-1:][0])]
				}
				fs, err := os.Open("." + string(url[1]))
				if err != nil {
					log.Println(err)
					toHttpError(conn, "404 Not Found", tp)
				} else {
					conn.Write([]byte("HTTP/1.1 " + "200" + " OK\r\n"))
					conn.Write([]byte("Server: FileServer\r\n"))
					conn.Write([]byte("Date: " + time.Now().String() + "\r\n"))
					conn.Write([]byte("Content-Type: " + tp + "\r\n\r\n"))
					cache = make([]byte, 40960)
					for {
						n, err := fs.Read(cache)
						if err == io.EOF || n == 0 {
							break
						}
						conn.Write(cache[:n])
					}
					if conn != nil {
						_ = conn.Close()
					}
				}
			}
		} else {
			// TCP protocol , this is pretty faster
			// 0-128 is filename area,its a built-in protocol,use 0 to end
			// true length 129
			filename := cache[:n][:128]
			for k, v := range filename {
				if v == 0 {
					filename = cache[:k]
					break
				}
			}
			err := ioutil.WriteFile(string(filename), cache[129:n], 777)
			if err != nil {
				conn.Write([]byte("error"))
			}
			conn.Write([]byte("success"))
			conn.Close()
		}
	}
}
func toHttpError(conn net.Conn, ok string, ty string) {
	if conn != nil {
		conn.Write([]byte("HTTP/1.1 " + ok + "\r\n"))
		conn.Write([]byte("Server: FileServer\r\n"))
		conn.Write([]byte("Date: " + time.Now().String() + "\r\n"))
		conn.Write([]byte("Content-Type: " + ty + "\r\n\r\n"))
		conn.Write([]byte("<h1>SUCCESS</h1>"))
		conn.Close()
	}
}
func toError(conn net.Conn, ok string, ty string) {
	if conn != nil {
		conn.Write([]byte("HTTP/1.1 " + ok + "\r\n"))
		conn.Write([]byte("Server: FileServer\r\n"))
		conn.Write([]byte("Date: " + time.Now().String() + "\r\n"))
		conn.Write([]byte("Content-Type: " + ty + "\r\n\r\n"))
		conn.Write([]byte("<h1>ERROR</h1>"))
		conn.Close()
	}
	return
}
func Equal(one []byte, two []byte) bool {
	if len(one) != len(two) {
		return false
	}
	for k, v := range one {
		if (v) != two[k] {
			return false
		}
	}
	return true
}
func SplitString(str []byte, p []byte) ([][]byte, []int) {
	group := make([][]byte, 0)
	postion := make([]int, 0)
	ps := 0
	for i := 0; i < len(str); i++ {
		if str[i] == p[0] && i < len(str)-len(p) {
			if len(p) == 1 {
				group = append(group, str[ps:i])
				postion = append(postion, ps)
				ps = i + len(p)
			} else {
				for j := 1; j < len(p); j++ {
					if str[i+j] != p[j] || j != len(p)-1 {
						continue
					} else {
						group = append(group, str[ps:i])
						postion = append(postion, ps)
						ps = i + len(p)
					}
				}
			}
		} else {
			continue
		}
	}
	group = append(group, str[ps:])
	return group, postion
}
