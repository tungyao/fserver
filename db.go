package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB
var errx error
var logg *log.Logger

func init() {
	fs,errx:=os.OpenFile("/var/log/fserver.log",os.O_RDWR|os.O_CREATE|os.O_APPEND,766)
	if errx != nil {
		log.Fatalln(errx)
	}
	logg = log.New(fs,"[fserver]",log.LstdFlags|log.Lshortfile|log.LUTC)
	logg.Println("hello")
}
type SaveTable struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	HashName   string `json:"hash_name"`
	CreateTime int64  `json:"create_time"`
}

func init() {
	DB, errx = sql.Open("mysql", fmt.Sprintf("%s:%s@%s(%s:%s)/%s", "fileuser", "Fileuser2232@", "tcp", "localhost", "3306", "file"))
	if errx != nil {
		logg.Fatalln(errx)
	}
}
func AddFile(name, hashName string) {
	stmt, err := DB.Prepare("insert into save set name=?,hash_name=?,create_time=?")
	if err != nil {
		logg.Panicln(err)
		return
	}
	_, err = stmt.Exec(name, hashName, time.Now().Unix())
	if err != nil {
		logg.Panicln(err)
	}
	logg.Println("save file :["+name+"]\t"+"=>["+hashName+"]")
}
