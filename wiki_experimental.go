package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"os"
	"code.google.com/p/go.net/websocket"
)

// log type
type Log struct {
	Name  string
	Data  []byte
}
func (l *Log) create() error {
	var logFile *os.File
	var err error
	filename := "logs/" + l.Name + ".log"
	//data, err := ioutil.ReadFile(filename)
	//l.Data = data
    logFile, err = os.Create(filename)
    defer logFile.Close()
    return err
}
func (l *Log) read() error {
	filename := "logs/" + l.Name + ".log"
	filedata, err := ioutil.ReadFile(filename)
	if err != nil { return err }
	l = &Log{Name: l.Name, Data: filedata}
	return err
}
func (l *Log) save() error {
	filename := "logs/" + l.Name + ".log"
	return ioutil.WriteFile(filename, l.Data, 0600)
}
func (l *Log) append(message string) error {
	temp := string(l.Data)
	temp += message + "\n"
	l = &Log{Name: l.Name, Data: []byte(temp)}
	err := l.save()
	return err
}
func Logln(message string) error {
	fmt.Println(message)
	//filename := "logs/" + "main" + ".log"
	//data, err := ioutil.ReadFile(filename)
	//if err != nil { return err }
	//l := &Log{Name: "main", Data: data}
	//l := &Log{Name: "main", Data: []byte(message)}
	l := &Log{Name: "main", Data: []byte("Starting Log...\n")}
	//err := l.read()
	err := l.append(message)
	//err := mainLog.append(message)
	return err
}

// file type
type File struct {
	Loc string
	Data  []byte
}
func loadFile(loc string) (*File, error) {
	data, err := ioutil.ReadFile(loc)
	if err != nil { return nil, err }
	return &File{Loc: loc, Data: data}, nil
}

// page type
type Page struct {
	Title string
	Body  []byte
}
func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}
func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil { return nil, err }
	return &Page{Title: title, Body: body}, nil
}

// handlers
func socketHandler(ws *websocket.Conn) {
    // meow
}
func rootHandler(w http.ResponseWriter, r *http.Request) {
	url := "home"
	if r.URL.Path[1:] != "" { url = r.URL.Path[1:] }
	Logln(r.Method + ": " + url)
	if strings.HasPrefix(url, "file/") {
		f, err := loadFile(url)
		if err != nil { return }
		fmt.Fprintf(w, "%s", f.Data)
	}
	http.Redirect(w, r, "/view/"+url, http.StatusFound)
}
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	Logln(r.Method + ": " + r.URL.Path[1:])
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	Logln(r.Method + ": " + r.URL.Path[1:])
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	Logln(r.Method + ": " + r.URL.Path[1:])
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// template file locations
var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))

// template rendering
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl + ".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// path length
const lenPath = len("/view/")

// title validation
var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")

// handler creator
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path[lenPath:]
		if !titleValidator.MatchString(title) {
			http.NotFound(w, r)
			return
		}
		fn(w, r, title)
	}
}

// run
func main() {
	//l := &Log{Name: "main", Data: []byte("Starting Log")}
	//l.create()
	//mainLog.create()
	go http.HandleFunc("/", rootHandler)
	go http.Handle("/sock/", websocket.Handler(socketHandler))
	go http.HandleFunc("/view/", makeHandler(viewHandler))
	go http.HandleFunc("/edit/", makeHandler(editHandler))
	go http.HandleFunc("/save/", makeHandler(saveHandler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
        fmt.Printf("ERR: %v", err)
    }
}
