package wallet

// AccountID describes account identification - account address (or unique key).
type AccountID struct {
	ID       string
	Currency string
}

// Account represents account state in the system.
type Account struct {
	ID      AccountID
	Balance float64
}

// BalanceAction describes one iteration of account balance modification.
type BalanceAction struct {
	Account AccountID
	Volume  float64
}

// Trans is a bussiness transaction, an atomic set of balance modifications for
// various accounts.
type Trans = []BalanceAction
