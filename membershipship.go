package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

type Page struct {
	Title string
	Body  []byte
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    t, err := template.ParseFiles(tmpl + ".html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err = t.Execute(w, p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func viewHomeHandler(w http.ResponseWriter, r *http.Request) {
	p, err := loadPage("home")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	renderTemplate(w, "home", p)
}

func main() {
	http.HandleFunc("/", viewHomeHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
