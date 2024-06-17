package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Manga struct {
	Title          string   `json:"title"`
	Regex          string   `json:"regex"`
	Users          []string `json:"users"`
	CurrentChapter int      `json:"currentChapter"`
}

type Image struct {
	URL string `json:"url"`
}

type Embed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
	URL         string `json:"url"`
	Image       Image  `json:"image"`
}

type WebhookMessage struct {
	Content string  `json:"content"`
	Embeds  []Embed `json:"embeds"`
}

var (
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
)

func updateChapter(title string) {
	// Update the chapter number in manga.json
	var mangaData []Manga
	jsonFile, _ := os.Open("manga.json")
	jsonBytes, _ := io.ReadAll(jsonFile)
	err := json.Unmarshal(jsonBytes, &mangaData)
	if err != nil {
		logger.Error("Unable to unmarshal manga.json: ", err)
	}
	err = jsonFile.Close()
	if err != nil {
		logger.Error("Unable to close manga.json: ", err)
	}

	// Update the chapter number
	for i, manga := range mangaData {
		if manga.Title == title {
			mangaData[i].CurrentChapter += 1
		}
	}

	// Write updates to file
	jsonBytes, _ = json.Marshal(mangaData)
	err = os.WriteFile("manga.json", jsonBytes, 0644)
	if err != nil {
		logger.Error("Unable to write to manga.json: ", err)
	}

}

func cleanString(s string) string {
	// Simply cleans the string
	return strings.Join(strings.Fields(strings.Trim(s, " ")), " ")
}

func getCover(link string) string {
	// Grabs the cover with an additional collector
	c := colly.NewCollector()

	var pages []string

	// Get all pages URLs
	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		if strings.Contains(link, "cdn") {
			pages = append(pages, link)
		}
	})

	err := c.Visit(link)
	if err != nil {
		logger.Error("Error visiting the manga chapter page: ", err)
	}

	c.Wait()

	return pages[0]
}

func sendManga(title string, link string, users []string) {
	webhookURL := os.Getenv("WEBHOOK_URL")

	// Construct webhook message and send to discord
	coverLink := getCover(link)
	data := WebhookMessage{
		Content: strings.Join(users, ", "),
		Embeds: []Embed{
			{
				Title:       title,
				Description: "Yohoho! A new chapter has released!",
				Color:       16726860,
				URL:         link,
				Image: Image{
					URL: coverLink,
				},
			},
		},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("Unable to marshal JSON for webhook posting: ", err)
	}
	logger.Info(fmt.Sprintf("Sending %s to webhook.", title))
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Unable to send JSON: ", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("Unable to close response object: ", err)
		}
	}(resp.Body)
}

func main() {
	err := godotenv.Load()

	mangaSite := "https://tcbscans.com"

	logger.Info("Starting EternalPose")

	// Import manga data from file
	var mangaData []Manga

	jsonFile, _ := os.Open("manga.json")
	jsonBytes, _ := io.ReadAll(jsonFile)
	err = jsonFile.Close()
	if err != nil {
		logger.Error("Unable to close file: ", err)
		return
	}

	err = json.Unmarshal(jsonBytes, &mangaData)
	if err != nil {
		logger.Error("Unable to unmarshal JSON: ", err)
		return
	}

	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		mangaLink := mangaSite + e.Attr("href")
		mangaTitle := cleanString(e.Text)
		// Loops through the regex terms to see if each title matches desired manga
		for _, manga := range mangaData {
			r := regexp.MustCompile(manga.Regex)
			matches := r.FindStringSubmatch(mangaTitle)

			//If it matches, get chapter number, send the manga, then update chapter
			if matches != nil {
				chapterNumber, _ := strconv.Atoi(matches[r.SubexpIndex("Chapter")])
				if chapterNumber > manga.CurrentChapter {
					sendManga(mangaTitle, mangaLink, manga.Users)
					updateChapter(manga.Title)
				}
			}
		}
	})

	err = c.Visit(mangaSite)
	if err != nil {
		logger.Error("Unable to visit TCBScans: ", err)
		return
	}
}
