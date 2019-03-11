# Wallet demo
Abstract wallet service example with REST-access. It covers the next stories:

- I want to be able to send a payment from one account to another (same currency) with account balance control
- I want to be able to see all payments
- I want to be able to see available accounts and their balances
- I want to add a new account
- I want to modify account balance as the manager, without an account balance control
- I want to communicate with the service by REST

## REST API

REST API described in [docs/api.md](https://github.com/palchukovsky/wallet/blob/master/docs/api.md).

## Components

### cmd/rest-server
REST-server, accepts and executes all service commands. To get command line arguments see the result of the command `rest-server -?`.

### cmd/rest-addaccount
REST-client example to add new accounts. To get command line arguments see the result of the command `rest-addaccount -?`

### cmd/rest-setbalance
REST-client example to set account balance as the manager. To get command line arguments see the result of the command `rest-setbalance -?`

### cmd/rest-info
REST-client example to get the account and payments list. To get command line arguments see the result of the command `rest-info -?`

### cmd/rest-payment
REST-client example to make payments. To get command line arguments see the result of the command `rest-payment -?`

## Install from source 

You can directly use the `go` tool to download and install the service sources:

    go get github.com/palchukovsky/wallet
    cd ${GOPATH}github.com/palchukovsky/wallet/cmd/rest-server
    go build
    
To start REST-server with default parameters use:

    rest-server
    
Information about additional parameters available by the command:

    rest-server -?

To initialize PostgreSQL database use SQL-script build/db/init.sql.

To build Docker image from the source use the command:

    make build
    
To run unit-tests run commands:

    make mock
    go test
    
### Install from [Docker Hub](https://hub.docker.com/r/palchukovsky/wallet.rest)

1. Get [docker-compose file](https://github.com/palchukovsky/wallet/blob/master/docker-compose.yml) for the service.
2. Edit docker-compose.yml to change database password in two places: db-service envelopment variable "POSTGRES_PASSWORD" and reset-service argument "-db_password" (the same for the database name, login, and ports if you require it).
3. Start service by the command `docker-compose up -d`.
4. Test service online status by example REST-client, for example by the commad `rest-info` (by default it uses `localhost:80` as service host).



