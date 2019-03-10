package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/palchukovsky/wallet"

	"github.com/gorilla/mux"
)

type server struct {
	service    wallet.Service
	protocol   Protocol
	server     *http.Server
	stopWaiter sync.WaitGroup
}

// createServerOrExit creates and start local server to handle REST-requests.
// To stop close must be called.
func createServerOrExit(
	service wallet.Service, protocol Protocol, port uint) *server {

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Panicf(`Failed to open server endpooint: "%s".`, err)
	}

	result := &server{service: service, protocol: protocol}

	router := mux.NewRouter()
	router.StrictSlash(true)
	router.HandleFunc("/account", result.handleAccountRequest)
	router.HandleFunc("/payment", result.handlePaymentRequest)

	result.server = &http.Server{Handler: router}

	go func() {
		result.stopWaiter.Add(1)
		defer result.stopWaiter.Done()
		defer listener.Close()
		result.server.Serve(listener)
	}()

	return result
}

// close stops the server and frees resources.
func (s *server) close() {
	s.server.Close()
	s.stopWaiter.Wait()
}

func (s *server) handleAccountRequest(
	response http.ResponseWriter,
	request *http.Request) {

	switch request.Method {
	case "POST":
		log.Printf(`Creating new account...`)
		err := s.service.CreateAccount(wallet.AccountID{
			ID: request.FormValue("id"), Currency: request.FormValue("currency")})
		if err != nil {
			log.Printf(`Failed to create account: "%s". Request: %v.`, err, *request)
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte("Failed to create account"))
			break
		}
		response.WriteHeader(http.StatusCreated)
		log.Println(`New account created.`)

	case "PUT":
		log.Println(`Updating account...`)
		action := wallet.BalanceAction{Account: wallet.AccountID{
			ID: request.FormValue("id"), Currency: request.FormValue("currency")}}
		var err error
		action.Volume, err = strconv.ParseFloat(request.FormValue("amount"), 64)
		if err != nil {
			log.Printf(`Failed to parse account setup amount: "%s". Request: %v.`,
				err, request.FormValue("amount"))
			response.WriteHeader(http.StatusBadRequest)
			response.Write([]byte("Failed to parse account setup amount"))
			break
		}
		if err := s.service.SetupAccount(action); err != nil {
			log.Printf(`Failed to setup account: "%s". Request: %v.`, err, *request)
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte("Failed to setup account"))
			break
		}
		response.WriteHeader(http.StatusOK)
		log.Println(`Account updated.`)

	case "GET":
		log.Println(`Account list requested...`)
		response.Write(s.protocol.SerializeAccounts(s.service.GetAccounts()))
		response.Header().Set("Content-Type", s.protocol.GetContentType())
		response.WriteHeader(http.StatusOK)

	default:
		log.Printf(`Requested unknown methods "%s" for account.`, request.Method)
		response.WriteHeader(http.StatusNotFound)
		response.Write([]byte("Unknown method"))

	}
}

func (s *server) handlePaymentRequest(
	response http.ResponseWriter,
	request *http.Request) {

	switch request.Method {
	case "POST":
		log.Println(`Processing payment...`)
		currency := request.FormValue("currency")
		amount, err := strconv.ParseFloat(request.FormValue("amount"), 64)
		if err != nil || amount < 0 {
			log.Printf(`Failed to parse payment amount: "%s". Request: %v.`,
				err, *request)
			response.WriteHeader(http.StatusBadRequest)
			response.Write([]byte("Failed to parse payment amount"))
			break
		}
		src := wallet.BalanceAction{Account: wallet.AccountID{
			ID: request.FormValue("from_account"), Currency: currency},
			Volume: -amount}
		dst := wallet.BalanceAction{Account: wallet.AccountID{
			ID: request.FormValue("to_account"), Currency: currency},
			Volume: amount}
		if err := s.service.MakePayment(src, dst); err != nil {
			log.Printf(`Failed to make payment: "%s". Request: %v.`, err, *request)
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte("Failed to make payment"))
			break
		}
		response.WriteHeader(http.StatusOK)
		log.Println(`Payment successfully processed.`)

	case "GET":
		log.Println(`Payment list requested...`)
		response.Write(s.protocol.SerializeTransList(s.service.GetPayments()))
		response.Header().Set("Content-Type", s.protocol.GetContentType())
		response.WriteHeader(http.StatusOK)

	default:
		log.Printf(`Requested unknown methods "%s" for payment.`, request.Method)
		response.WriteHeader(http.StatusNotFound)
		response.Write([]byte("Unknown method"))

	}
}
