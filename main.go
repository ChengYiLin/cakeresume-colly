package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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

type FrontendSkills struct {
	html       bool
	css        bool
	javascript bool
	typescript bool
	jquery     bool
	svelte     bool
	react      bool
	vue        bool
	angular    bool
	angularjs  bool
	git        bool
	unittest   bool
	webpack    bool
}

func getFrontendSkillList() [][]string {
	frontendLanguagues := []string{"html", "css", "javascript", "typescript", "jquery"}
	frontendFramworks := []string{"svelte", "react", "vue", "angular", "angularjs"}
	// cssFramworks := []string{"tailwind", "materialui", "antdesign", "bootstrap"}
	others := []string{"git", "unittest", "webpack"}

	return [][]string{
		frontendLanguagues,
		frontendFramworks,
		others,
	}
}

func getSalaryFromText(salaryText string, timeUnit string) int {
	moneyUnit := string(salaryText[len(salaryText)-1])
	moneyNum, err := strconv.ParseFloat(salaryText[:len(salaryText)-1], 32)
	if err != nil {
		return 0
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

func extractSkillsFromText(text string, skillsList []string, skillsData map[string]string) map[string]string {
	data := strings.ToLower(text)

	for _, skill := range skillsList {
		if skillsData[skill] == "1" {
			continue
		}

		if strings.Contains(data, skill) {
			skillsData[skill] = "1"
			continue
		}

		skillsData[skill] = "0"
	}

	return skillsData
}

func main() {
	var searchPageURlBase string = "https://www.cakeresume.com/jobs?ref=navs_jobs&refinementList[profession][0]=it_front-end-engineer&refinementList[job_type][0]=full_time&page="
	var pageNum int = 1
	var isUrlCollectionFinished bool = false

	// Set up the Column
	baseColumns := []string{"Title", "Company", "Link", "MinSalary", "MaxSalary", "Currency"}
	frontendSkills := getFrontendSkillList()

	var skillsList []string
	for _, r := range frontendSkills {
		skillsList = append(skillsList, r...)
	}

	// Setup CSV File
	csvData, _ := os.Create("urlCollector.csv")
	csvWriter := csv.NewWriter(csvData)
	ColumnNameList := [][]string{append(baseColumns, skillsList...)}
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
		var skillsData = map[string]string{
			"html":       "0",
			"css":        "0",
			"javascript": "0",
			"typescript": "0",
			"jquery":     "0",
			"svelte":     "0",
			"react":      "0",
			"vue":        "0",
			"angular":    "0",
			"angularjs":  "0",
			"git":        "0",
			"unittest":   "0",
			"webpack":    "0",
		}

		jobCollector := urlCollector.Clone()

		jobTitleDOM := h.DOM.Find(".job-link")
		jobCompanyDOM := h.DOM.Find(".page-name a")

		// Get Job Inform
		recordData.Link, _ = jobTitleDOM.Attr("href")
		recordData.Title = jobTitleDOM.Text()
		recordData.Company = jobCompanyDOM.Text()

		// Get Job Salary
		jobCollector.OnHTML(".job-meta-section", func(h *colly.HTMLElement) {
			// Step 0. ??????????????????
			salaryText := h.DOM.Find(".job-salary").Text()
			if len(salaryText) == 0 {
				recordData.MinSalary = 0
				recordData.MaxSalary = 0
				recordData.Currency = "TWD"
				return
			}

			// Step 1. ??????????????????
			textForTimeUnit := strings.Split(salaryText, "/")
			timeUnit := textForTimeUnit[1]

			// Step 2. ?????? Currency
			textForCurrency := strings.Split(textForTimeUnit[0], " ")
			recordData.Currency = textForCurrency[len(textForCurrency)-1]

			// Step 3. ?????? ?????? ??? ?????? ??????
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

		// Get Job Skills
		jobCollector.OnHTML("#job-description", func(h *colly.HTMLElement) {
			extractSkillsFromText(h.DOM.Text(), skillsList, skillsData)
		})

		jobCollector.OnHTML("#requirements", func(h *colly.HTMLElement) {
			extractSkillsFromText(h.DOM.Text(), skillsList, skillsData)
		})

		jobCollector.OnHTML(".labels .label", func(h *colly.HTMLElement) {
			extractSkillsFromText(h.DOM.Text(), skillsList, skillsData)
		})

		// Visit the Detail Page
		jobCollector.Visit(recordData.Link)

		csvWriter.Write([]string{
			recordData.Title,
			recordData.Company,
			recordData.Link,
			strconv.Itoa(recordData.MinSalary),
			strconv.Itoa(recordData.MaxSalary),
			recordData.Currency,
			skillsData["html"],
			skillsData["css"],
			skillsData["javascript"],
			skillsData["typescript"],
			skillsData["jquery"],
			skillsData["svelte"],
			skillsData["react"],
			skillsData["vue"],
			skillsData["angular"],
			skillsData["angularjs"],
			skillsData["git"],
			skillsData["unittest"],
			skillsData["webpack"],
		})
		csvWriter.Flush()
	})

	urlCollector.Limit(&colly.LimitRule{
		DomainGlob: "*cakeresume.*",
		Delay:      1 * time.Second,
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
