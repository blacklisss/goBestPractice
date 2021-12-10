package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	urlParser "net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Crawler interface {
	Scan(ctx context.Context, url string, curDepth int)
	GetResultChan() <-chan CrawlResult
	Wait()
	Start()
}

type CrawlResult struct {
	Title string
	Url   string
	Err   error
}

type crawler struct {
	maxDepth  int
	req       Requester
	res       chan CrawlResult
	visited   map[string]struct{}
	visitedMu sync.RWMutex
}

func (c *crawler) GetResultChan() <-chan CrawlResult {
	return c.res
}

func NewCrawler(maxDepth int, req Requester) *crawler {
	return &crawler{
		maxDepth: maxDepth,
		req:      req,
		res:      make(chan CrawlResult, 100),
		visited:  make(map[string]struct{}),
	}
}

func (c *crawler) Scan(ctx context.Context, url string, curDepth int) {
	ctx2, _ := context.WithTimeout(ctx, time.Second*1)
	ctx3, _ := context.WithTimeout(ctx, time.Second*3)

	c.visitedMu.RLock()
	if _, ok := c.visited[url]; ok {
		c.visitedMu.RUnlock()
		return
	}
	c.visitedMu.RUnlock()
	if curDepth >= c.maxDepth {
		return
	}
	select {
	case <-ctx.Done():
		return
	default:
		page, err := c.req.GetPage(ctx3, url)
		c.visitedMu.Lock()
		c.visited[url] = struct{}{}
		c.visitedMu.Unlock()
		if err != nil {
			c.res <- CrawlResult{Url: url, Err: err}
			return
		}
		title := page.GetTitle(ctx2)

		if title == "" {
			c.res <- CrawlResult{Url: url, Err: errors.New("page title timeout")}
			return
		}

		c.res <- CrawlResult{
			Title: title,
			Url:   url,
			Err:   nil,
		}
		links := page.GetLinks()
		for _, link := range links {
			go c.Scan(ctx, link, curDepth+1)
		}
	}
}

type Requester interface {
	GetPage(ctx context.Context, url string) (Page, error)
}

type reqWithDelay struct {
	delay time.Duration
	req   Requester
}

func NewRequestWithDelay(delay time.Duration, req Requester) *reqWithDelay {
	return &reqWithDelay{delay: delay, req: req}
}

func (r reqWithDelay) GetPage(ctx context.Context, url string) (Page, error) {
	time.Sleep(r.delay)
	return r.req.GetPage(ctx, url)
}

/*
type HttpClient interface {
	 Do(r *http.Request) (*http.Response, error)
}
*/

type requester struct {
	timeout time.Duration
}

func NewRequester(timeout time.Duration) *requester {
	return &requester{timeout: timeout}
}

func (r requester) GetPage(ctx context.Context, url string) (Page, error) {

	select {
	case <-ctx.Done():
		return nil, errors.New("waiting too long response from " + url)
	default:
		cl := &http.Client{
			Timeout: r.timeout,
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		rawPage, err := cl.Do(req)
		if err != nil {
			return nil, err
		}
		defer rawPage.Body.Close()
		return NewPage(rawPage.Body)
	}

}

type Page interface {
	GetTitle(ctx context.Context) string
	GetLinks() []string
}

type page struct {
	doc *goquery.Document
}

func NewPage(raw io.Reader) (page, error) {
	doc, err := goquery.NewDocumentFromReader(raw)
	if err != nil {
		return page{}, err
	}
	return page{doc}, nil
}

func (p page) GetTitle(ctx context.Context) string {
	select {
	case <-ctx.Done():
		return ""
	default:
		return p.doc.Find("title").First().Text()
	}
}

func (p page) GetLinks() []string {
	var urls []string
	startUrlInfo, _ := urlParser.Parse(startUrl)
	p.doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		url, ok := s.Attr("href")
		if ok {
			if strings.HasPrefix(url, "#") {
				return
			}

			urlInfo, err := urlParser.Parse(url)
			if err != nil {
				return
			}

			if urlInfo.Host == "" {
				if !strings.HasPrefix(url, "/") {
					url = "/" + url
				}

				url = startUrlInfo.Scheme + "://" + startUrlInfo.Host + url
			}
			//Здесь может быть относительная ссылка, нужно абсолютную
			urls = append(urls, url)
		}
	})
	return urls
}

const startUrl = "https://www.w3.org/Consortium/"

func processResult(ctx context.Context, in <-chan CrawlResult, cancel context.CancelFunc) {
	var errCount int
	var ticker = time.NewTicker(time.Second * 1)
	var sleeping time.Time

	for {
		select {
		case res := <-in:
			sleeping = time.Now()
			if res.Err != nil {
				errCount++
				fmt.Printf("ERROR Link: %s, err: %v\n", res.Url, res.Err)
				if errCount >= 1 {
					//cancel()
				}
			} else {
				fmt.Printf("Link: %s, Title: %s\n", res.Url, res.Title)
			}
		case <-ctx.Done():
			fmt.Printf("context canceled\n")
			return
		case <-ticker.C:
			now := time.Now()
			if now.Sub(sleeping).Seconds() > 7 {
				fmt.Printf("process timeout\n")
				cancel()
				return
			}
		}
	}
}

func main() {
	pid := os.Getpid()
	fmt.Printf("My PID is: %d\n", pid)
	var r Requester
	r = NewRequester(time.Minute)
	//r = NewRequestWithDelay(2*time.Second, r)
	ctx, cancel := context.WithCancel(context.Background())
	crawler := NewCrawler(2, r)
	crawler.Scan(ctx, startUrl, 0)
	chSig := make(chan os.Signal)
	signal.Notify(chSig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGUSR1)
	go processResult(ctx, crawler.GetResultChan(), cancel)
	for {
		select {
		case v := <-chSig:
			if v == syscall.SIGUSR1 {
				crawler.maxDepth += 2
				fmt.Printf("MaxDepth chanched to %d\n", crawler.maxDepth)
			} else {
				fmt.Printf("Signal SIGTERM catched\n")
				cancel()
			}
		case <-ctx.Done():
			fmt.Printf("context canceled\n")
			return
		}
	}
	//cancel()
}
