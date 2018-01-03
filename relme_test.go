package relme

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"hawx.me/code/assert"
)

func TestFinydVerified(t *testing.T) {
	assert := assert.New(t)

	html := func(link string) string {
		return `
<!doctype html>
<html>
<head>

</head>
<body>
  <a rel="me" href="` + link + `">ok</a>
  <a rel="me" href="http://localhost/unknown">what</a>
</body>
`
	}

	var rURL, sURL string

	r := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, html(sURL))
	}))
	defer r.Close()
	rURL = r.URL

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, html(rURL))
	}))
	defer s.Close()
	sURL = s.URL

	links, err := FindVerified(s.URL)
	assert.Nil(err)

	if assert.Len(links, 1) {
		assert.Equal(links[0], r.URL)
	}
}

func TestFind(t *testing.T) {
	assert := assert.New(t)

	html := `
<!doctype html>
<html>
<head>

</head>
<body>
  <a rel="me" href="https://example.com/a">what</a>
  <div>
    <a rel="me" href="https://example.com/b">another</a>
  </div>
</body>
`

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, html)
	}))
	defer s.Close()

	links, err := Find(s.URL)
	assert.Nil(err)

	if assert.Len(links, 2) {
		assert.Equal(links[0], "https://example.com/a")
		assert.Equal(links[1], "https://example.com/b")
	}
}

func TestLinksTo(t *testing.T) {
	assert := assert.New(t)

	html := `
<!doctype html>
<html>
<head>

</head>
<body>
  <a rel="me" href="https://example.com/a">ok</a>
</body>
`

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, html)
	}))
	defer s.Close()

	ok, err := LinksTo(s.URL, "https://example.com/a")
	assert.Nil(err)
	assert.True(ok)
	ok, err = LinksTo(s.URL, "https://example.com/a/")
	assert.Nil(err)
	assert.True(ok)
	ok, err = LinksTo(s.URL, "http://example.com/a")
	assert.Nil(err)
	assert.True(ok)

	ok, err = LinksTo(s.URL, "https://example.com/b")
	assert.Nil(err)
	assert.False(ok)
}
