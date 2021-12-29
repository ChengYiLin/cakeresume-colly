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
	Skills    string
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

	// Set up the Column
	baseColumns := []string{"Title", "Company", "Link", "MinSalary", "MaxSalary", "Currency"}
	frontendLanguagues := []string{"html", "css", "javascript", "typescript", "webpack", "jquery"}
	frontendFramworks := []string{"svelte", "react", "vue", "angular", "angularjs"}
	// cssFramworks := []string{"tailwind", "materialui", "antdesign", "bootstrap"}
	others := []string{"git", "unittest"}

	frontendSkills := [][]string{
		baseColumns,
		frontendLanguagues,
		frontendFramworks,
		others,
	}

	var columnList []string
	for _, r := range frontendSkills {
		columnList = append(columnList, r...)
	}

	// Setup CSV File
	csvData, _ := os.Create("urlCollector.csv")
	csvWriter := csv.NewWriter(csvData)
	ColumnNameList := [][]string{columnList}
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
			// Step 0. 取得頁面元素
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

		jobCollector.OnHTML(".labels .label", func(h *colly.HTMLElement) {
			recordData.Skills += h.Text + ", "
		})

		jobCollector.Visit(recordData.Link)

		csvWriter.Write([]string{
			recordData.Title,
			recordData.Company,
			recordData.Link,
			strconv.Itoa(recordData.MinSalary),
			strconv.Itoa(recordData.MaxSalary),
			recordData.Currency,
			recordData.Skills,
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
		if isUrlCollectionFinished {
			break
		}
		urlCollector.Visit(searchPageURlBase + strconv.Itoa(pageNum))
		pageNum++
	}

	urlCollector.Wait()
	defer csvData.Close()
}
