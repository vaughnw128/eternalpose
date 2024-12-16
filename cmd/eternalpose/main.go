package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"github.com/gocolly/colly"
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
	CurrentChapter float64  `json:"currentChapter"`
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
	mangaSite = "https://tcbscans.me"
	logger    = slog.New(slog.NewTextHandler(os.Stdout, nil))
)

func updateChapter(title string, cNum float64) {
	// Update the chapter number in manga.json
	var mangaData []Manga
	jsonFile, _ := os.Open("manga.json")
	jsonBytes, _ := io.ReadAll(jsonFile)
	err := json.Unmarshal(jsonBytes, &mangaData)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to unmarshal manga.json: %s", err))
	}
	err = jsonFile.Close()
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to close manga.json: %s", err))
	}

	// Update the chapter number
	for i, manga := range mangaData {
		if manga.Title == title {
			mangaData[i].CurrentChapter = cNum
		}
	}

	// Write updates to file
	jsonBytes, _ = json.Marshal(mangaData)
	err = os.WriteFile("manga.json", jsonBytes, 0644)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to write to manga.json: %s", err))
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
		logger.Error(fmt.Sprintf("Error visiting the manga chapter page: %s", err))
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
		logger.Error(fmt.Sprintf("Unable to marshal JSON for webhook posting: %s", err))
	}
	logger.Info(fmt.Sprintf("Sending %s to webhook.", title))
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to send JSON: %s", err))
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error(fmt.Sprintf("Unable to close response object: %s", err))
		}
	}(resp.Body)
}

func scrapeManga() {
	// Import manga data from file
	var mangaData []Manga

	jsonFile, _ := os.Open("manga.json")
	jsonBytes, _ := io.ReadAll(jsonFile)
	err := jsonFile.Close()
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to close file: %s", err))
	}

	err = json.Unmarshal(jsonBytes, &mangaData)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to unmarshal JSON: %s", err))
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
				chapterNumber, _ := strconv.ParseFloat(matches[r.SubexpIndex("Chapter")], 64)
				if chapterNumber > manga.CurrentChapter {
					sendManga(mangaTitle, mangaLink, manga.Users)
					updateChapter(manga.Title, chapterNumber)
				}
			}
		}
	})

	c.OnRequest(func(r *colly.Request) {
		logger.Info(fmt.Sprintf("Visiting %s", r.URL.String()))
	})

	err = c.Visit(mangaSite)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to visit TCBScans: %s", err))
	}
}

func main() {

	logger.Info("Starting EternalPose")

	// Initialize cron scheduler
	s, err := gocron.NewScheduler()
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to start scheduler: %s", err))
	}

	_, err = s.NewJob(
		gocron.CronJob(
			"0 * * * *",
			false,
		),
		gocron.NewTask(
			scrapeManga,
		),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to create job: %s", err))
	}

	// Log the job
	logger.Info("Manga scraping cron job started [0 * * * *]")

	// start the scheduler
	s.Start()

	select {}
}
