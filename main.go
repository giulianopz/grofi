package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type PkgSearchResponse struct {
	Items []struct {
		PackagePath string `json:"packagePath"`
		ModulePath  string `json:"modulePath"`
		Version     string `json:"version"`
		Synopsis    string `json:"synopsis"`
	} `json:"items"`
	Total int `json:"total"`
}

var pkgPathRegex *regexp.Regexp = regexp.MustCompile(`\(.*\)`)

func main() {

	retv := os.Getenv("ROFI_RETV")

	switch retv {
	case "0": // first call to script
		fmt.Print("")
	case "1": // entry was selected
		{
			entry := strings.TrimSpace(os.Args[1])
			pkgPath := strings.NewReplacer("(", "", ")", "").Replace(pkgPathRegex.FindStringSubmatch(entry)[0])
			open(pkgPath)
		}
	case "2": // input typed by user
		searchQuery := os.Args[1]

		resp, err := http.Get("https://pkg.go.dev/v1/search?q=" + searchQuery + "&limit=100")
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		bs, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		pkgSearchResp := &PkgSearchResponse{}
		err = json.Unmarshal(bs, &pkgSearchResp)
		if err != nil {
			log.Fatal(err)
		}
		if len(pkgSearchResp.Items) == 0 {
			fmt.Print("no results...")
			return
		}
		var searchResults []string
		for _, i := range pkgSearchResp.Items {
			searchResults = append(searchResults, fmt.Sprintf("%s (%s)", filepath.Base(i.PackagePath), i.PackagePath))
		}
		fmt.Print(strings.Join(searchResults, "\n"))
	default:
		log.Default().Println("cannot handle retv=" + retv)
	}
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
	return exec.Command(bin, url).Start()
}
