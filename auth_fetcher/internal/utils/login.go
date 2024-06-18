package utils

import (
	"bytes"
	"encoding/json"
	api2captcha "github.com/2captcha/2captcha-go"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
	"vacancies/auth_fetcher/internal/config"
)

type Auth struct {
	LoginURL      string
	CaptchaClient *api2captcha.Client
	Captcha       api2captcha.ReCaptcha
	Credentials   config.Credentials
}

func NewAuth(cfg config.Config) *Auth {
	client := api2captcha.NewClient(cfg.Captcha.ApiKey)
	captcha := api2captcha.ReCaptcha{
		SiteKey: cfg.Captcha.SiteKey,
		Url:     cfg.Captcha.PageUrl,
	}
	return &Auth{
		LoginURL:      cfg.LoginURL,
		CaptchaClient: client,
		Captcha:       captcha,
		Credentials:   cfg.Credentials,
	}
}

func (a *Auth) Solve() (string, error) {
	code, err := a.CaptchaClient.Solve(a.Captcha.ToRequest())
	if err != nil {
		return "", err
	}
	return code, nil
}

func (a *Auth) Login() (*http.Cookie, error) {
	req, err := a.formLoginRequest()
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tmpErr := resp.Body.Close(); tmpErr != nil && err == nil {
			err = tmpErr
		}
	}()

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "acc_sess_id" {
			return cookie, nil
		}
	}
	return nil, ErrNoSessionIDCookie
}

func (a *Auth) formLoginRequest() (*http.Request, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	err := w.WriteField("email", a.Credentials.Email)
	if err != nil {
		return nil, err
	}

	err = w.WriteField("password", a.Credentials.Password)
	if err != nil {
		return nil, err
	}

	err = w.WriteField("state", "")
	if err != nil {
		return nil, err
	}

	err = w.WriteField("consumer", "default")
	if err != nil {
		return nil, err
	}

	err = w.WriteField("captcha", "")
	if err != nil {
		return nil, err
	}

	captchaCode, err := a.Solve()
	if err != nil {
		return nil, err
	}

	err = w.WriteField("g-recaptcha-response", captchaCode)
	if err != nil {
		return nil, err
	}

	err = w.WriteField("captcha_type", "recaptcha")
	if err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("POST", a.LoginURL, &body)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", w.FormDataContentType())
	return r, nil
}

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
