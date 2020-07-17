package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var conType = map[string]string{
	"png":  "image/png",
	"jpg":  "image/jpeg",
	"gif":  "image/gif",
	"svg":  "text/xml",
	"txt":  "text/plain",
	"zip":  "application/x-zip-compressed",
	"mp4":  "video/mpeg4",
	"pdf":  "application/pdf",
	"avi":  "video/avi",
	"mp3":  "audio/mp3",
	"json": "application/json",
	"gz":   "application/octet-stream",
	"tar":  "application/octet-stream",
	"7z":   "application/octet-stream",
	"ico":  "application/x-ico",
	"exe":  "application/x-msdownload",
}
var logg *log.Logger
var DOMAIN string = "http://123.207.198.60/"
var MOUNT string = "./mount/"
var LOG = "./log/fserver.log"

func init() {

	fs, errx := os.OpenFile(LOG, os.O_RDWR|os.O_CREATE|os.O_APPEND, 766)
	if errx != nil {
		log.Fatalln(errx)
	}
	logg = log.New(fs, "[fserver]", log.LstdFlags|log.Lshortfile|log.LUTC)
}

func sha(data string) string {
	t := sha1.New()
	t.Write([]byte(data))
	return fmt.Sprintf("%x", t.Sum(nil))
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
func Last(s string) string {
	a := strings.Split(s, ".")
	return a[len(a)-1]
}
func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/" {
			html := "<table><tr><td>name</td><td>operation</td><td>url</td></tr>"
			filepath.Walk(MOUNT[:len(MOUNT)-1], func(path string, info os.FileInfo, err error) error {
				if x := conType[Last(info.Name())]; x != "" {
					if x == "image/png" || x == "image/jpeg" || x == "audio/mp3" {
						html += "<tr><td><img width='100' height='100' src='" + DOMAIN + path[len(MOUNT):] + "'/><p>" + path[len(MOUNT):] + "</p></td><td><a href='/" + path[len(MOUNT):] + "'>download</a></td><td><input value='" + DOMAIN + path[len(MOUNT):] + "'/></td></tr>"
					} else {
						html += "<tr><td>" + info.Name() + "</td><td><a href='/" + path[len(MOUNT):] + "'>download</a></td><td><input value='" + DOMAIN + path[len(MOUNT):] + "'/></td></tr>"

					}
				}
				return nil
			})
			html += "</table>"
			writer.Write([]byte(html))
			return
		}
		writer.Header().Set("Cache-Control", "max-age=604800")
		writer.Header().Add("Accept-Ranges", "bytes")
		fileName := request.URL.Path
		fs, err := os.Open(MOUNT[:len(MOUNT)-1] + fileName)
		defer fs.Close()
		if os.IsNotExist(err) {
			http.NotFound(writer, request)
			return
		} else {
			var start, end int64
			suffix, _ := SplitString([]byte(fileName), []byte("."))
			fi, _ := fs.Stat()
			writer.Header().Add("Etag", `T/"`+string(suffix[0][1:])+`"`)
			writer.Header().Add("Last-Modify", fi.ModTime().Format(time.RFC1123))
			writer.Header().Set("Content-Type", conType[Last(fi.Name())])
			if cc := request.Header.Get("Cache-Control"); cc != "" && cc != "no-cache" {
				writer.WriteHeader(304)
			}
			if r := request.Header.Get("Range"); r != "" {
				if strings.Contains(r, "bytes=") && strings.Contains(r, "-") {

					fmt.Sscanf(r, "bytes=%d-%d", &start, &end)
					if end == 0 {
						end = fi.Size() - 1
					}
					if start > end || start < 0 || end < 0 || end >= fi.Size() {
						writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
						logg.Println("sendFile2 start:", start, "end:", end, "size:", fi.Size())
						writer.WriteHeader(http.StatusBadRequest)
						return
					}
					writer.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
					writer.Header().Add("Content-Range", fmt.Sprintf("bytes %v-%v/%v", start, end, fi.Size()))
				} else {
					writer.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
					writer.WriteHeader(http.StatusBadRequest)
					return
				}
				writer.WriteHeader(206)
			} else {
				writer.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
				start = 0
				end = fi.Size() - 1
			}
			_, err = fs.Seek(start, 0)
			if err != nil {
				logg.Println("sendFile3", err.Error())
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			n := 409600
			buf := make([]byte, n)
			for {
				if end-start+1 < int64(n) {
					n = int(end - start + 1)
				}
				_, err := fs.Read(buf[:n])
				if err != nil {
					if err != io.EOF {
						logg.Println("error:", err)
					}
					return
				}
				err = nil
				_, err = writer.Write(buf[:n])
				if err != nil {
					// log.Println(err, start, end, info.Size(), n)
					return
				}
				start += int64(n)
				if start >= end+1 {
					return
				}
			}
			// io.Copy(writer, fs)
		}
	})
	http.HandleFunc("/upload", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "POST" {
			writer.Header().Set("Access-Control-Allow-Origin", "*")
			writer.Header().Set("Access-Control-Allow-Methods", "POST")
			request.ParseMultipartForm(204800000)
			// request.ParseForm()
			// mr,err:=request.MultipartReader()
			// from,err:=mr.ReadForm(204800000)
			// frs:=from.File["file"]
			// fmt.Println(frs)
			file, head, err := request.FormFile("file")
			if err != nil {
				// log.Panicln(err)
				file, head, err = request.FormFile("filepond")
				if err != nil {
					writer.WriteHeader(503)
					return
				} else {
					goto safe
				}
				writer.WriteHeader(503)
				return
			}
			goto safe
		safe:
			suffix, _ := SplitString([]byte(head.Filename), []byte("."))
			var shaName = sha(head.Filename+time.Now().String()) + "." + string(suffix[len(suffix)-1])
			fs, err := os.Open(MOUNT + request.FormValue("mount") + shaName)
			if os.IsNotExist(err) {
				fs.Close()
				os.Mkdir(MOUNT+request.FormValue("mount"), 666)
				fs, err := os.OpenFile(MOUNT+request.FormValue("mount")+shaName, os.O_CREATE|os.O_WRONLY, 666)
				if err != nil {
					logg.Panicln(err)

				}
				_, err = io.Copy(fs, file)
				defer fs.Close()
				if err != nil {
					logg.Panicln(err)
				}
				logg.Println("save file :[" + head.Filename + "]\t" + "=>[" + shaName + "]")
				//AddFile(head.Filename, shaName)
			}

			writer.Header().Add("Content-Type", "application/json")
			writer.Write([]byte(`{"t":1,"ok":"yes","msg":"success","url":"` + DOMAIN + request.FormValue("mount") + shaName + `"}`))
		}
		if request.Method == "GET" {
			writer.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <title>FilePond from CDN</title>
<link href="` + DOMAIN + `40abc96eb3a5d4a496083dcb0ac4b6d58515537d.css" rel="stylesheet">

</head>
<body>
  <input type="file" class="filepond">
<script src="` + DOMAIN + `5e4ec15f0335954186c76f7711f6ab9f1c0be308.js"></script>
  <script>
 FilePond.parse(document.body);
FilePond.setOptions({
    server: '/upload'
});
  </script>

</body>
</html>`))
		}
	})
	if err := http.ListenAndServe(":7777", nil); err != nil {
		logg.Panicln(err)
	}
}
