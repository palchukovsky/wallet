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
	resp http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "POST":
		s.createAccount(resp, req)
	case "PUT":
		s.UpdateAccount(resp, req)
	case "GET":
		s.sendAccountList(resp, req)
	default:
		log.Printf(`Requested unknown methods "%s" for account.`, req.Method)
		resp.WriteHeader(http.StatusNotFound)
		resp.Write([]byte("Unknown method"))
	}
}

func (s *server) createAccount(resp http.ResponseWriter, req *http.Request) {
	log.Printf(`Creating new account...`)
	err := s.service.CreateAccount(wallet.AccountID{
		ID: req.FormValue("id"), Currency: req.FormValue("currency")})
	if err != nil {
		log.Printf(`Failed to create account: "%s". Request: %v.`, err, *req)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte("Failed to create account"))
		return
	}
	resp.WriteHeader(http.StatusCreated)
	log.Println(`New account created.`)
}

func (s *server) UpdateAccount(
	resp http.ResponseWriter, req *http.Request) {

	log.Println(`Updating account...`)
	action := wallet.BalanceAction{Account: wallet.AccountID{
		ID: req.FormValue("id"), Currency: req.FormValue("currency")}}
	var err error
	action.Volume, err = strconv.ParseFloat(req.FormValue("amount"), 64)
	if err != nil {
		log.Printf(`Failed to parse account setup amount: "%s". Request: %v.`,
			err, req.FormValue("amount"))
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte("Failed to parse account setup amount"))
		return
	}
	if err := s.service.SetupAccount(action); err != nil {
		log.Printf(`Failed to setup account: "%s". Request: %v.`, err, *req)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte("Failed to setup account"))
		return
	}
	resp.WriteHeader(http.StatusOK)
	log.Println(`Account updated.`)
}

func (s *server) sendAccountList(resp http.ResponseWriter, req *http.Request) {
	log.Println(`Account list requested...`)
	resp.WriteHeader(http.StatusOK)
	resp.Write(s.protocol.SerializeAccounts(s.service.GetAccounts()))
	resp.Header().Set("Content-Type", s.protocol.GetContentType())
}

func (s *server) handlePaymentRequest(
	resp http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "POST":
		s.processPayment(resp, req)
	case "GET":
		s.sendPaymentList(resp, req)
	default:
		log.Printf(`Requested unknown methods "%s" for payment.`, req.Method)
		resp.WriteHeader(http.StatusNotFound)
		resp.Write([]byte("Unknown method"))
	}
}

func (s *server) processPayment(resp http.ResponseWriter, req *http.Request) {
	log.Println(`Processing payment...`)
	currency := req.FormValue("currency")
	amount, err := strconv.ParseFloat(req.FormValue("amount"), 64)
	if err != nil || amount < 0 {
		log.Printf(`Failed to parse payment amount: "%s". Request: %v.`, err, *req)
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte("Failed to parse payment amount"))
		return
	}
	src := wallet.BalanceAction{Account: wallet.AccountID{
		ID: req.FormValue("from_account"), Currency: currency},
		Volume: -amount}
	dst := wallet.BalanceAction{Account: wallet.AccountID{
		ID: req.FormValue("to_account"), Currency: currency},
		Volume: amount}
	if err := s.service.MakePayment(src, dst); err != nil {
		log.Printf(`Failed to make payment: "%s". Request: %v.`, err, *req)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte("Failed to make payment"))
		return
	}
	resp.WriteHeader(http.StatusOK)
	log.Println(`Payment successfully processed.`)
}

func (s *server) sendPaymentList(resp http.ResponseWriter, req *http.Request) {
	log.Println(`Payment list requested...`)
	resp.WriteHeader(http.StatusOK)
	resp.Write(s.protocol.SerializeTransList(s.service.GetPayments()))
	resp.Header().Set("Content-Type", s.protocol.GetContentType())
}
