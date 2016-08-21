package test

import (
	"strings"
	"testing"

	"github.com/nes77/godit"
	"github.com/stretchr/testify/assert"
)

const creationString string = `{
    "ClientId": "abcdefg",
    "ClientSecret": "secret",
    "Timeout": 5
    }`

func TestParamsLoad(t *testing.T) {
	p, err := godit.LoadParamsFromReader(strings.NewReader(creationString))

	assert.Nil(t, err)
	assert.Equal(t, "abcdefg", p.ClientId)
	assert.Equal(t, "secret", p.ClientSecret)
	assert.EqualValues(t, int64(5), p.Timeout)
}
