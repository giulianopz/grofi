package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
)

func main() {

	var searchQuery string

	if len(os.Args) == 2 {
		searchQuery = os.Args[1]
	} else {
		searchQuery = "golang"
	}

	URI := fmt.Sprintf("https://pkg.go.dev/search?q=%s", searchQuery)

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
					if strings.HasPrefix(url, "/github.com") && strings.Count(url, "/") == 3 && !strings.Contains(url, "?") {
						searchResults = append(searchResults, url)
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
