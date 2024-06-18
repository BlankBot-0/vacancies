package internal

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type VacanciesFetcher struct {
	VacanciesURL string
	KeyWord      string
}

func NewVacanciesFetcher(cfg *Config) *VacanciesFetcher {
	return &VacanciesFetcher{
		VacanciesURL: cfg.VacanciesURL,
		KeyWord:      cfg.KeyWord,
	}
}

func (f *VacanciesFetcher) FetchVacancies() ([]Vacancy, error) {
	initialPage, err := f.FetchPage(1)
	if err != nil {
		return nil, err
	}

	totalResults := initialPage.Vacancies.Meta.TotalResults
	vacancies := make([]Vacancy, 0, totalResults)
	vacancies = append(vacancies, initialPage.Vacancies.Vacancies...)

	for i := 2; i < initialPage.Vacancies.Meta.TotalPages+1; i++ {
		curPage, err := f.FetchPage(i)
		if err != nil {
			return nil, err
		}
		vacancies = append(vacancies, curPage.Vacancies.Vacancies...)
	}
	return vacancies, nil
}

func (f *VacanciesFetcher) FetchPage(page int) (VacanciesJSON, error) {
	req, err := f.fetchFromPageRequest(page)
	if err != nil {
		return VacanciesJSON{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	defer func() {
		if tmpErr := resp.Body.Close(); tmpErr != nil && err == nil {
			err = tmpErr
		}
	}()
	if err != nil {
		return VacanciesJSON{}, err
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return VacanciesJSON{}, err
	}

	vacancies, err := pageJSON(string(rawBody))
	if err != nil {
		return VacanciesJSON{}, err
	}
	return vacancies, nil
}

func pageJSON(responseBody string) (VacanciesJSON, error) {
	start := `<script type="application/json" data-ssr-state="true">`
	startIndex := strings.Index(responseBody, start)
	if startIndex == -1 {
		return VacanciesJSON{}, ErrNoVacancies
	}
	startIndex += len(start)

	endIndex := strings.Index(responseBody[startIndex:], "</script>")
	if endIndex == -1 {
		return VacanciesJSON{}, ErrParsingVacancies
	}

	jsonStr := responseBody[startIndex : startIndex+endIndex]
	var vacJSON VacanciesJSON
	err := json.Unmarshal([]byte(jsonStr), &vacJSON)
	if err != nil {
		return VacanciesJSON{}, err
	}
	return vacJSON, nil
}

func (f *VacanciesFetcher) fetchFromPageRequest(page int) (*http.Request, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("q", f.KeyWord)
	req, err := http.NewRequest("GET", f.VacanciesURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}
