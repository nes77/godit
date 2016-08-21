package godit

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

const VERSION string = "0.1"
const nanosecMult int64 = 1000000000
const redditApi string = "https://www.reddit.com/api/v1"
const login string = "/authorize"

type Client struct {
	Params
	client      *http.Client
	lastRequest time.Time
	token       string
	lock        *sync.Mutex
	mux         *http.ServeMux
}

type Params struct {
	UserAgent    string
	ClientId     string
	ClientSecret string
	RedirectUri  string
	LoginToken   string
	RefreshToken string
	Timeout      int64
}

type AsyncResult struct {
	c      chan interface{}
	result interface{}
}

func NewAsyncResult() *AsyncResult {
	out := new(AsyncResult)
	out.c = make(chan interface{}, 1)
	return out
}

func (this *AsyncResult) Wait() {
	if this.result == nil {
		this.result = <-this.c
	}
}

func (this *AsyncResult) Get() (interface{}, error) {
	this.Wait()
	switch t := this.result.(type) {
	case error:
		return nil, t

	case nil:
		return nil, nil

	default:
		return t, nil
	}
}

func (this *AsyncResult) Err() error {
	this.Wait()
	_, err := this.Get()
	return err
}

func New(config Params) *Client {
	out := new(Client)
	out.Params = config
	out.client = &http.Client{Timeout: time.Duration(out.Params.Timeout)}
	out.lock = &sync.Mutex{}
	out.mux = http.NewServeMux()
	return out
}

func (this *Client) TimeoutSeconds() time.Duration {
	return time.Duration(this.Timeout) * time.Second
}

func defaultParams() Params {
	var out Params
	out.Timeout = 5
	buf := bytes.NewBuffer(nil)
	buf.WriteString(runtime.GOOS)
	buf.WriteRune(':')
	buf.WriteString("godit-client")
	buf.WriteRune(':')
	buf.WriteString(VERSION)

	out.UserAgent = buf.String()

	return out
}

func LoadParamsFromFileName(fileName string) (Params, error) {
	var out Params
	f, err := os.Open(fileName)
	if err != nil {
		return out, err
	}
	defer f.Close()

	out, err = LoadParamsFromReader(f)

	return out, err
}

func LoadParamsFromReader(reader io.Reader) (Params, error) {
	jsdc := json.NewDecoder(reader)

	var out Params
	err := jsdc.Decode(&out)

	return out, err
}

func RedditIsUp() error {
	res, err := http.Get("https://www.reddit.com")
	if err != nil {
		return err
	}

	res.Body.Close()

	return nil
}
