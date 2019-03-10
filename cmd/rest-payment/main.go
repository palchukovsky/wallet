package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var (
	host        = flag.String("host", "localhost:80", "service host and port")
	fromAccount = flag.String("from_account", "", "source account")
	toAccount   = flag.String("to_account", "", "destinaction account")
	currency    = flag.String("currency", "USD", "accounts currency")
	amount      = flag.Float64("amount", .0, "transaction amount")
)

func main() {
	flag.Parse()

	req := url.URL{Scheme: "http", Host: *host, Path: "/payment"}
	resp, err := http.PostForm(
		req.String(), url.Values{
			"from_account": {*fromAccount},
			"to_account":   {*toAccount},
			"currency":     {*currency},
			"amount":       {fmt.Sprintf("%f", *amount)}})
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
