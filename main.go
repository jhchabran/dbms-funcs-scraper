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

const pgdocUrl = "https://www.postgresql.org/docs/12/functions.html"

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

	err := c.Visit(pgdocUrl)
	if err != nil {
		panic(err)
	}

	return funcs
}

func main() {
	log.SetOutput(os.Stderr)
	funcs := grabPG()
	b, err := json.Marshal(funcs)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
