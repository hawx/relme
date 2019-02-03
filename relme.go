// Package relme provides functions to retrieve and verify profiles marked up
// with the rel="me" microformat.
//
// See http://microformats.org/wiki/rel-me
package relme

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

var relMe = &RelMe{Client: http.DefaultClient}

func Find(profile string) (links []string, err error) {
	return relMe.Find(profile)
}

func FindAuth(profile string) (links []string, err error) {
	return relMe.FindAuth(profile)
}

func Verify(profile string, links []string) (verifiedLinks []string, err error) {
	return relMe.Verify(profile, links)
}

func LinksTo(remote, test string) (ok bool, err error) {
	return relMe.LinksTo(remote, test)
}

type RelMe struct {
	Client *http.Client
}

// Find takes a profile URL and returns a list of all hrefs in <a rel="me"/>
// elements on the page.
func (me *RelMe) Find(profile string) (links []string, err error) {
	req, err := http.NewRequest("GET", profile, nil)
	if err != nil {
		return
	}

	resp, err := me.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return parseLinks(resp.Body, isRelMe)
}

// FindAuth takes a profile URL and returns a list of all hrefs in <a rel="me
// authn"/> elements on the page, if none are found then it will return a list
// of all hrefs in <a rel="me"/> elements instead.
func (me *RelMe) FindAuth(profile string) (links []string, err error) {
	req, err := http.NewRequest("GET", profile, nil)
	if err != nil {
		return
	}

	resp, err := me.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return parseLinks(resp.Body, isRelAuthn, isRelMe)
}

// Verify takes a profile URL and a list of profile links, and returns a list of
// all hrefs in <a rel="me"/> elements on the page that also link back to the
// profile.
func (me *RelMe) Verify(profile string, links []string) (verifiedLinks []string, err error) {
	for _, link := range links {
		if ok, err := me.LinksTo(link, profile); err == nil && ok {
			verifiedLinks = append(verifiedLinks, link)
		}
	}

	return
}

// LinksTo takes a remote profile URL and checks whether any of the hrefs in <a
// rel="me"/> elements match the test URL.
func (me *RelMe) LinksTo(remote, test string) (ok bool, err error) {
	testURL, err := url.Parse(test)
	if err != nil {
		return
	}

	testRedirects, err := follow(testURL)
	if err != nil {
		return
	}

	links, err := me.Find(remote)
	if err != nil {
		return
	}

	for _, link := range links {
		linkURL, err := url.Parse(link)
		if err != nil {
			continue
		}

		linkRedirects, err := follow(linkURL)
		if err != nil {
			continue
		}

		if compare(linkRedirects, testRedirects) {
			return true, nil
		}
	}

	return false, nil
}

func normalizeInPlace(urls []string) {
	for i, a := range urls {
		aURL, err := url.Parse(a)
		if err != nil {
			continue
		}
		aURL.Scheme = "https"

		urls[i] = strings.TrimRight(aURL.String(), "/")
	}
}

func compare(as, bs []string) bool {
	normalizeInPlace(as)
	normalizeInPlace(bs)

	for _, a := range as {
		for _, b := range bs {
			if a == b {
				return true
			}
		}
	}

	return false
}

func follow(remote *url.URL) (redirects []string, err error) {
	noRedirectClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	previous := map[string]struct{}{}
	current := remote

	for {
		redirects = append(redirects, current.String())

		req, err := http.NewRequest("GET", current.String(), nil)
		if err != nil {
			break
		}
		previous[current.String()] = struct{}{}

		resp, err := noRedirectClient.Do(req)
		if err != nil {
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode < 300 || resp.StatusCode >= 400 {
			break
		}

		current, err = current.Parse(resp.Header.Get("Location"))
		if err != nil {
			break
		}

		if _, ok := previous[current.String()]; ok {
			break
		}
	}

	return
}

func parseLinks(r io.Reader, preds ...func(*html.Node) bool) (links []string, err error) {
	root, err := html.Parse(r)
	if err != nil {
		return
	}

	for _, pred := range preds {
		rels := searchAll(root, pred)
		for _, node := range rels {
			links = append(links, getAttr(node, "href"))
		}
		if len(links) > 0 {
			return
		}
	}

	return
}

func searchAll(node *html.Node, pred func(*html.Node) bool) (results []*html.Node) {
	if pred(node) {
		results = append(results, node)
		return
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		result := searchAll(child, pred)
		if len(result) > 0 {
			results = append(results, result...)
		}
	}

	return
}

func isRelAuthn(node *html.Node) bool {
	var me, authn bool
	if node.Type == html.ElementNode && (node.Data == "a" || node.Data == "link") {
		rels := strings.Split(getAttr(node, "rel"), " ")
		for _, rel := range rels {
			if rel == "me" {
				me = true
			}
			if rel == "authn" {
				authn = true
			}
			if me && authn {
				return true
			}
		}
	}

	return false
}

func isRelMe(node *html.Node) bool {
	if node.Type == html.ElementNode && (node.Data == "a" || node.Data == "link") {
		rels := strings.Split(getAttr(node, "rel"), " ")
		for _, rel := range rels {
			if rel == "me" {
				return true
			}
		}
	}

	return false
}

func getAttr(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}

	return ""
}
