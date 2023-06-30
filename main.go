package main

/*
refs:
- https://man.archlinux.org/man/rofi-script.5.en
- https://github.com/davatorium/rofi-scripts
*/

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
)

func main() {

	retv := os.Getenv("ROFI_RETV")

	var searchQuery string

	switch retv {
	case "0":
		searchQuery = "strings"
	case "1":
		open(os.Args[1])
	case "2":
		searchQuery = os.Args[1]
	}

	URI := fmt.Sprintf("https://pkg.go.dev/search?q=%s&limit=100&m=package#more-results", searchQuery)

	client := resty.New()

	resp, err := client.R().SetDoNotParseResponse(true).Get(URI)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.RawBody().Close()

	searchResults, err := getSearchResults(resp.RawBody())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(strings.Join(searchResults, "\n"))
}

func open(pkgPath string) {
	bin, err := exec.LookPath("xdg-open")
	if err != nil {
		log.Fatal(err)
	}

	args := []string{"https://pkg.go.dev/" + pkgPath}

	cmd := exec.Command(bin, args...)
	if err = cmd.Start(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

var pkgNameCleaner *strings.Replacer = strings.NewReplacer("(", "", ")", "")

func getText(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		getText(c, buf)
	}
}

func getSearchResults(htmlDoc io.Reader) ([]string, error) {

	htmlTree, err := html.Parse(htmlDoc)
	if err != nil {
		return nil, err
	}

	searchResults := make([]string, 0)

	var f func(*html.Node)
	f = func(n *html.Node) {

		if n.Type == html.ElementNode && n.Data == "span" {
			for _, att := range n.Attr {
				if att.Key == "class" && att.Val == "SearchSnippet-header-path" {

					text := &bytes.Buffer{}
					getText(n, text)

					qualifiedName := text.String()
					simpleName := pkgNameCleaner.Replace(qualifiedName)
					if qualifiedName != "" && simpleName != "" {
						searchResults = append(searchResults, fmt.Sprintf("%s %s", simpleName, qualifiedName))
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(htmlTree)

	return searchResults, nil
}
