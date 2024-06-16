package utils

import (
	"fmt"
	"github.com/2captcha/2captcha-go"
	"log"
	"testing"
	"vacancies/auth_fetcher/internal/config"
)

func TestCaptchaID(t *testing.T) {
	cfg := config.Config{
		Credentials: config.Credentials{},
		Captcha: config.Captcha{
			CaptchaInURL:     "http://rucaptcha.com/in.php",
			CaptchaResultURL: "http://rucaptcha.com/res.php",
			ApiKey:           "6569786017c676f413db22a6d0ef4b69",
			SiteKey:          "6LfD3PIbAAAAAJs_eEHvoOl75_83eXSqpPSRFJ_u",
			PageUrl:          "https://rucaptcha.com/demo/recaptcha-v2",
			JsonFlag:         "1",
		},
	}

	resp, err := CaptchaID(cfg.Captcha)
	if err != nil {
		t.Errorf("CaptchaID() error = %v", err)
	}
	fmt.Println(resp.Status)
	fmt.Println(resp.Request)
}

func TestCaptchaClient(t *testing.T) {
	client := api2captcha.NewClient("6569786017c676f413db22a6d0ef4b69")
	capt := api2captcha.ReCaptcha{
		SiteKey: "6LfD3PIbAAAAAJs_eEHvoOl75_83eXSqpPSRFJ_u",
		Url:     "https://rucaptcha.com/demo/recaptcha-v2",
	}
	code, err := client.Solve(capt.ToRequest())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("code " + code)
}
