package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"regexp"
	"strings"
)

func cleanString(s string) string {
	return strings.Join(strings.Fields(strings.Trim(s, " ")), " ")
}

func main() {
	c := colly.NewCollector()

	mangaRegex := [2]string{"One Piece Chapter \\d{4}$", "Jujutsu Kaisen Chapter \\d{3}$"}

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// link := e.Attr("href")
		if e.Text == "" {
			return
		}
		mangaTitle := cleanString(e.Text)
		for _, regexTerm := range mangaRegex {
			match, _ := regexp.MatchString(regexTerm, mangaTitle)
			if match {
				fmt.Println(mangaTitle)
			}
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	err := c.Visit("https://tcbscans.com/")
	if err != nil {
		fmt.Println("Error visiting TCBScans.")
	}
}
