package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/palchukovsky/wallet"
)

var (
	dbHost     = flag.String("db_host", "localhost", "database host")
	dbName     = flag.String("db_name", "wallet", "database name")
	dbLogin    = flag.String("db_login", "wallet", "database user login name")
	dbPassword = flag.String(
		"db_password", "WaLlEtSeCrEtPaSsWoRd4", "database user login password")
	port = flag.Uint("port", 80, "HTTP server port")
)

func main() {
	flag.Parse()

	db, err := wallet.CreateDB(*dbHost, *dbName, *dbLogin, *dbPassword)
	if err != nil {
		log.Panicf(`Failed to connect to the database: "%s".`, err)
	}
	defer db.Close()

	repo := wallet.CreateRepo(db)
	defer repo.Close()

	managerExec := wallet.CreateManagerExecutor()
	clientExec := wallet.CreateClientExecutor()

	service := wallet.CreateService(repo, clientExec, managerExec)
	defer service.Close()

	server := createServerOrExit(service, CreateProtocol(), *port)
	defer server.close()

	interruptChan := make(chan os.Signal, 1)
	defer close(interruptChan)
	signal.Notify(interruptChan, os.Interrupt)
	<-interruptChan
}
