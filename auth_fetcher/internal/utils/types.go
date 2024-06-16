package utils

type CaptchaResponse struct {
	Status  int    `json:"status"`
	Request string `json:"request"`
}
