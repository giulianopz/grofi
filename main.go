package main

/*
refs:
- https://man.archlinux.org/man/rofi-script.5.en
- https://github.com/davatorium/rofi-scripts
*/

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/go-resty/resty/v2"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"
)

func main() {

	var searchQuery string

	if len(os.Args) == 2 {
		searchQuery = os.Args[1]
	} else {
		searchQuery = "golang"
	}

	if isPkgPath(searchQuery) {
		openPkgPage(searchQuery)
	}

	URI := fmt.Sprintf("https://pkg.go.dev/search?q=%s&limit=100&m=package#more-results", searchQuery)

	client := resty.New()

	resp, err := client.R().SetDoNotParseResponse(true).EnableTrace().Get(URI)
	defer resp.RawBody().Close()

	doc, err := html.Parse(resp.RawBody())
	if err != nil {
		log.Fatal(err)
	}

	searchResults := make([]string, 0)

	var f func(*html.Node)
	f = func(n *html.Node) {

		if n.Type == html.ElementNode && n.Data == "a" {
			for _, att := range n.Attr {
				if att.Key == "href" {
					url := att.Val
					if isPkgPath(url) {

						url := trimSubPath(url)

						if !slices.Contains(searchResults, url) {
							// TODO: order by imports
							searchResults = append(searchResults, url)
						}
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	fmt.Print(strings.Join(searchResults, "\n"))
}

func isPkgPath(url string) bool {
	// TODO: include other git hosting services
	return strings.HasPrefix(url, "/github.com") && !strings.Contains(url, "?")
}

func openPkgPage(pkgPath string) {
	xdgOpen, err := exec.LookPath("xdg-open")
	if err != nil {
		log.Fatal(err)
	}

	args := []string{"https://" + pkgPath}

	cmd := exec.Command(xdgOpen, args...)
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	} else {
		os.Exit(0)
	}
}

func trimSubPath(url string) string {
	if strings.Count(url, "/") > 3 {
		splitted := strings.Split(url, "/")
		url = strings.Join(splitted[:4], "/")
	}
	return url
}
