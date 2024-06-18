package internal

type VacanciesJSON struct {
	Vacancies VacanciesList `json:"vacancies"`
}

type VacanciesList struct {
	Vacancies []Vacancy `json:"list"`
	Meta      MetaInfo  `json:"meta"`
}

type Vacancy struct {
	Reference string `json:"href"`
	Title     string `json:"title"`
}

type MetaInfo struct {
	TotalResults int `json:"totalResults"`
	CurrentPage  int `json:"currentPage"`
	TotalPages   int `json:"totalPages"`
}
