package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var (
	host     = flag.String("url", "localhost:80", "service host and port")
	id       = flag.String("id", "", "account ID")
	currency = flag.String("currency", "USD", "account currency")
	amount   = flag.Float64("amount", .0, "transaction amount")
)

func main() {
	flag.Parse()

	reqBody := url.Values{
		"id":       {*id},
		"currency": {*currency},
		"amount":   {fmt.Sprintf("%f", *amount)}}

	reqURL := url.URL{Scheme: "http", Host: *host, Path: "/account"}
	req, err := http.NewRequest(
		"PUT", reqURL.String(), strings.NewReader(reqBody.Encode()))
	if err != nil {
		log.Panicf(`Failed to request: "%s".`, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
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
