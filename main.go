// package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gocolly/colly/v2"
)

func main() {
	handler()
}

func handler() {
	// Create a context with a timeout of 15 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		scrape(ctx)
	}()

	select {
	case <-ctx.Done():
		// The scraping has reached the timeout
		fmt.Println("Scraping has completed or timed out.")
	}
}

func scrape(ctx context.Context) {
	// Create a new Collector
	c := colly.NewCollector()

	var (
		title          string
		documentNumber string
		feiEinNumber   string
		dateFiled      string
		state          string
		status         string
		lastEvent      string
		eventDateFiled string
		linkDetail     string
	)

	// Define a callback to be executed when a visited HTML element is found
	c.OnHTML(".large-width a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		// Visit the detailed card page
		linkDetail = "https://search.sunbiz.org" + link

		err := c.Visit(linkDetail)
		if err != nil {
			log.Println("Error visiting detailed card page:", err)
		}
	})

	var index = 1

	c.OnHTML("body", func(e *colly.HTMLElement) {
		title = e.DOM.Find(".detailSection.corporationName p").First().Text()

	})

	c.OnHTML("div.detailSection.filingInformation", func(e *colly.HTMLElement) {
		e.ForEach("label", func(_ int, el *colly.HTMLElement) {
			switch label := strings.TrimSpace(el.Text); label {
			case "Document Number":
				documentNumber = el.DOM.Next().Text()
			case "FEI/EIN Number":
				feiEinNumber = el.DOM.Next().Text()
			case "Date Filed":
				dateFiled = el.DOM.Next().Text()
			case "State":
				state = el.DOM.Next().Text()
			case "Status":
				status = el.DOM.Next().Text()
			case "Last Event":
				lastEvent = el.DOM.Next().Text()
			case "Event Date Filed":
				eventDateFiled = el.DOM.Next().Text()
			}
		})

		dateFiledTime, err := time.Parse("01/02/2006", dateFiled)
		if err != nil {
			log.Fatal(err)
		}

		eventDateFiledTime, err := time.Parse("01/02/2006", eventDateFiled)
		if err != nil {
			log.Fatal(err)
		}

		valuesToCheck := []string{"N/A", "NONE", "APPLIED FOR", "00-0000000"}

		if !containsString(valuesToCheck, feiEinNumber) && dateFiledTime.Year() <= 2016 && state == "FL" && status == "INACTIVE" && lastEvent == "ADMIN DISSOLUTION FOR ANNUAL REPORT" && eventDateFiledTime.Year() <= 2020 && strings.Contains(title, "Florida") {
			index++
			id := index - 1
			file, errr := excelize.OpenFile("scraping.xlsx")
			if errr != nil {
				log.Fatal(err)
			}
			var test = "Added to excel" + title
			fmt.Println(test)
			column := strconv.FormatInt(int64(index), 10)
			a_column := "A" + column
			b_column := "B" + column
			c_column := "C" + column
			d_column := "D" + column
			e_column := "E" + column
			f_column := "F" + column
			g_column := "G" + column
			h_column := "H" + column
			i_column := "I" + column
			j_column := "J" + column
			file.SetCellValue("Sheet1", a_column, id)
			file.SetCellValue("Sheet1", b_column, title)
			file.SetCellValue("Sheet1", c_column, documentNumber)
			file.SetCellValue("Sheet1", d_column, feiEinNumber)
			file.SetCellValue("Sheet1", e_column, dateFiled)
			file.SetCellValue("Sheet1", f_column, state)
			file.SetCellValue("Sheet1", g_column, status)
			file.SetCellValue("Sheet1", h_column, lastEvent)
			file.SetCellValue("Sheet1", i_column, eventDateFiledTime)
			file.SetCellValue("Sheet1", j_column, linkDetail)

			if err := file.SaveAs("scraping.xlsx"); err != nil { //checking for er>
				log.Fatal(err)
			}
		}
	})

	// Set up error handling
	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Request URL:", err)
	})

	// Define a callback to be executed when the "Next List" link is found
	c.OnHTML("span:has(a[title='Next List'])", func(e *colly.HTMLElement) {
		nextPageLink := e.ChildAttr("a", "href")
		if nextPageLink != "" {
			// Construct the full URL for the next page
			nextPageURL := "https://search.sunbiz.org" + nextPageLink

			fmt.Println("nextPageURL ---> " + nextPageURL)

			// Visit the next page
			err := c.Visit(nextPageURL)
			if err != nil {
				log.Println("Error visiting next page:", err)
			}
		}
	})

	// Start scraping
	url := "https://search.sunbiz.org/Inquiry/CorporationSearch/SearchResults/EntityName/a/Page1?searchNameOrder=A"
	err := c.Visit(url)
	if err != nil {
		log.Fatal(err)
	}

}

func containsString(slice []string, target string) bool {
	for _, value := range slice {
		if value == target {
			return true
		}
	}
	return false
}
