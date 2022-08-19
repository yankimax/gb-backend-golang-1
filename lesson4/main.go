package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type helloHandler struct {
	subject string
}
type UploadHandler struct {
	HostAddr  string
	UploadDir string
}

type Employee struct {
	Name   string  `json:"name" xml:"name"`
	Age    int     `json:"age" xml:"age"`
	Salary float32 `json:"salary" xml:"salary"`
}

func (h *helloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", h.subject)
}

type Handler struct {
}

func main() {
	uploadHandler := &UploadHandler{
		UploadDir: "upload",
	}
	http.Handle("/upload", uploadHandler)
	handler := &Handler{}
	http.Handle("/", handler)
	srv := &http.Server{
		Addr:         ":8081",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	srv.ListenAndServe()
	dirToServe := http.Dir(uploadHandler.UploadDir)
	fs := &http.Server{
		Addr:         ":8082",
		Handler:      http.FileServer(dirToServe),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fs.ListenAndServe()
}
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		name := r.FormValue("name")
		fmt.Fprintf(w, "Parsed query-param with key \"name\": %s", name)
	case http.MethodPost:
		var employee Employee
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			err := json.NewDecoder(r.Body).Decode(&employee)
			if err != nil {
				http.Error(w, "Unable to unmarshal JSON", http.StatusBadRequest)
				return
			}
		case "application/xml":
			err := xml.NewDecoder(r.Body).Decode(&employee)
			if err != nil {
				http.Error(w, "Unable to unmarshal XML", http.StatusBadRequest)
				return
			}
		default:
			http.Error(w, "Unknown content type", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "Got a new employee!\nName: %s\nAge: %dy.o.\nSalary %0.2f\n",
			employee.Name,
			employee.Age,
			employee.Salary,
		)
	}
}
func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Если отправить GET запрос, то покажем список файлов и не будем обрабатывать загрузку файла
	if r.Method == http.MethodGet {
		files, err := ioutil.ReadDir(h.UploadDir)
		if err != nil {
			log.Fatal(err)
		}
		var counter = 0
		for _, file := range files {
			var ext = filepath.Ext(h.UploadDir + "/" + file.Name())
			// если нет требования к расширению или оно совпадает с расширением текущего файла, то выводим.
			if r.FormValue("ext") == "" || ext == "."+r.FormValue("ext") {
				fmt.Fprintln(w, file.Name(), file.Size(), ext)
				counter++
			}
		}
		fmt.Fprintln(w, fmt.Sprintf("Count: %v", counter))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file"+r.Method, http.StatusBadRequest)
		return
	}
	filePath := fmt.Sprintf("%v_%s", time.Now().Unix(), header.Filename) // добавил таймштамп
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s", h.UploadDir, filePath), data, 0777)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	fileLink := h.HostAddr + "/" + filePath
	fmt.Fprintln(w, fileLink)
}
