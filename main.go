package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

type Func struct {
	Name        string
	Description string
	Theme       string
}

const pgDocsURL = "https://www.postgresql.org/docs/12/functions.html"
const sqlitleDocsURL = "https://www.sqlite.org/lang_corefunc.html"
const mysqlDocsURL = "https://dev.mysql.com/doc/refman/8.0/en/functions.html"

func main() {
	log.SetOutput(os.Stderr)
	// funcs := grabPG()
	// funcs := grabSQLite()
	funcs := grabMySQL()
	b, err := json.Marshal(funcs)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}

func grabMySQL() []Func {
	funcs := []Func{}
	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "Function") {
			err := e.Request.Visit(e.Attr("href"))
			if err != nil {
				if errors.Is(err, colly.ErrAlreadyVisited) {
					return
				}

				if errors.Is(err, colly.ErrMissingURL) {
					return
				}

				panic(err)
			}
		}
	})

	var theme string
	c.OnHTML(".titlepage", func(e *colly.HTMLElement) {
		numR := regexp.MustCompile(`((?:\d+\.?)+)`)
		theme = strings.TrimSpace(numR.ReplaceAllString(e.Text, ""))
	})

	c.OnHTML("tbody tr", func(e *colly.HTMLElement) {
		f := Func{Theme: theme}
		skip := false
		e.ForEachWithBreak("td", func(i int, td *colly.HTMLElement) bool {
			switch i {
			case 0:
				if strings.HasSuffix(td.Text, ")") {
					// Only crawl functions.
					f.Name = td.Text
				} else {
					// That's a statement, not a function.
					skip = true
					return false
				}
			case 1:
				f.Description = strings.TrimSpace(td.Text)
			}
			return true
		})
		if !skip {
			funcs = append(funcs, f)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		if !strings.Contains(r.URL.String(), "8") {
			r.Abort()
		}
		log.Println("Visiting", r.URL)
	})

	err := c.Visit(mysqlDocsURL)
	// err := c.Visit("https://dev.mysql.com/doc/refman/8.0/en/loadable-function-reference.html")
	if err != nil {
		panic(err)
	}

	return funcs
}

func grabSQLite() []Func {
	funcs := []Func{}
	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "functions") {
			err := e.Request.Visit(e.Attr("href"))
			if err != nil {
				if errors.Is(err, colly.ErrAlreadyVisited) {
					return
				}

				if errors.Is(err, colly.ErrMissingURL) {
					return
				}

				panic(err)
			}
		}
	})

	c.OnHTML("h1+dl", func(e *colly.HTMLElement) {
		theme := e.DOM.Prev().Text()
		theme = strings.TrimPrefix(theme[2:], " Descriptions of ")
		var f *Func
		e.ForEach("dt,dd", func(_ int, d *colly.HTMLElement) {
			switch d.Name {
			case "dt":
				// <dt> tag holds the function name.
				f = &Func{Theme: theme}
				f.Name = d.Text
				// In some cases, <dt> contains two func names, separated by a <br>.
				// When Text is called on these, the two func names will end up as:
				// "foo()bar(), so we need to adjust that.
				f.Name = strings.ReplaceAll(d.Text, ")", ") ")
				f.Name = strings.TrimSpace(f.Name)
			case "dd":
				// <dd> tag holds the description.
				f.Description = strings.TrimSpace(d.Text)
				funcs = append(funcs, *f)
				f = nil
			}
		})
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})

	err := c.Visit(sqlitleDocsURL)
	if err != nil {
		panic(err)
	}

	return funcs
}

func grabPG() []Func {
	funcs := []Func{}
	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "Functions") {
			err := e.Request.Visit(e.Attr("href"))
			if err != nil {
				if errors.Is(err, colly.ErrAlreadyVisited) {
					return
				}

				panic(err)
			}
		}
	})

	c.OnHTML("table[summary$=Functions] ", func(e *colly.HTMLElement) {
		f := Func{Theme: e.Attr("summary")}
		e.ForEach("tr", func(_ int, tr *colly.HTMLElement) {
			tr.ForEach("td", func(i int, td *colly.HTMLElement) {
				switch i {
				case 0:
					if name := td.ChildText("code[class=function]"); name != "" {
						// Sometimes, it's a <code><code> intead of just <code>
						// leading to doubled strings.
						f.Name = name
					} else {
						f.Name = td.ChildText("code")
					}
					log.Println(f.Name)
				case 1:
					f.Description = td.Text
				}
			})
		})

		funcs = append(funcs, f)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})

	err := c.Visit(pgDocsURL)
	if err != nil {
		panic(err)
	}

	return funcs
}
