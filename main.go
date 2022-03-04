package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"flag"
	"fmt"
	"hash"
	"hash/crc32"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

const (
	UPLOAD_DIR = "./uploads"
)

var uploadTemplate = template.Must(template.ParseFiles("encrypt.html"))

func indexHandle(w http.ResponseWriter, r *http.Request) {
	if err := uploadTemplate.Execute(w, nil); err != nil {
		log.Fatal("Execute: ", err.Error())
		return
	}
}

func uploadHandle(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	f, err := os.OpenFile("./uploads/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666) // 此处假设当前目录下已存在test目录
	if err != nil {
		return
	}
	defer f.Close()
	io.Copy(f, file)

	mod := flag.String("mod", "md5", "md5,sha1,sha256,crc32")
	h, ok := map[string]hash.Hash{
		"md5":    md5.New(),
		"sha1":   sha1.New(),
		"sha256": sha256.New(),
		"crc32":  crc32.NewIEEE(),
	}[*mod]
	if !ok {
		h = md5.New()
	}
	// 进行文件
	err = encFile("./uploads/"+handler.Filename, h)
	if err != nil {
		fmt.Println("err")
	}
	file, err = os.Open("./uploads/" + handler.Filename + "dst")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	fileNames := url.QueryEscape("./uploads/" + handler.Filename + "dst") // 防止中文乱码
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", "attachment; filename=\""+fileNames+"\"")

	if err != nil {
		fmt.Println("Read File Err:", err.Error())
	} else {
		w.Write(content)
	}
}

func main() {
	http.HandleFunc("/", indexHandle)
	http.HandleFunc("/uploads", uploadHandle)
	http.ListenAndServe(":8080", nil)
}
