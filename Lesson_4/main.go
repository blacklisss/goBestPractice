package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"io"
	"l1g2/configuration"
	"net/http"
	urlParser "net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

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
	ctx2, _ := context.WithTimeout(ctx, time.Second*2)
	ctx3, _ := context.WithTimeout(ctx, time.Second*5)

	c.visitedMu.RLock()
	if _, ok := c.visited[url]; ok {
		c.visitedMu.RUnlock()
		return
	}
	c.visitedMu.RUnlock()
	if curDepth >= c.maxDepth {
		//log.Infof("Max Depth on URL: %s\n", url)
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
	timeout  time.Duration
	client   HTTPClient
	startUrl string
}

func NewRequester(timeout time.Duration, startUrl string) *requester {
	cl := &http.Client{
		Timeout: timeout,
	}
	return &requester{timeout: timeout, client: cl, startUrl: startUrl}
}

func (r requester) GetPage(ctx context.Context, url string) (Page, error) {

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
		return NewPage(rawPage.Body, r.startUrl)
	}

}

type Page interface {
	GetTitle(ctx context.Context) string
	GetLinks() []string
}

type page struct {
	doc      *goquery.Document
	startUrl string
}

func NewPage(raw io.Reader, startUrl string) (page, error) {
	doc, err := goquery.NewDocumentFromReader(raw)
	if err != nil {
		return page{}, err
	}
	return page{doc, startUrl}, nil
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
	startUrlInfo, _ := urlParser.Parse(p.startUrl)
	p.doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		url, ok := s.Attr("href")
		if ok {
			if strings.HasPrefix(url, "#") {
				//log.Infof("Anchor link %s\n", url)
				return
			}

			urlInfo, err := urlParser.Parse(url)
			if err != nil {
				log.Warnf("error parse URL: %s\n", url)
				return
			}

			if urlInfo.Host == "" {
				if strings.HasPrefix(url, "/") {
					url = startUrlInfo.Scheme + "://" + startUrlInfo.Host + url
				} else {
					url = p.startUrl + url
				}
			}
			//Здесь может быть относительная ссылка, нужно абсолютную
			urls = append(urls, url)
		}
	})
	return urls
}

func processResult(ctx context.Context, in <-chan CrawlResult, cancel context.CancelFunc) {
	var errCount int
	var ticker = time.NewTicker(time.Second * 1)
	var sleeping time.Time

	for {
		select {
		case res := <-in:
			sleeping = time.Now()
			if res.Err != nil {
				log.Errorf("ERROR Link: %s, err: %v\n", res.Url, res.Err)
				errCount++
				if errCount >= 1 {
					//cancel()
				}
			} else {
				log.Infof("Link: %s, Title: %s\n", res.Url, res.Title)
			}
		case <-ctx.Done():
			log.Warnf("context canceled\n")
			return
		case <-ticker.C:
			now := time.Now()
			if now.Sub(sleeping).Seconds() > 7 {
				log.Warnf("process timeout\n")
				cancel()
				return
			}
		}
	}
}

func main() {
	pid := os.Getpid()
	log.Infof("My PID is: %d\n", pid)
	config, err := configuration.Load("configuration/config.yaml")
	if err != nil {
		log.Fatalf("cannot load config")
	}

	var r Requester
	r = NewRequester(time.Minute, config.StartUrl)
	//r = NewRequestWithDelay(2*time.Second, r)
	ctx, cancel := context.WithCancel(context.Background())
	crawler := NewCrawler(2, r)
	crawler.Scan(ctx, config.StartUrl, 0)
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
				log.Infof("Signal SIGTERM catched\n")
				cancel()
			}
		case <-ctx.Done():
			log.Warnf("context canceled\n")
			return
		}
	}
	//cancel()
}
