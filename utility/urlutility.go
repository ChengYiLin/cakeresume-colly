package urlutility

import (
	"log"
	"net/url"
)

type keyValueParameter struct {
	key   string
	value string
}

func AppendQueryString(pageURL string, queryParameters ...keyValueParameter) string {
	visitURL, err := url.Parse(pageURL)
	if err != nil {
		log.Fatal(err)
	}

	originalQueryString := visitURL.Query()

	for _, queryParameter := range queryParameters {
		originalQueryString.Add(queryParameter.key, queryParameter.value)
	}

	visitURL.RawQuery = originalQueryString.Encode()

	return visitURL.String()
}
