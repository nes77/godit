package test

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/nes77/godit"
	"github.com/stretchr/testify/assert"
)

var configParams *godit.Params

func TestMain(m *testing.M) {
	if _, err := os.Stat("config.json"); !os.IsNotExist(err) {
		c, err := godit.LoadParamsFromFileName("config.json")
		if err == nil {
			configParams = &c
		} else {
			log.Fatalln("Error while loading source file.")
		}
	}

	m.Run()
}

func TestLoginUri(t *testing.T) {
	client := godit.New(godit.Params{
		ClientId:     "abcdefg",
		ClientSecret: "secret",
		RedirectUri:  "https://localhost:8080",
	})

	uri := client.LoginURL("xxx", []string{"a", "b", "c"}, true)

	t.Log(uri)
	assert.Equal(t, uri,
		"https://www.reddit.com/api/v1/authorize?client_id=abcdefg&response_type=code&state=xxx&redirect_uri=https%3A%2F%2Flocalhost%3A8080&duration=permanent&scope=a%2Cb%2Cc")
}

func TestBrowserLogin(t *testing.T) {
	if configParams == nil {
		t.Log("Not testing due to missing config\n")
		return
	}

	if testing.Short() {
		t.Log("Not testing because short specified.\n")
	}

	t.Log("Testing browser login\n")

	client := godit.New(*configParams)
	as := client.StartLoginCallbackServer("auth_callback", 9999, 120, "xxx")
	err := client.LoginWithBrowser("xxx", []string{"vote"}, false)

	assert.Nil(t, err)
	assert.Nil(t, as.Err())

	l, _ := as.Get()
	assert.IsType(t, godit.LoginResponse{}, l)
	t.Logf("%v", l)

}

func TestTimeout(t *testing.T) {
	if configParams == nil {
		t.Log("Not testing due to missing config\n")
		return
	}

	if testing.Short() {
		t.Log("Not testing because short specified.\n")
	}

	t.Log("Testing browser login timeout\n")

	client := godit.New(*configParams)
	as := client.StartLoginCallbackServer("auth_callback", 9999, 10, "xxx")
	assert.NotNil(t, as.Err())

	err := as.Err()
	t.Log(err)
	assert.True(t, strings.Contains(err.Error(), "Timeout occurred"))
}
