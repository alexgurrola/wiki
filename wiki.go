// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"os"
	"strings"
)

// log type
type Log struct {
	Name string
	Data []byte
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
func (l *Log) save() error {
	filename := "logs/" + l.Name + ".log"
	return ioutil.WriteFile(filename, l.Data, 0600)
}
func (l *Log) append(message string) error {
	//filename := "logs/" + l.Name + ".log"
	//f, err := os.OpenFile(filename, os.O_APPEND, 0666) 
	//if err != nil { return err }
	//f.WriteString(message + "\n")
	//f.Close()

	//l.Data = append(l.Data, data)
	//temp := string(l.Data)
	//temp += data
	//l.Data = []byte(temp)

	err := l.save()
	return err
}
func Logln(message string) error {
	fmt.Println(message)
	l := &Log{Name: "main", Data: []byte(message)}
	err := l.append(message)
	//err := mainLog.append(message)
	return err
}

// file type
type File struct {
	Loc  string
	Data []byte
}

func loadFile(loc string) (*File, error) {
	data, err := ioutil.ReadFile(loc)
	if err != nil {
		return nil, err
	}
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
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

// handlers
func rootHandler(w http.ResponseWriter, r *http.Request) {
	url := "home"
	if r.URL.Path[1:] != "" {
		url = r.URL.Path[1:]
	}
	Logln(r.Method + ": " + url)
	if strings.HasPrefix(url, "file/") {
		f, err := loadFile(url)
		if err != nil {
			return
		}
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
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

// run
func main() {
	l := &Log{Name: "main", Data: []byte("Starting Log")}
	l.create()
	//mainLog.create()
	go http.HandleFunc("/", rootHandler)
	go http.HandleFunc("/view/", makeHandler(viewHandler))
	go http.HandleFunc("/edit/", makeHandler(editHandler))
	go http.HandleFunc("/save/", makeHandler(saveHandler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("ERR: %v", err)
	}
}
