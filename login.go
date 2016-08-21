package godit

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/facebookgo/httpdown"
	"github.com/skratchdot/open-golang/open"
)

type LoginResponse struct {
	Code  string
	State string
}

func (this *Client) LoginURL(state string, scopes []string, refreshable bool) string {
	this.lock.Lock()
	defer this.lock.Unlock()
	buf := bytes.NewBuffer(nil)
	buf.WriteString(redditApi)
	buf.WriteString(login)
	buf.WriteString("?client_id=")
	buf.WriteString(url.QueryEscape(this.ClientId))
	buf.WriteString("&response_type=code")
	buf.WriteString("&state=")
	buf.WriteString(url.QueryEscape(state))
	buf.WriteString("&redirect_uri=")
	buf.WriteString(url.QueryEscape(this.RedirectUri))
	buf.WriteString("&duration=")

	if refreshable {
		buf.WriteString("permanent")
	} else {
		buf.WriteString("temporary")
	}

	buf.WriteString("&scope=")
	buf.WriteString(url.QueryEscape(strings.Join(scopes, ",")))

	return buf.String()
}

func (this *Client) LoginWithBrowser(state string, scopes []string, refreshable bool) error {
	uri := this.LoginURL(state, scopes, refreshable)
	return open.Run(uri)
}

func (this *Client) SetOAuthToken(token string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.LoginToken = token
}

func (this *Client) StartLoginCallbackServer(callbackEndpoint string, port uint16, timeout time.Duration, expectedState string) *AsyncResult {

	this.lock.Lock()
	defer this.lock.Unlock()
	out := NewAsyncResult()

	servRes := make(chan bool, 1)
	timeOut := make(chan bool, 1)

	handFunc := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			servRes <- true
		}()

		errStr := r.FormValue("error")
		code := r.FormValue("code")
		state := r.FormValue("state")

		if errStr != "" {
			out.c <- errors.New(errStr)
			return
		}

		if state != expectedState {
			out.c <- fmt.Errorf("State %s did not match expected state %s", state, expectedState)
			return
		}

		out.c <- LoginResponse{
			Code:  code,
			State: state,
		}
	}

	hd := &httpdown.HTTP{
		StopTimeout: 1 * time.Second,
		KillTimeout: 1 * time.Second,
	}

	serv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(handFunc),
	}

	if s, err := hd.ListenAndServe(serv); err != nil {
		out.c <- err
	} else {
		go func() {
			time.Sleep(timeout * time.Second)
			timeOut <- true
			close(timeOut)
		}()

		go func() {
			select {
			case _ = <-servRes:
				// do nothing
			case _ = <-timeOut:
				out.c <- errors.New("Timeout occurred while waiting for login.")
			}

			s.Stop()
		}()
	}

	return out
}
