package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var pkgPathRegex *regexp.Regexp = regexp.MustCompile(`\(.*\)`)

func main() {

	retv := os.Getenv("ROFI_RETV")

	var (
		searchQuery string
		options     string
	)

	switch retv {
	case "0": // first call to script
		searchQuery = "http"
		options = "m=package&limit=3"
	case "1": // entry was selected
		{
			selectedEntry := strings.TrimSpace(os.Args[1])
			if pkgPathRegex.MatchString(selectedEntry) {
				selectedEntry = pkgNameCleaner.Replace(pkgPathRegex.FindStringSubmatch(selectedEntry)[0])
			}
			open(selectedEntry)
		}
	case "2": // input typed by user
		searchQuery = os.Args[1]
		options = "limit=100&m=package#more-results"
	}

	resp, err := http.Get("https://pkg.go.dev/search?q=" + searchQuery + "&" + options)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	searchResults, err := getSearchResults(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(strings.Join(searchResults, "\n"))
}

func open(pkgPath string) {
	if err := openWithDefaultBrowser("https://pkg.go.dev/" + pkgPath); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func openWithDefaultBrowser(url string) error {
	bin, err := exec.LookPath("xdg-open")
	if err != nil {
		return err
	}

	cmd := exec.Command(bin, url)
	if err = cmd.Start(); err != nil {
		return err
	}
	return nil
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

					bs := &bytes.Buffer{}
					getText(n, bs)

					simpleName, qualifiedName := getNames(bs)

					searchResults = append(searchResults,
						fmt.Sprintf("%s %s", simpleName, qualifiedName),
					)
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

func getNames(bs *bytes.Buffer) (string, string) {
	var (
		qualifiedName = bs.String()
		simpleName    = pkgNameCleaner.Replace(qualifiedName)
	)

	if strings.Contains(simpleName, "/") {
		splitted := strings.Split(simpleName, "/")
		simpleName = splitted[len(splitted)-1]
	}
	return simpleName, qualifiedName
}
