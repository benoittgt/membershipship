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
	Members []Member
}

const baseUrl = "https://walletobjects.googleapis.com/walletobjects/v1"

func googleApplicationCredentials() (string, error) {
	credentials := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentials == "" {
		return "", fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS environment variable is not set")
	}
	return credentials, nil
}

func googleClassId() (string, error) {
	classId := os.Getenv("GOOGLE_CLASS_ID")
	if classId == "" {
		return "", fmt.Errorf("GOOGLE_CLASS_ID environment variable is not set")
	}
	return classId, nil
}

func parseDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"02/01/2006", // DD/MM/YYYY
		"2/1/2006",   // D/M/YYYY
		"1/2/2006",   // M/D/YYYY
		"02/1/2006",  // DD/M/YYYY
		"2/01/2006",  // D/MM/YYYY
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

func renderHtmlTemplate(w http.ResponseWriter, tmpl string, p *Page) {
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
	p := &Page{}

	members, err := fetchMemberData()
	if err != nil {
		http.Error(w, "Error fetching member data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	p.Members = members

	renderHtmlTemplate(w, "home", p)
}

func renderJsonTemplate(firstName, lastName, expirationDate string) (string, error) {
	templateFile := "./google_card.json"
	templateBytes, err := os.ReadFile(templateFile)
	if err != nil {
		return "", fmt.Errorf("error reading JSON template file: %v", err)
	}
	templateStr := string(templateBytes)

	data := struct {
		FirstName      string
		LastName       string
		ExpirationDate string
	}{
		FirstName:      firstName,
		LastName:       lastName,
		ExpirationDate: expirationDate,
	}
	tmpl, err := template.New("jsonTemplate").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("error parsing JSON template: %v", err)
	}

	var renderedTemplate strings.Builder
	err = tmpl.Execute(&renderedTemplate, data)
	if err != nil {
		return "", fmt.Errorf("error rendering JSON template: %v", err)
	}

	return renderedTemplate.String(), nil
}

func generateGoogleCardHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	firstName := query.Get("firstName")
	lastName := query.Get("lastName")
	expirationDate := query.Get("ExpirationDate")

	jsonPayload, err := renderJsonTemplate(firstName, lastName, expirationDate)
	if err != nil {
		http.Error(w, "Error generating JSON payload: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, jsonPayload)
}

func generateGoogleCard(jsonPayload string) (string, error) {
	return "", nil
}

func generateAppleCardHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func main() {
	http.HandleFunc("/", viewHomeHandler)
	http.HandleFunc("/card/generate_google", generateGoogleCardHandler)
	http.HandleFunc("/card/generate_apple", generateAppleCardHandler)
	fmt.Println("Listening http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
