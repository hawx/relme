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

var Client = http.DefaultClient

// FindVerified takes a profile URL and returns a list of all hrefs in <a
// rel="me"/> elements on the page that also link back to the profile.
func FindVerified(profile string) (links []string, err error) {
	profileLinks, err := Find(profile)
	if err != nil {
		return
	}

	for _, link := range profileLinks {
		if ok, err := LinksTo(link, profile); err == nil && ok {
			links = append(links, link)
		}
	}

	return
}

// Find takes a profile URL and returns a list of all hrefs in <a rel="me"/>
// elements on the page.
func Find(profile string) (links []string, err error) {
	req, err := http.NewRequest("GET", profile, nil)
	if err != nil {
		return
	}

	resp, err := Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return parseLinks(resp.Body)
}

// LinksTo takes a remote profile URL and checks whether any of the hrefs in <a
// rel="me"/> elements match the test URL.
func LinksTo(remote, test string) (ok bool, err error) {
	testURL, err := url.Parse(test)
	if err != nil {
		return
	}
	testURL.Scheme = "https"

	links, err := Find(remote)
	if err != nil {
		return
	}

	for _, link := range links {
		linkURL, err := url.Parse(link)
		if err != nil {
			continue
		}
		linkURL.Scheme = "https"

		if strings.TrimRight(linkURL.String(), "/") == strings.TrimRight(testURL.String(), "/") {
			return true, nil
		}
	}

	return false, nil
}

func parseLinks(r io.Reader) (links []string, err error) {
	root, err := html.Parse(r)
	if err != nil {
		return
	}

	rels := searchAll(root, isRelMe)
	for _, node := range rels {
		links = append(links, getAttr(node, "href"))
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

func isRelMe(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "a" {
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
