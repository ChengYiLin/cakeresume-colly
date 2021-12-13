package main

import (
	"fmt"
	"strings"

	urlutility "github.com/ChengYiLin/cakeresume-colly/utility"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func main() {
	pageURL := "https://www.cakeresume.com/companies/pinkoi/jobs/frontend-engineer-d15130"
	languageParameter := urlutility.KeyValueParameter("locale", "en")
	visitURL := urlutility.AppendQueryString(pageURL, languageParameter)

	// Create Colly Collector For CakeResume
	c := colly.NewCollector(
		colly.AllowedDomains("www.cakeresume.com"),
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("------------------\nVisiting %s\n------------------\n", r.URL.String())
	})

	c.OnHTML(".job-meta-section", func(h *colly.HTMLElement) {
		// 薪資範圍
		salaryText := strings.Split(h.DOM.Find(".job-salary").Text(), "/")

		salaryValue := salaryText[0]
		timeUnit := salaryText[1]

		fmt.Println(salaryValue)
		fmt.Println(timeUnit)
		fmt.Println("------------------")

		// 技能 tag
		h.DOM.Find(".labels .label").Each(func(i int, s *goquery.Selection) {
			fmt.Println(s.Text())
		})
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("Request URL : ", r.Request.URL, "\nfailed with response : ", r, "\nError : ", e)
	})

	c.Visit(visitURL)
}
