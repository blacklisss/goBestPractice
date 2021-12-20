package page

// File is not `goimports`-ed with -local github.com/golangci/golangci-lint (goimports)
import (
	"context"
	"io"
	urlParser "net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

type page struct {
	doc      *goquery.Document
	StartURL string // ST1003: struct field StartUrl should be StartURL (stylecheck)
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

	startURLInfo, _ := urlParser.Parse(p.StartURL) // ST1003: var startUrlInfo should be startURLInfo (stylecheck)

	p.doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		url, ok := s.Attr("href")
		if ok {
			if strings.HasPrefix(url, "#") {
				return
			}

			urlInfo, err := urlParser.Parse(url)
			if err != nil {
				log.Warnf("error parse URL: %s\n", url)
				return
			}

			if urlInfo.Host == "" {
				if strings.HasPrefix(url, "/") {
					url = startURLInfo.Scheme + "://" + startURLInfo.Host + url
				} else {
					url = p.StartURL + url
				}
			}
			// Здесь может быть относительная ссылка, нужно абсолютную
			urls = append(urls, url)
		}
	})
	return urls
}

func NewPage(raw io.Reader, startURL string) (page, error) { // ST1003: func parameter startUrl should be startURL (stylecheck)
	doc, err := goquery.NewDocumentFromReader(raw)
	if err != nil {
		return page{}, err
	}

	return page{doc, startURL}, nil
}
