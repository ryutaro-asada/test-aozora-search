package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
)

var pageURLFormat = "https://www.aozora.gr.jp/cards/%s/card%s.html"

type Entry struct {
	AuthorID string
	Author   string
	TitleID  string
	Title    string
	SiteURL  string
	ZipURL   string
}

func findAuthorAndZIP(siteURL string) (string, string) {
	//log.Println("findAuthorAndZIP")
	doc, err := goquery.NewDocument(siteURL)
	if err != nil {
		return "", ""
	}
	author := doc.Find("table[summary=作家データ] tr:nth-child(2) td:nth-child(2)").First().Text()

	zipURL := ""
	doc.Find("table.download a").Each(func(n int, elem *goquery.Selection) {
		href := elem.AttrOr("href", "")
		if strings.HasSuffix(href, ".zip") {
			zipURL = href
		}
	})
	//log.Println("zipURL", zipURL)
	if zipURL == "" {
		return author, ""
	}
	if strings.HasPrefix(zipURL, "http://") || strings.HasPrefix(zipURL, "https://") {
		return author, zipURL
	}

	u, err := url.Parse(siteURL)
	if err != nil {
		return author, ""
	}
	u.Path = path.Join(path.Dir(u.Path), zipURL)
	return author, u.String()
}

func findEntries(siteURL string) ([]Entry, error) {
	log.Println("findEntries")
	doc, err := goquery.NewDocument(siteURL)
	if err != nil {
		return nil, err
	}
	// get card num
	pat := regexp.MustCompile(`.*/cards/([0-9]+)/card([0-9]+).html$`)
	entries := []Entry{}
	// ol -> li -> a
	// ex:
	// <ol>
	// <li><a href="../cards/000879/card4872.html">愛読書の印象</a>　（新字旧仮名、作品ID：4872）　</li>
	doc.Find("ol li a").Each(func(n int, elem *goquery.Selection) {
		// get href=""
		token := pat.FindStringSubmatch(elem.AttrOr("href", ""))
		if len(token) != 3 {
			return
		}
		title := elem.Text()
		//log.Println("result", title, token[1], token[2])
		pageURL := fmt.Sprintf(pageURLFormat, token[1], token[2])
		//log.Println(title, pageURL)
		author, zipURL := findAuthorAndZIP(pageURL) // ZIP ファイルの URL を得る
		if zipURL != "" {
			entries = append(entries, Entry{
				AuthorID: token[1],
				Author:   author,
				TitleID:  token[2],
				Title:    title,
				SiteURL:  siteURL,
				ZipURL:   zipURL,
			})
		}
	})
	return entries, nil
}

func extractText(zipURL string) (string, error) {

	log.Println("query", zipURL)
	resp, err := http.Get(zipURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	for _, file := range r.File {
		if path.Ext(file.Name) == ".txt" {
			f, err := file.Open()
			if err != nil {
				return "", err
			}
			b, err := ioutil.ReadAll(f)
			f.Close()
			if err != nil {
				return "", err
			}
			b, err = japanese.ShiftJIS.NewDecoder().Bytes(b)
			if err != nil {
				return "", err
			}
			return string(b), nil
		}
	}
	return "", errors.New("contents not found")
}

func main() {
	listURL := "https://www.aozora.gr.jp/index_pages/person879.html"
	entries, err := findEntries(listURL)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		log.Printf("adding %+v\n", entry)
		content, err := extractText(entry.ZipURL)
		fmt.Println(content)
		if err != nil {
			log.Println(err)
		}
	}
	//for _, entry := range entries {
	//	fmt.Println(entry.Title, entry.ZipURL)
	//}
}
