package wallet

// AccountID describes account identification - account address (or unique key).
type AccountID struct {
	ID       string `json:"id"`
	Currency string `json:"currency"`
}

// Account represents account state in the system.
type Account struct {
	ID      AccountID `json:"id"`
	Balance float64   `json:"balance"`
}

// BalanceAction describes one iteration of account balance modification.
type BalanceAction struct {
	Account AccountID `json:"account"`
	Volume  float64   `json:"volume"`
}

// Trans is a bussiness transaction, an atomic set of balance modifications for
// various accounts.
type Trans = []BalanceAction
