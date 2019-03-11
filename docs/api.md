# Wallet REST API

## Accounts

### Request
| Path | Method | Describtion | Arguments | Example |
------ | ------------|-----------|--------|---------|
|/account|POST|Add (create) new account with zero balance.|**id** (string): new account ID (name); **currency** (string): new account currency|[cmd/rest-addaccount](https://github.com/palchukovsky/wallet/blob/master/cmd/rest-addaccount/main.go)|
|/account|PUT|Update account balance without account final balance control.|**id** (string): existing account ID; **currency** (string): existing account currency; **amount** (float) amount of applying difference (ex.: "100" to increase account balance, "-100" - to decrease account balance)|[cmd/rest-setbalance](https://github.com/palchukovsky/wallet/blob/master/cmd/rest-setbalance/main.go)|
|/account|GET|Get the account list. Returns list of all accounts with their balances as a JSON string in response.||[cmd/rest-info](https://github.com/palchukovsky/wallet/blob/master/cmd/rest-info/main.go)|

### Account list request response
Account list request response is a JSON-formatted list of all accounts with their balances. Response format:

    [
      {
        "id": {
          "id": string with account ID (account name),
          "currency": string with account currency
        },
        "balance": float value account balance
      },
    ...
    ]
    
## Payments

### Request
| Path | Method | Describtion | Arguments | Example |
------ | ------------|-----------|--------|---------|
|/payment|POST|Make a payment.|**from_account** (string): existing source-account ID; **to_account** (string): existing destination-account ID; **currency** (string): payment currency; **amount** (float): payment amount;|[cmd/rest-payment](https://github.com/palchukovsky/wallet/blob/master/cmd/rest-payment/main.go)|
|/payment|GET|Get the payment list. Returns list of transactions as a JSON string in response.||[cmd/rest-info](https://github.com/palchukovsky/wallet/blob/master/cmd/rest-info/main.go)|

### Payment list request response
Payment list request response is a JSON-formatted list of all transactions. One transaction includes one or more actions, each action describes changes for one account. Response format:

    [
      [
        {
          "account": {
            "id": string with account ID (account name),
            "currency": string with account currency
          },
          "volume": float value of applying difference (ex.: "100" increases account balance, "-100" - decreases account balance)
        },
        ...
      ],
      ...
    ]
