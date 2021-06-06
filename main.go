package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
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

func main() {
	log.SetOutput(os.Stderr)
	// funcs := grabPG()
	funcs := grabSQLite()
	b, err := json.Marshal(funcs)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
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
