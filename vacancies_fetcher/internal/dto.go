package internal

type VacancyDTO struct {
	KeyWord string `json:"key_word"`
	Href    string `json:"href"`
	Title   string `json:"title"`
}

type VacancyQueueDTO struct {
	KeyWord string   `json:"key_word"`
	Hrefs   []string `json:"hrefs"`
	Titles  []string `json:"titles"`
}
