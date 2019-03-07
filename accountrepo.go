package wallet

// AccountRepoTrans represents accounts repository atomic modification.
type AccountRepoTrans interface {
	// GetAccount returns references to account state, or error if the account
	// record was not prefetched.
	GetAccount(id AccountID) (*Account, error)
	// StoreTrans stores information about bussiness transaction.
	StoreTrans(trans Trans) error
}

// AccountRepo describes accounts repository interface.
type AccountRepo interface {
	// Store stores (adds or update existing) account without conditions.
	Store(account Account) error
	// Modify takes accounts to prefetch data, then calls f with prefetched data
	// and applies changes by a transaction if f has not returned an error.
	Modify(
		accounts map[AccountID]interface{},
		f func(tans AccountRepoTrans) error) error
}
