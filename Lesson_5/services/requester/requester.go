package requester

import (
	"context"
	"errors"
	"l1g2/services/crawler"
	"l1g2/services/page"
	"net/http"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type requester struct {
	timeout  time.Duration
	client   HTTPClient
	startUrl string
}

func (r requester) GetPage(ctx context.Context, url string) (crawler.Page, error) {

	select {
	case <-ctx.Done():
		return nil, errors.New("waiting too long response from " + url)
	default:
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		rawPage, err := r.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer rawPage.Body.Close()
		return page.NewPage(rawPage.Body, r.startUrl)
	}

}

func NewRequester(timeout time.Duration, startUrl string) *requester {
	cl := &http.Client{
		Timeout: timeout,
	}
	return &requester{timeout: timeout, client: cl, startUrl: startUrl}
}
