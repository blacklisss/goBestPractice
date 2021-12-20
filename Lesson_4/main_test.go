package main

import (
	"bytes"
	"context"
	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func Test_page_GetTitle(t *testing.T) {
	assert := assert.New(t)
	ctx, _ := context.WithCancel(context.Background())

	type fields struct {
		doc *goquery.Document
	}

	type TestCase struct {
		name     string
		fields   fields
		want     string
		testFunc func(want string, p page)
	}

	doc, _ := goquery.NewDocumentFromReader(bytes.NewBuffer([]byte("<title>Test</title>")))

	tests := []TestCase{
		TestCase{
			"Equal Title",
			fields{doc},
			"Test",
			func(want string, p page) {
				assert.Equalf(want, p.GetTitle(ctx), "GetTitle(%v)", p.GetTitle(ctx))
			},
		},

		TestCase{
			"NotEqual Title",
			fields{doc},
			"Tets",
			func(want string, p page) {
				assert.NotEqualf(want, p.GetTitle(ctx), "GetTitle(%v)", p.GetTitle(ctx))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := page{
				doc:      tt.fields.doc,
				startUrl: "",
			}
			tt.testFunc(tt.want, p)
		})
	}
}

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do is the mock client's `Do` func
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

var (
	// GetDoFunc fetches the mock client's `Do` func
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

func Test_requester_GetPage(t *testing.T) {
	assert := assert.New(t)
	ctx, _ := context.WithCancel(context.Background())

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(bytes.NewReader([]byte("Test")))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	type fields struct {
		timeout time.Duration
		client  HTTPClient
	}
	type args struct {
		ctx context.Context
		url string
	}

	client := &MockClient{}
	doc, _ := goquery.NewDocumentFromReader(bytes.NewBuffer([]byte("Test")))

	type TestCase struct {
		name   string
		fields fields
		args   args
		want   page
	}

	tests := []TestCase{
		TestCase{
			"test 1",
			fields{
				time.Duration(time.Second * 30),
				client,
			},
			args{
				ctx,
				"http://google.com",
			},
			page{doc, ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := requester{
				timeout: tt.fields.timeout,
				client:  tt.fields.client,
			}
			got, err := r.GetPage(tt.args.ctx, tt.args.url)
			if err != nil {
				t.Errorf("error Get Page")
			}
			assert.Equalf(tt.want, got, "Equal")
		})
	}
}
