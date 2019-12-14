package fserver

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var con_type map[string]string = map[string]string{
	"png": "image/png",
	"jpg": "image/jpeg",
	"svg": "text/xml",
	"txt": "text/plain",
	"zip": "application/x-zip-compressed",
	"mp4": "video/mpeg4",
	"pdf": "application/pdf",
	"avi": "video/avi",
	"mp3": "audio/mp3"}

func Start(youserver string) {
	client(youserver)
}
func client(youserver string) {
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
		go handle(con, youserver)
	}
}
func handle(conn net.Conn, youserver string) {
	cache := make([]byte, 2048000)
	n, err := conn.Read(cache)
	if err != nil && err != io.EOF {
		log.Println(err, 1)
	}
	//fmt.Println((cache[:n]))
	if len(cache[:n]) >= 14 {
		// HTTP protocol
		a, _ := SplitString(cache[:n], []byte("\n"))
		//fmt.Println(a[0][len(a[0])-9:])
		if Equal(a[0][len(a[0])-9:len(a[0])-3], []byte{72, 84, 84, 80, 47, 49}) { // HTTP/1
			url, _ := SplitString(a[0], []byte(" "))
			//fmt.Println(12313213)
			//fmt.Println(url)
			//fmt.Println(string(url[1]))
			//fmt.Println("---------------------------------")
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
						fmt.Println("up file :" + string(filename))
						po = len(v)
						break
					}
				}
				contype, _ := SplitString(filename, []byte("."))
				//fmt.Println(len(a))
				//for _, v := range a {
				//	fmt.Println((v))
				//	fmt.Println("----------------------------------------------------")
				//}
				if Equal(contype[len(contype)-1], []byte("png")) {
					a, p := SplitString(cache[:n], []byte{137, 80, 78, 71, 13, 10})
					//fmt.Println(len(a))
					if len(a) >= 2 {
						//fmt.Println(cache[:n][p[len(p)-1]-6:len(cache[:n])-po-2])
						filen := sha(cache[:n][p[1]-6 : len(cache[:n])-po-3])
						fs, err := os.OpenFile(filen+"."+string(contype[len(contype)-1]), os.O_CREATE|os.O_WRONLY, 666)
						if err != nil {
							log.Println(err)
						}
						fs.Write(cache[:n][p[1]-6 : len(cache[:n])-po-3])
						fs.Close()
						if err != nil {
							log.Println(err)
						}
						toJson(conn, "200 OK", `{"url":"`+youserver+filen+"."+string(contype[len(contype)-1])+`"}`)
						//toHttpError(conn, "200 OK", "application/json")
						return
					}
				}
				if len(a) >= 2 {
					_, p := SplitString(cache[:n], []byte{13, 10, 13, 10})
					filen := sha(cache[:n][p[1]-6 : len(cache[:n])-po-3])
					fs, err := os.OpenFile(filen+"."+string(contype[len(contype)-1]), os.O_CREATE|os.O_WRONLY, 666)
					if err != nil {
						log.Println(err)
					}
					fs.Write(cache[:n][p[len(p)-1]:])
					fs.Close()
					toJson(conn, "200 OK", `{"url":"`+youserver+filen+"."+string(contype[len(contype)-1])+`"}`)
					return
				}
				toJson(conn, "415 Unsupported Media Type", ``)
				return
			}
			// http download file
			contype, _ := SplitString(url[1], []byte("."))
			//fmt.Println(string(contype[1]))
			if con_type[string(contype[1])] != "" {
				//fmt.Println(123123)
				tp := ""
				if con_type[string(contype[len(contype)-1:][0])] == "" {
					tp = "application/octet-stream"
				} else {
					tp = con_type[string(contype[len(contype)-1:][0])]
				}
				//fs, err := os.Open("." + string(url[1]))
				fs, err := os.OpenFile("."+string(url[1]), os.O_RDONLY, 666)
				//fmt.Println(tp)
				fmt.Println("get file :" + string(url[1]))
				if err != nil {
					log.Println(err)
					toHttpError(conn, "404 Not Found", tp)
				} else {
					conn.Write([]byte("HTTP/1.1 " + "200" + " OK\r\n"))
					conn.Write([]byte("Server: FileServer\r\n"))
					conn.Write([]byte("Date: " + time.Now().String() + "\r\n"))
					conn.Write([]byte("Content-Type: " + tp + "\r\n\r\n"))
					//cache = make([]byte, 40960)
					for {
						n, err := fs.Read(cache)
						if err == io.EOF || n == 0 {
							break
						}
						conn.Write(cache[:n])
					}
					fs.Close()
					conn.Close()
					//fs.Close()
					return
				}
			}
		} else {
			// TCP protocol , this is pretty faster
			// 0-128 is filename area,its a built-in protocol,use 0 to end
			// true length 129
			//fmt.Println(123)
			filename := cache[:n][:128]
			if len(filename) < 5 {
				conn.Write([]byte("error"))
				conn.Close()
				return
			}
			for k, v := range filename {
				if v == 0 {
					filename = cache[:k]
					break
				}
			}
			fmt.Println("tpc up file:", string(filename))
			contype, _ := SplitString(filename, []byte("."))
			filen := sha(cache[128:n])
			fs, err := os.OpenFile(filen+"."+string(contype[len(contype)-1]), os.O_CREATE|os.O_WRONLY, 666)
			if err != nil {
				log.Println(err)
			}
			fs.Write(cache[128:n])
			fs.Close()
			conn.Write([]byte(youserver + filen))
			conn.Close()
		}
	}
}
func toJson(conn net.Conn, ok string, body string) {
	if conn != nil {
		conn.Write([]byte("HTTP/1.1 " + ok + "\r\n"))
		conn.Write([]byte("Server: FileServer\r\n"))
		conn.Write([]byte("Date: " + time.Now().String() + "\r\n"))
		conn.Write([]byte("Content-Type: application/json\r\n\r\n"))
		conn.Write([]byte(body))
		conn.Close()
	}
	return
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
	return
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
		if v != two[k] {
			return false
		}
	}
	return true
}
func SplitString(str []byte, p []byte) ([][]byte, []int) {
	group := make([][]byte, 0)
	portion := make([]int, 0)
	ps := 0
	for i := 0; i < len(str); i++ {
		if str[i] == p[0] && i < len(str)-len(p) {
			if len(p) == 1 {
				group = append(group, str[ps:i])
				portion = append(portion, ps)
				ps = i + len(p)
			} else {
				for j := 1; j < len(p); j++ {
					if str[i+j] != p[j] || j != len(p)-1 {
						continue
					} else {
						group = append(group, str[ps:i])
						portion = append(portion, ps)
						ps = i + len(p)
					}
				}
			}
		} else {
			continue
		}
	}
	group = append(group, str[ps:])
	portion = append(portion, ps)
	return group, portion
}
func sha(data []byte) string {
	t := sha1.New()
	t.Write(data)
	return fmt.Sprintf("%x", t.Sum(nil))
}
