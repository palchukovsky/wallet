package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"

	"github.com/palchukovsky/wallet"
)

var (
	host     = flag.String("host", "localhost:80", "service host and port")
	id       = flag.String("id", "", "new account ID")
	currency = flag.String("currency", "USD", "new account currency")
)

func printAccounts() {
	req := url.URL{Scheme: "http", Host: *host, Path: "/account"}
	resp, err := http.Get(req.String())
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

	list := []wallet.Account{}
	err = json.Unmarshal([]byte(body), &list)
	if err != nil {
		log.Fatalf(`Failed to parse server response: "%s".`, err)
	}

	log.Println("==========================================================================")
	log.Printf("ACCOUNTS (%d):", len(list))
	log.Println("")
	for _, account := range list {
		log.Printf("id: %s", account.ID.ID)
		log.Printf("balance: %f", account.Balance)
		log.Printf("currency: %s", account.ID.Currency)
		log.Println("")
	}
}

func printPayments() {
	req := url.URL{Scheme: "http", Host: *host, Path: "/payment"}
	resp, err := http.Get(req.String())
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

	list := []wallet.Trans{}
	err = json.Unmarshal([]byte(body), &list)
	if err != nil {
		log.Fatalf(`Failed to parse server response: "%s".`, err)
	}

	log.Println("==========================================================================")
	log.Printf("Transactions (%d):", len(list))
	log.Println("")
	for i, trans := range list {
		log.Printf("transaction %d:", i+1)
		if len(trans) != 2 ||
			(trans[0].Volume < 0) == (trans[1].Volume < 0) ||
			(trans[0].Volume == 0 || trans[1].Volume == 0) {

			for _, action := range trans {
				log.Printf("account: %s (%s)", action.Account.ID, action.Account.Currency)
				log.Printf("amount: %f", math.Abs(action.Volume))
				var direction string
				if action.Volume == 0 {
					direction = "none"
				} else if action.Volume < 0 {
					direction = "outgoing"
				} else {
					direction = "incoming"
				}
				log.Printf("direction: %s", direction)
			}
		} else {
			var outgoing wallet.BalanceAction
			var incoming wallet.BalanceAction
			if trans[0].Volume < 0 {
				incoming = trans[1]
				outgoing = trans[0]
			} else {
				incoming = trans[0]
				outgoing = trans[1]
			}
			log.Printf("account: %s (%s)",
				outgoing.Account.ID, outgoing.Account.Currency)
			log.Printf("amount: %f", -outgoing.Volume)
			log.Printf("to_account: %s (%s)",
				incoming.Account.ID, incoming.Account.Currency)
			log.Println("direction: outgoing")
			log.Println("")
			log.Printf("account: %s (%s)",
				incoming.Account.ID, incoming.Account.Currency)
			log.Printf("amount: %f", incoming.Volume)
			log.Printf("from_account: %s (%s)",
				outgoing.Account.ID, outgoing.Account.Currency)
			log.Println("direction: incoming")
		}
		log.Println("------------------------------------")
	}
}

func main() {
	flag.Parse()
	printAccounts()
	printPayments()
}
