package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	urlUtility "github.com/ChengYiLin/cakeresume-colly/utility"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

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
	// pageURL := "https://www.cakeresume.com/companies/funpodium/jobs/web-developer-react-web-engineer-react"
	// pageURL := "https://www.cakeresume.com/companies/ailabs/jobs/sr-front-end-full-stack-engineer"
	pageURL := "https://www.cakeresume.com/companies/meettheone/jobs/growth-marketing-manager-2eecee"
	languageParameter := urlUtility.KeyValueParameter("locale", "en")
	visitURL := urlUtility.AppendQueryString(pageURL, languageParameter)

	// Create Colly Collector For CakeResume
	c := colly.NewCollector(
		colly.AllowedDomains("www.cakeresume.com"),
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("------------------\nVisiting %s\n------------------\n", r.URL.String())
	})

	c.OnHTML(".job-meta-section", func(h *colly.HTMLElement) {
		var timeUnit string
		var currency string
		var minSalary int
		var maxSalary int

		// Step 0. 取得頁面元素
		salaryText := h.DOM.Find(".job-salary").Text()
		if len(salaryText) == 0 {
			fmt.Println("No Salary Data")
			return
		}

		// Step 1. 取得時間單位
		textForTimeUnit := strings.Split(salaryText, "/")
		timeUnit = textForTimeUnit[1]

		// Step 2. 取得 Currency
		textForCurrency := strings.Split(textForTimeUnit[0], " ")
		currency = textForCurrency[len(textForCurrency)-1]

		// Step 3. 取得 最高 及 最低 月薪
		if strings.Contains(textForCurrency[0], "+") {
			minSalaryText := strings.ReplaceAll(textForCurrency[0], "+", "")

			minSalary = getSalaryFromText(minSalaryText, timeUnit)
			maxSalary = 0
		} else {
			minSalaryText := textForCurrency[0]
			maxSalaryText := textForCurrency[2]

			minSalary = getSalaryFromText(minSalaryText, timeUnit)
			maxSalary = getSalaryFromText(maxSalaryText, timeUnit)
		}

		fmt.Println("== Salary ==")
		fmt.Printf("Time Unit  : %s\n", timeUnit)
		fmt.Printf("Currency   : %s\n", currency)
		fmt.Printf("Min Salary : %d\n", minSalary)
		fmt.Printf("Max Salary : %d\n", maxSalary)
		fmt.Println("------------------")

		// 技能 tag
		h.DOM.Find(".labels .label").Each(func(i int, s *goquery.Selection) {
			// fmt.Println(s.Text())
		})
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("Request URL : ", r.Request.URL, "\nfailed with response : ", r, "\nError : ", e)
	})

	c.Visit(visitURL)
}
