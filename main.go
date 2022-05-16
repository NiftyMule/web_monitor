package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"net/url"
	"strings"
	"time"
)

const ConfigPath = "config.json"
const DbPath = "db.json"

var CheckInterval int // check source interval in minutes

type Config struct {
	CheckInterval int          `json:"checkInterval"`
	Sources       []SourceConf `json:"sources"`
}

type SourceConf struct {
	Name       string        `json:"name"`
	Active     bool          `json:"active"`
	Url        string        `json:"url"`
	ItemPath   string        `json:"itemPath"`
	TitlePath  string        `json:"titlePath"`
	FooterPath string        `json:"footerPath"`
	Contents   []ContentConf `json:"contents"`
}

type ContentConf struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	ContentType string `json:"type"`
}

type Item struct {
	Title    string    `json:"title"`
	Source   string    `json:"source"`
	Contents []Content `json:"contents"`
}

type Content struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (i Item) equals(item Item) bool {
	if i.Title == item.Title && contentArrayEqual(i.Contents, item.Contents) {
		return true
	}
	return false
}

func (i Item) in(items []Item) bool {
	for _, item := range items {
		if i.equals(item) {
			return true
		}
	}
	return false
}

func (i Item) print() {
	fmt.Println("========================================")
	fmt.Printf("%-12s - %s\n", "Source", i.Source)
	fmt.Printf("%-12s - %s\n", "Title", i.Title)
	for _, content := range i.Contents {
		if len(content.Value) != 0 {
			fmt.Printf("%-12s - %s\n", content.Name, content.Value)
		}
	}
	fmt.Println()
}

func (c Content) in(contents []Content) bool {
	// we don't compare URLs because it might be generated dynamically
	if len(c.Value) > 4 && c.Value[:4] == "http" {
		return true
	}
	for _, content := range contents {
		if c == content {
			return true
		}
	}
	return false
}

func contentArrayEqual(a []Content, b []Content) bool {
	for _, c := range a {
		if !c.in(b) {
			return false
		}
	}
	return true
}

func loadConfig(filepath string) (source []SourceConf, err error) {
	fileData, err := ioutil.ReadFile(filepath)
	if err == nil {
		var c Config
		err = json.Unmarshal(fileData, &c)
		CheckInterval = c.CheckInterval
		source = c.Sources
	}
	return
}

func checkSource(source SourceConf) (result []Item) {
	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	var res = ""

	// use chromedp to simulate browser (load and execute Javascripts, render dynamic contents)
	err := chromedp.Run(ctx,
		chromedp.Navigate(source.Url),
		chromedp.ScrollIntoView(source.FooterPath),
		chromedp.WaitReady(source.ItemPath+" "+source.TitlePath),
		chromedp.OuterHTML(`html`, &res, chromedp.ByQuery),
	)

	if err == nil {
		// goquery only accept reader to create document
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(res))
		if err == nil {
			doc.Find(source.ItemPath).Each(func(i int, s *goquery.Selection) {
				var item Item
				// Get title
				item.Title = strings.Trim(s.Find(source.TitlePath).Text(), " \t\r\n")
				// Get custom contents
				item.Contents = make([]Content, len(source.Contents))
				for j, contentConf := range source.Contents {
					item.Contents[j].Name = contentConf.Name
					switch contentConf.ContentType {
					case "text":
						item.Contents[j].Value = strings.Trim(s.Find(contentConf.Path).Text(), " \t\r\n")
					case "url":
						item.Contents[j].Value =
							s.Find(contentConf.Path).AttrOr("href", "No link found!")
						if len(item.Contents[j].Value) > 0 && item.Contents[j].Value[0] == '/' {
							u, _ := url.Parse(source.Url)
							item.Contents[j].Value = u.Scheme + "://" + u.Host + item.Contents[j].Value
						}
					case "list":
						var tmp string
						for i, listItem := range s.Find(contentConf.Path).Nodes {
							if i != 0 {
								tmp += fmt.Sprintf("\n%15s", "")
							}
							tmp += listItem.FirstChild.Data
						}
						item.Contents[j].Value = tmp
					case "list-inline":
						var tmp string
						for i, listItem := range s.Find(contentConf.Path).Nodes {
							if i != 0 {
								tmp += fmt.Sprintf("  ")
							}
							tmp += listItem.FirstChild.Data
						}
						item.Contents[j].Value = tmp
					}
				}
				if len(item.Title) > 0 {
					item.Source = source.Name
					result = append(result, item)
				}
			})
		}
	}
	return
}

func checkDaemon(sources []SourceConf) {
	db := make(map[string][]Item)
	fileData, _ := ioutil.ReadFile(DbPath)
	json.Unmarshal(fileData, &db)

	for {
		var toPrint []Item
		for _, source := range sources {
			if source.Active {
				items := checkSource(source)
				for _, item := range items {
					if !item.in(db[source.Name]) {
						db[source.Name] = append(db[source.Name], item)
						toPrint = append(toPrint, item)
					}
				}
			}
		}
		if len(toPrint) > 0 {
			fmt.Println()
			fmt.Println("**********************************************************")
			fmt.Println("************************ Refresh *************************")
			fmt.Printf("****************** %s *******************\n", time.Now().Format(time.RFC822))
			fmt.Println("**********************************************************")
			fmt.Println()
			for _, item := range toPrint {
				item.print()
			}
			data, _ := json.MarshalIndent(db, "", "    ")
			ioutil.WriteFile(DbPath, data, 0777)
		}
		time.Sleep(time.Duration(CheckInterval) * time.Minute)
	}
}

func main() {
	sources, err := loadConfig(ConfigPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	go checkDaemon(sources)

	for {
		var input string
		fmt.Scanf("%s\n", &input)

		if input == "quit" {
			break
		}
	}
}
