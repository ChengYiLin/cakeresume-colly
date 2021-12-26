package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

type SalaryData struct {
	Title     string
	Company   string
	Link      string
	MinSalary int
	MaxSalary int
	Currency  string
}

func getSalaryFromText(salaryText string, timeUnit string) int {
	moneyUnit := string(salaryText[len(salaryText)-1])
	moneyNum, err := strconv.ParseFloat(salaryText[:len(salaryText)-1], 32)
	if err != nil {
		log.Fatal(err)
	}

	if timeUnit == "year" {
		moneyNum = moneyNum / 12
	}

	switch moneyUnit {
	case "K":
		return int(moneyNum * 1000)

	case "M":
		return int(moneyNum * 1000000)

	default:
		return int(moneyNum)
	}
}

func main() {
	var searchPageURlBase string = "https://www.cakeresume.com/jobs?ref=navs_jobs&refinementList[profession][0]=it_front-end-engineer&refinementList[job_type][0]=full_time&page="
	var pageNum int = 1
	var isUrlCollectionFinished bool = false

	// Setup CSV File
	csvData, _ := os.Create("urlCollector.csv")
	csvWriter := csv.NewWriter(csvData)
	ColumnNameList := [][]string{
		{"Title", "Company", "Link", "MinSalary", "MaxSalary", "Currency"},
	}
	csvWriter.WriteAll(ColumnNameList)

	// Create Colly Collector For CakeResume
	urlCollector := colly.NewCollector(
		colly.AllowedDomains("www.cakeresume.com"),
	)

	urlCollector.OnRequest(func(r *colly.Request) {
		fmt.Printf("Visiting %s\n------------------\n", r.URL.String())
	})

	// Check Search Result is null
	urlCollector.OnHTML(".jobs-search-container .no-result", func(h *colly.HTMLElement) {
		fmt.Println("No Result, Time To Break")
		isUrlCollectionFinished = true
	})

	// Get Job URL
	urlCollector.OnHTML(".job-list-item-content", func(h *colly.HTMLElement) {
		var recordData SalaryData
		jobCollector := urlCollector.Clone()

		jobTitleDOM := h.DOM.Find(".job-link")
		jobCompanyDOM := h.DOM.Find(".page-name a")

		recordData.Link, _ = jobTitleDOM.Attr("href")
		recordData.Title = jobTitleDOM.Text()
		recordData.Company = jobCompanyDOM.Text()

		// Web Scrape Job Detail
		jobCollector.OnHTML(".job-meta-section", func(h *colly.HTMLElement) {
			// 	// Step 0. 取得頁面元素
			salaryText := h.DOM.Find(".job-salary").Text()
			if len(salaryText) == 0 {
				recordData.MinSalary = 0
				recordData.MaxSalary = 0
				recordData.Currency = "TWD"
				return
			}

			// Step 1. 取得時間單位
			textForTimeUnit := strings.Split(salaryText, "/")
			timeUnit := textForTimeUnit[1]

			// Step 2. 取得 Currency
			textForCurrency := strings.Split(textForTimeUnit[0], " ")
			recordData.Currency = textForCurrency[len(textForCurrency)-1]

			// Step 3. 取得 最高 及 最低 月薪
			if strings.Contains(textForCurrency[0], "+") {
				minSalaryText := strings.ReplaceAll(textForCurrency[0], "+", "")

				recordData.MinSalary = getSalaryFromText(minSalaryText, timeUnit)
				recordData.MaxSalary = 0
			} else {
				minSalaryText := textForCurrency[0]
				maxSalaryText := textForCurrency[2]

				recordData.MinSalary = getSalaryFromText(minSalaryText, timeUnit)
				recordData.MaxSalary = getSalaryFromText(maxSalaryText, timeUnit)
			}
		})

		jobCollector.Visit(recordData.Link)

		fmt.Println("Write")
		csvWriter.Write([]string{
			recordData.Title,
			recordData.Company,
			recordData.Link,
			strconv.Itoa(recordData.MinSalary),
			strconv.Itoa(recordData.MaxSalary),
			recordData.Currency,
		})
		csvWriter.Flush()
	})

	urlCollector.Limit(&colly.LimitRule{
		DomainGlob: "*cakeresume.*",
		// Delay:      1 * time.Second,
	})

	urlCollector.OnError(func(r *colly.Response, e error) {
		isUrlCollectionFinished = true
		fmt.Println("Request URL : ", r.Request.URL, "\nError : ", e)
	})

	for {
		if pageNum == 4 {
			break
		}

		if isUrlCollectionFinished {
			break
		}
		urlCollector.Visit(searchPageURlBase + strconv.Itoa(pageNum))
		pageNum++
	}

	urlCollector.Wait()
	defer csvData.Close()

	// 	// 技能 tag
	// 	h.DOM.Find(".labels .label").Each(func(i int, s *goquery.Selection) {
	// 		// fmt.Println(s.Text())
	// 	})
	// })
}
