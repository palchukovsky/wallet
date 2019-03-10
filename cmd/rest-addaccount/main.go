package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var (
	host     = flag.String("host", "localhost:80", "service host and port")
	id       = flag.String("id", "", "new account ID")
	currency = flag.String("currency", "USD", "new account currency")
)

func main() {
	flag.Parse()

	req := url.URL{Scheme: "http", Host: *host, Path: "/account"}
	resp, err := http.PostForm(
		req.String(), url.Values{"id": {*id}, "currency": {*currency}})
	if err != nil {
		log.Panicf(`Failed to request: "%s".`, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicf(`Failed to read response: "%s".`, err)
	}

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {

		log.Printf(`Server has returned error: "%s" (code %d).`,
			body, resp.StatusCode)
		return
	}

	if len(body) == 0 {
		log.Printf(`OK (code %d).`, resp.StatusCode)
	} else {
		log.Printf(`%s (code %d).`, body, resp.StatusCode)
	}

}
