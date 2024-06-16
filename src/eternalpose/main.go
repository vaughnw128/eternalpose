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

func getCover(l string) string {
	return "lol"
}

func main() {
	c := colly.NewCollector()

	// Additional manga can be added here to alert on
	mangaSite := "https://tcbscans.com"
	mangaRegex := [2]string{"One Piece Chapter \\d{4}$", "Jujutsu Kaisen Chapter \\d{3}$"}

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if e.Text == "" {
			return
		}
		mangaTitle := cleanString(e.Text)
		for _, regexTerm := range mangaRegex {
			match, _ := regexp.MatchString(regexTerm, mangaTitle)
			if match {
				fmt.Printf("%s - %s\n", mangaTitle, mangaSite+link)
			}
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	err := c.Visit(mangaSite)
	if err != nil {
		fmt.Println("Error visiting TCBScans.")
	}
}
