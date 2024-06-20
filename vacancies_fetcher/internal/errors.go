package internal

import "errors"

var ErrNoVacancies = errors.New("no vacancies found")
var ErrParsingVacancies = errors.New("error parsing vacancies")
