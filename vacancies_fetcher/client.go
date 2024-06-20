package vacancies_fetcher

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"sync"
	"time"
	"vacancies/protobuf"
	"vacancies/vacancies_fetcher/config"
	"vacancies/vacancies_fetcher/internal"
	"vacancies/vacancies_fetcher/repository"
)

type (
	Fetcher interface {
		FetchVacancies() ([]internal.Vacancy, error)
	}
	MessageBroker interface {
		Send(vacancies []internal.Vacancy, keyWord string) error
		StartReceiving(wg *sync.WaitGroup) error
	}
	Repo interface {
		AddVacancies(ctx context.Context, arg repository.AddVacanciesParams) error
	}
)

type Deps struct {
	Fetcher       Fetcher
	MessageBroker MessageBroker
	Repo          Repo
}

type Client struct {
	Deps
}

func NewClient(deps Deps) *Client {
	return &Client{
		Deps: deps,
	}
}

func (c *Client) Start(cfg *config.Config, wg *sync.WaitGroup) error {
	vacancies, err := c.Fetcher.FetchVacancies()
	if err != nil {
		return err
	}
	err = c.MessageBroker.Send(vacancies, cfg.KeyWord)
	if err != nil {
		return err
	}
	err = c.MessageBroker.StartReceiving(wg)
	return err
}

func CheckAuth(authCookies []*protobuf.Cookie) (bool, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var existsAvatar bool
	tasks := chromedp.Tasks{
		setCookies(authCookies),
		chromedp.Navigate(internal.VacanciesURL),
		chromedp.WaitVisible(internal.AvatarSelector, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.
				Evaluate(fmt.Sprintf(`document.querySelector('%s') !== null`, internal.AvatarSelector), &existsAvatar).
				Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
	}

	err := chromedp.Run(ctx, tasks)
	if err != nil {
		return false, err
	}

	return existsAvatar, nil
}

func setCookies(cookies []*protobuf.Cookie) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		for _, cookie := range cookies {
			expr := cdp.TimeSinceEpoch(time.Unix(cookie.Expires, 0))
			network.SetCookie(cookie.Name, cookie.Value).
				WithDomain(cookie.Domain).
				WithPath(cookie.Path).
				WithSecure(cookie.Secure).
				WithHTTPOnly(cookie.HttpOnly).
				WithExpires(&expr)
		}
		return nil
	}
}
