package utils

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
	"vacancies/auth_fetcher/internal/config"
)

type CaptchaSolver struct {
	SendURL        string
	GetSolutionURL string
}

func NewCaptchaSolver(cfg config.Captcha) *CaptchaSolver {
	sendParams := url.Values{}
	sendParams.Add("key", cfg.ApiKey)
	sendParams.Add("method", "userrecaptcha")
	sendParams.Add("googlekey", cfg.SiteKey)
	sendParams.Add("pageurl", cfg.PageUrl)
	if cfg.JsonFlag == "1" {
		sendParams.Add("json", "1")
	}

	getParams := url.Values{}
	getParams.Add("key", cfg.ApiKey)
	getParams.Add("action", "get")
	if cfg.JsonFlag == "1" {
		getParams.Add("json", "1")
	}

	return &CaptchaSolver{
		SendURL:        cfg.CaptchaInURL + "?" + sendParams.Encode(),
		GetSolutionURL: cfg.CaptchaResultURL + "?" + getParams.Encode(),
	}
}

func (s *CaptchaSolver) captchaID() (CaptchaResponse, error) {
	resp, err := http.Get(s.SendURL)
	if err != nil {
		return CaptchaResponse{}, err
	}
	defer func() {
		err = resp.Body.Close()
	}()

	var respData CaptchaResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return CaptchaResponse{}, err
	}
	return respData, nil
}

func (s *CaptchaSolver) captchaSolution(id string) (string, error) {
	resp, err := http.Get(s.GetSolutionURL + "?id=" + id)
	if err != nil {
		return "", err
	}
	defer func() {
		err = resp.Body.Close()
	}()
	var respData CaptchaResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return "", err
	}
	return respData.Request, nil
}

//func (s *CaptchaSolver) Login(id, answer string) error {
//
//}

func CaptchaID(cfg config.Captcha) (CaptchaResponse, error) {
	queryParams := url.Values{}
	queryParams.Add("key", cfg.ApiKey)
	queryParams.Add("method", "userrecaptcha")
	queryParams.Add("googlekey", cfg.SiteKey)
	queryParams.Add("pageurl", cfg.PageUrl)
	if cfg.JsonFlag == "1" {
		queryParams.Add("json", "1")
	}
	URL := cfg.CaptchaInURL + "?" + queryParams.Encode()

	resp, err := http.Get(URL)
	if err != nil {
		return CaptchaResponse{}, err
	}
	defer func() {
		err = resp.Body.Close()
	}()

	var respData CaptchaResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return CaptchaResponse{}, err
	}
	return respData, nil
}

//func CaptchaKey(id string) (string, error) {
//	queryParams := url.Values{}
//
//}

const CaptchaWaitTime = 20 * time.Second

// MsgCaptchaNotReady has a wierd spelling, but it is as in the documentation
const MsgCaptchaNotReady = "CAPCHA_NOT_READY"
