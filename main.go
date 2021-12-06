package main

import (
	"fmt"

	urlutility "github.com/ChengYiLin/cakeresume-colly/utility"
)

func main() {
	pageURL := "https://www.cakeresume.com/companies/StarkTech/jobs/front-end-engineer-ad1aa5?locale=en"
	languageParameter := urlutility.KeyValueParameter("locale", "en")
	visitURL := urlutility.AppendQueryString(pageURL, languageParameter)

	fmt.Println(visitURL)
}
