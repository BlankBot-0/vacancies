package vacancies_fetcher

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"vacancies/protobuf"
	"vacancies/vacancies_fetcher/internal"
)

type VacanciesFetcher struct {
	KeyWord     string
	AuthCookies map[string]*http.Cookie
}

func New(keyword string, data *protobuf.AuthorizationData) *VacanciesFetcher {
	cookies := map[string]*http.Cookie{
		"mid":                 toHTTPCookie(data.Mid),
		"_career_session":     toHTTPCookie(data.XCareerSession),
		"remember_user_token": toHTTPCookie(data.RememberUserToken),
		"check_cookies":       toHTTPCookie(data.CheckCookies),
	}

	return &VacanciesFetcher{
		KeyWord:     keyword,
		AuthCookies: cookies,
	}
}

func (f *VacanciesFetcher) FetchVacancies() ([]internal.Vacancy, error) {
	initialPage, err := f.FetchPage(1)
	if err != nil {
		return nil, err
	}

	totalResults := initialPage.Vacancies.Meta.TotalResults
	vacancies := make([]internal.Vacancy, 0, totalResults)
	vacancies = append(vacancies, initialPage.Vacancies.Vacancies...)

	for i := 2; i < initialPage.Vacancies.Meta.TotalPages+1; i++ {
		curPage, err := f.FetchPage(i)
		if err != nil {
			return nil, err
		}
		vacancies = append(vacancies, curPage.Vacancies.Vacancies...)
	}
	for i, vacancy := range vacancies {
		vacancies[i].Reference = internal.BaseURL + vacancy.Reference
	}

	return vacancies, nil
}

func (f *VacanciesFetcher) FetchPage(page int) (internal.VacanciesJSON, error) {
	req, err := f.fetchFromPageRequest(page)
	if err != nil {
		return internal.VacanciesJSON{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	defer func() {
		if tmpErr := resp.Body.Close(); tmpErr != nil && err == nil {
			err = tmpErr
		}
	}()
	if err != nil {
		return internal.VacanciesJSON{}, err
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return internal.VacanciesJSON{}, err
	}

	vacancies, err := pageJSON(string(rawBody))
	if err != nil {
		return internal.VacanciesJSON{}, err
	}

	f.updateCookies(resp)

	return vacancies, nil
}

func pageJSON(responseBody string) (internal.VacanciesJSON, error) {
	start := `<script type="application/json" data-ssr-state="true">`
	startIndex := strings.Index(responseBody, start)
	if startIndex == -1 {
		return internal.VacanciesJSON{}, internal.ErrNoVacancies
	}
	startIndex += len(start)

	endIndex := strings.Index(responseBody[startIndex:], "</script>")
	if endIndex == -1 {
		return internal.VacanciesJSON{}, internal.ErrParsingVacancies
	}

	jsonStr := responseBody[startIndex : startIndex+endIndex]
	var vacJSON internal.VacanciesJSON
	err := json.Unmarshal([]byte(jsonStr), &vacJSON)
	if err != nil {
		return internal.VacanciesJSON{}, err
	}
	return vacJSON, nil
}

func (f *VacanciesFetcher) fetchFromPageRequest(page int) (*http.Request, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("q", f.KeyWord)
	req, err := http.NewRequest("GET", internal.VacanciesURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	for _, cookie := range f.AuthCookies {
		req.AddCookie(cookie)
	}

	return req, nil
}

func (f *VacanciesFetcher) updateCookies(r *http.Response) {
	for _, cookie := range r.Cookies() {
		if _, ok := f.AuthCookies[cookie.Name]; ok {
			f.AuthCookies[cookie.Name] = cookie
		}
	}
}

func toHTTPCookie(cookie *protobuf.Cookie) *http.Cookie {
	return &http.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Path:     cookie.Path,
		Domain:   cookie.Domain,
		Expires:  time.Unix(cookie.Expires, 0),
		Secure:   cookie.Secure,
		HttpOnly: cookie.HttpOnly,
	}
}
