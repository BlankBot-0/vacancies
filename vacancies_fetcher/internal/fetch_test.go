package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
)

func TestUnauthorizedVacanciesRetrieval(t *testing.T) {
	resp, err := http.Get("https://career.habr.com/vacancies?q=go&type=all")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if tmpErr := resp.Body.Close(); tmpErr != nil && err == nil {
			err = tmpErr
		}
	}()
	bodyRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	body := string(bodyRaw)
	start := `<script type="application/json" data-ssr-state="true">`
	startIndex := strings.Index(body, start)
	if startIndex == -1 {
		log.Fatal("start not found")
	}
	startIndex += len(start)

	// Find the ending point of the JSON data
	endIndex := strings.Index(body[startIndex:], "</script>")
	if endIndex == -1 {
		log.Fatal("end not found")
	}

	// Extract and return the JSON data
	jsonStr := body[startIndex : startIndex+endIndex]
	var vacJSON VacanciesJSON
	err = json.Unmarshal([]byte(jsonStr), &vacJSON)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Vacancies: %+v\n", vacJSON)
}

func TestFetchVacancies(t *testing.T) {
	fetcher := VacanciesFetcher{
		VacanciesURL: "https://career.habr.com/vacancies",
		KeyWord:      "go",
	}
	vacancies, err := fetcher.FetchVacancies()
	if err != nil {
		log.Fatal(err)
	}
	for _, vacancy := range vacancies {
		fmt.Println(vacancy)
	}
}
