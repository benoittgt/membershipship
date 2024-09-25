package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Member struct {
	FirstName      string
	LastName       string
	Email          string
	JoinDate       time.Time
	ExpirationDate time.Time
}

type Page struct {
	Title   string
	Body    []byte
	Members []Member
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func parseDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"02/01/2006", // DD/MM/YYYY
		"2/1/2006",   // D/M/YYYY
		"1/2/2006",   // M/D/YYYY
		"02/1/2006",  // DD/M/YYYY
		"2/01/2006",  // D/MM/YYYY
		"2006",       // YYYY
	}

	dateStr = strings.TrimSpace(dateStr)

	for _, layout := range layouts {
		if parsedTime, err := time.Parse(layout, dateStr); err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func readCSVFromUrl(url string) ([]Member, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.Comma = ','
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var members []Member
	for i, row := range data {
		if i == 0 || len(row) < 6 { // Skip header row and ensure row has enough columns
			continue
		}
		joinDate, err := parseDate(row[5]) // Assuming join date is in column 6
		if err != nil {
			log.Printf("Error parsing join date for row %d: %v. Using current date instead.", i, err)
			joinDate = time.Now() // Use current date as a fallback
		}
		member := Member{
			FirstName:      strings.TrimSpace(row[1]),
			LastName:       strings.TrimSpace(row[2]),
			Email:          strings.TrimSpace(row[3]),
			JoinDate:       joinDate,
			ExpirationDate: joinDate.AddDate(1, 0, 0), // Add 1 year to join date
		}
		members = append(members, member)
	}
	return members, nil
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

func fetchMemberData() ([]Member, error) {
	url := os.Getenv("CSV_URL")
	if url == "" {
		return nil, fmt.Errorf("CSV_URL environment variable is not set")
	}
	return readCSVFromUrl(url)
}

func viewHomeHandler(w http.ResponseWriter, r *http.Request) {
	p, err := loadPage("home")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	members, err := fetchMemberData()
	if err != nil {
		http.Error(w, "Error fetching member data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	p.Members = members

	renderTemplate(w, "home", p)
}

func main() {
	http.HandleFunc("/", viewHomeHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
