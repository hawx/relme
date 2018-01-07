package relme

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"hawx.me/code/assert"
)

func testPage(link string) string {
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

func TestFindVerified(t *testing.T) {
	assert := assert.New(t)

	var rURL, sURL string

	r := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, testPage(sURL))
	}))
	defer r.Close()
	rURL = r.URL

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, testPage(rURL))
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
    <a rel="what me ok" href="https://example.com/b">another</a>
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

func TestLinksToWithRedirects(t *testing.T) {
	// Although this isn't stated anywhere it seems that some sites (like Twitter)
	// wrap your rel="me" link with a short version, so this needs expanding
	//
	// In the real-world example I have on my homepage https://hawx.me
	//
	//     <a rel="me" href="https://twitter.com/hawx">
	//
	// But on https://twitter.com/hawx there is only
	//
	//     <a rel="me" href="https://t.co/qsNrcG2afz">
	//
	// So I need to follow this short link to check that _any_ page it redirects
	// to matches what I expect for my homepage.

	assert := assert.New(t)
	var twitterURL string

	// my homepage links to my twitter
	homepage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, testPage(twitterURL))
	}))
	defer homepage.Close()

	// tco redirects to my homepage
	tco := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, homepage.URL, http.StatusFound)
	}))
	defer tco.Close()

	// twitter has a link to tco
	twitter := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, testPage(tco.URL))
	}))
	twitterURL = twitter.URL
	defer twitter.Close()

	// then we can verify that my homepage links twitter
	ok, err := LinksTo(homepage.URL, twitter.URL)
	assert.Nil(err)
	assert.True(ok)

	// and twitter links to my homepage
	ok, err = LinksTo(twitter.URL, homepage.URL)
	assert.Nil(err)
	assert.True(ok)
}

func TestLinksToWithMoreRedirects(t *testing.T) {
	// Now take the example in TestLinksToWithRedirects but pretend that both
	// links resolve to redirects. So
	//
	//     https://example.com/my-homepage
	//       302 -> https://a-really-long-domain-name.com/me
	//        me -> https://twitter.com/me
	//
	//     https://twitter.com/me
	//        me -> https://tco.com/RANDOM
	//       302 -> https://example.com/my-homepage

	assert := assert.New(t)
	var twitterURL string

	// my homepage links to my twitter
	homepage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, testPage(twitterURL))
	}))
	defer homepage.Close()

	// my short homepage redirects to my homepage
	shortHomepage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, homepage.URL, http.StatusFound)
	}))
	defer shortHomepage.Close()

	// tco redirects to my short homepage
	tco := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, shortHomepage.URL, http.StatusFound)
	}))
	defer tco.Close()

	// twitter has a link to tco
	twitter := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, testPage(tco.URL))
	}))
	twitterURL = twitter.URL
	defer twitter.Close()

	// then we can verify that my homepage links twitter
	ok, err := LinksTo(homepage.URL, twitter.URL)
	assert.Nil(err)
	assert.True(ok)

	// and twitter links to my homepage
	ok, err = LinksTo(twitter.URL, homepage.URL)
	assert.Nil(err)
	assert.True(ok)

	// and we can verify that my short homepage links twitter
	ok, err = LinksTo(shortHomepage.URL, twitter.URL)
	assert.Nil(err)
	assert.True(ok)

	// and twitter links to my short homepage
	ok, err = LinksTo(twitter.URL, shortHomepage.URL)
	assert.Nil(err)
	assert.True(ok)
}
