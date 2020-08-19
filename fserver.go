package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
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
	"mov":  "video/quicktime",
	"js":   "application/x-javascript",
	"css":  "text/css",
}
var logg *log.Logger

const (
	MOUNT   = "./mount/"
	QUALITY = "./quality/"
	LOG     = "./log/fserver.log"
)

var (
	DOMAIN string
	USER   string
	PASS   string
)

func init() {
	flag.Parse()
	flag.StringVar(&DOMAIN, "domino", "https://you_domino/", "")
	flag.StringVar(&USER, "user", "you_name", "")
	flag.StringVar(&PASS, "pass", "you_pass", "")
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
func BasicAuth(writer http.ResponseWriter, request *http.Request) bool {
	user, pass, ok := request.BasicAuth()
	if !ok {
		writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		writer.WriteHeader(http.StatusUnauthorized)
		return true
	}
	if user != USER && pass != PASS {
		http.Error(writer, " need authorized!", http.StatusUnauthorized)
		return true
	}
	return false
}
func OutHtml() {

}

type FileDirs struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

// 压缩图片
func compressImageResource(filename string, q int, reader io.Reader) string {
	if Last(filename) == "jpg" { // jpg
		img, _, err := image.Decode(reader)
		if err != nil {
			return MOUNT + filename
		}
		fname := strconv.Itoa(q) + "_" + filename
		if fs, err := os.Open(QUALITY + fname); err != nil {
			if os.IsExist(err) {
				return MOUNT + filename
			}
			fs.Close()
			fs, err := os.OpenFile(QUALITY+fname, os.O_CREATE|os.O_RDWR, 666)
			if err != nil {
				return MOUNT + filename
			}
			err = jpeg.Encode(fs, img, &jpeg.Options{Quality: q})
			fs.Close()
			if err != nil {
				return MOUNT + filename
			}
			return QUALITY + fname
		} else {
			fs.Close()
			return QUALITY + fname
		}
	}
	if Last(filename) == "png" {
		imgSrc, _, err := image.Decode(reader)
		if err != nil {
			log.Panicln(err)
			return MOUNT + filename
		}
		fname := strconv.Itoa(q) + "_" + filename
		if fs, err := os.Open(QUALITY + fname); err != nil {
			if os.IsExist(err) {
				log.Panicln(err)

				return MOUNT + filename
			}
			fs.Close()
			fs, err := os.OpenFile(QUALITY+fname, os.O_CREATE|os.O_RDWR, 666)
			if err != nil {
				log.Panicln(err)

				return MOUNT + filename
			}
			newImg := image.NewRGBA(imgSrc.Bounds())
			draw.Draw(newImg, newImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
			draw.Draw(newImg, newImg.Bounds(), imgSrc, imgSrc.Bounds().Min, draw.Over)
			err = jpeg.Encode(fs, newImg, &jpeg.Options{Quality: q})
			if err != nil {
				return MOUNT + filename
			}
			fs.Close()
			return QUALITY + fname
		} else {
			fs.Close()
			return QUALITY + fname
		}

	}
	return MOUNT + filename
}
func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/" {
			// 校验用户名
			if BasicAuth(writer, request) {
				return
			}
			arr := make([]FileDirs, 0)
			filepath.Walk(MOUNT[:len(MOUNT)-1], func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() {
					truePath := path[len(MOUNT)-2:]
					arr = append(arr, FileDirs{
						Name: truePath,
						Url:  DOMAIN + truePath,
					})
				}
				return nil
			})
			pf, _ := template.ParseFiles("./dist/index.html")
			pf.Execute(writer, arr)
			return
		}
		writer.Header().Set("Cache-Control", "max-age=604800")
		writer.Header().Add("Accept-Ranges", "bytes")
		fileName := request.URL.Path
		// 获取到质量信息
		fs, err := os.Open(MOUNT[:len(MOUNT)-1] + fileName)
		defer fs.Close()
		if os.IsNotExist(err) {
			http.NotFound(writer, request)
			return
		} else {
			if que := request.URL.Query().Get("quality"); que != "" {
				quality, err := strconv.Atoi(que)
				if err == nil {
					fpath := compressImageResource(fileName[1:], quality, fs)
					fs.Close()
					fs, err = os.Open(fpath)
					if err != nil {
						log.Panicln(err)
					}
				}
			}
			var start, end int64
			suffix, _ := SplitString([]byte(fileName), []byte("."))
			fi, _ := fs.Stat()
			writer.Header().Add("Etag", `T/"`+string(suffix[0][1:])+`"`)
			writer.Header().Add("Last-Modify", fi.ModTime().Format(time.RFC1123))
			fmt.Println(Last(fi.Name()))
			writer.Header().Add("Content-Type", conType[Last(fi.Name())])
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
			n := 4096000
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
			mountDir := request.FormValue("mount")
			if mountDir == "" {
				mountDir = request.URL.Query().Get("dir")
			}
			if mountDir != "" {
				mountDir += "/"
			}
			md5 := md5.New()
			tr := io.TeeReader(file, md5)
			MD5Str := hex.EncodeToString(md5.Sum([]byte(strconv.Itoa(int(head.Size)))))
			var shaName = MD5Str + "." + string(suffix[len(suffix)-1])
			fs, err := os.Open(MOUNT + mountDir + shaName)
			defer file.Close()
			if os.IsNotExist(err) {
				fs.Close()
				os.Mkdir(MOUNT+request.FormValue("mount"), 666)
				fs, err := os.OpenFile(MOUNT+mountDir+shaName, os.O_CREATE|os.O_WRONLY, 666)
				if err != nil {
					logg.Panicln(err)

				}
				_, err = io.Copy(fs, tr)
				defer fs.Close()
				if err != nil {
					logg.Panicln(err)
				}
				logg.Println("save file :[" + head.Filename + "]\t" + "=>[" + MOUNT + mountDir + shaName + "]")
			}
			writer.Write([]byte(`{"t":1,"ok":"yes","msg":"success","url":"` + DOMAIN + mountDir + shaName + `"}`))
			writer.Header().Add("Content-Type", "application/json")
			return
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
</html>
			`))
		}
	})
	if err := http.ListenAndServe(":8105", nil); err != nil {
		logg.Panicln(err)
	}
}
