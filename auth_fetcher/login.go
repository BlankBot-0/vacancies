package auth_fetcher

import (
	"context"
	api2captcha "github.com/2captcha/2captcha-go"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"vacancies/auth_fetcher/config"
	"vacancies/auth_fetcher/internal"
	"vacancies/protobuf"
)

type Auth struct {
	CaptchaClient *api2captcha.Client
	Captcha       api2captcha.ReCaptcha
	Credentials   config.Credentials
}

func NewAuth(cfg *config.Config) *Auth {
	client := api2captcha.NewClient(cfg.CaptchaApiKey)
	captcha := api2captcha.ReCaptcha{
		SiteKey: internal.CaptchaSiteKey,
		Url:     internal.AccountLoginURL,
	}
	return &Auth{
		CaptchaClient: client,
		Captcha:       captcha,
		Credentials:   cfg.Credentials,
	}
}

func (a *Auth) Login() (map[string]*protobuf.Cookie, error) {
	captchaToken, err := a.solve()
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	authCookies := newAuthCookies()

	tasks := chromedp.Tasks{
		a.accountLoginTasks(captchaToken),
		a.careerLoginTasks(),
		getAuthCookiesTask(authCookies),
	}

	err = chromedp.Run(ctx, tasks)
	if err != nil {
		return nil, err
	}

	return authCookies, nil
}

func (a *Auth) accountLoginTasks(captchaToken string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(internal.AccountLoginURL),
		chromedp.WaitVisible(`#email_field`, chromedp.ByID),
		chromedp.WaitVisible(`#password_field`, chromedp.ByID),
		chromedp.WaitVisible(`button[type="submit"][name="go"]`, chromedp.ByQuery),
		chromedp.SendKeys(`#email_field`, a.Credentials.Email, chromedp.ByID),
		chromedp.SendKeys(`#password_field`, a.Credentials.Password, chromedp.ByID),
		chromedp.SetValue(`#g-recaptcha-response`, captchaToken, chromedp.ByID),
		chromedp.Click(`button[type="submit"][name="go"]`, chromedp.ByQuery),
	}
}

func (a *Auth) careerLoginTasks() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(internal.CareerLoginURL),
		chromedp.WaitVisible(`button[data-header-dropdown-toggle="user-auth-menu-mobile"]`, chromedp.ByQuery),
		chromedp.Click(`button[data-header-dropdown-toggle="user-auth-menu-mobile"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`a.button-comp--appearance-primary[data-sign-in="header"]`, chromedp.ByQuery),
		chromedp.Click(`a.button-comp--appearance-primary[data-sign-in="header"]`, chromedp.ByQuery),
	}
}

func (a *Auth) solve() (string, error) {
	code, err := a.CaptchaClient.Solve(a.Captcha.ToRequest())
	if err != nil {
		return "", err
	}
	return code, nil
}

func getAuthCookiesTask(authCookies map[string]*protobuf.Cookie) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		cookies, err := network.GetCookies().Do(ctx)
		if err != nil {
			return err
		}
		for _, cookie := range cookies {
			if _, ok := authCookies[cookie.Name]; ok {
				authCookies[cookie.Name] = toProtobufCookie(cookie)
			}
		}
		return allKeysUpdated(authCookies)
	}
}

func allKeysUpdated(target map[string]*protobuf.Cookie) error {
	for _, value := range target {
		if value == nil {
			return internal.ErrNoAuthCookies
		}
	}
	return nil
}

func newAuthCookies() map[string]*protobuf.Cookie {
	return map[string]*protobuf.Cookie{
		"mid":                 nil,
		"_career_session":     nil,
		"check_cookies":       nil,
		"remember_user_token": nil,
	}
}

func toProtobufCookie(cookie *network.Cookie) *protobuf.Cookie {
	return &protobuf.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Path:     cookie.Path,
		Domain:   cookie.Domain,
		Expires:  int64(cookie.Expires),
		Secure:   cookie.Secure,
		HttpOnly: cookie.HTTPOnly,
	}
}
