package wallet

import (
	"fmt"
	"sort"
	"time"
)

// RepoTrans represents repository atomic modification.
type RepoTrans interface {
	// GetAccount returns references to account state, or error if the account
	// record was not prefetched.
	GetAccount(id AccountID) (*Account, error)
}

// Repo describes repository interface.
type Repo interface {
	// Close closes repository and frees resources.
	Close()
	// AddAccount adds new account.
	AddAccount(account Account) error
	// Modify takes bussiness transaction to prefetch data, then calls f with
	// prefetched data and applies changes by a transaction if f has not
	// returned an error.
	Modify(trans Trans, author string, f func(tans RepoTrans) error) error
	// GetAccounts returns full account list.
	GetAccounts() ([]Account, error)
	// GetTransList returns full transaction list.
	GetTransList() ([]Trans, error)
}

////////////////////////////////////////////////////////////////////////////////

type repoTrans struct {
	db       DBTrans
	accounts map[AccountID]struct {
		account *Account
		pk      int
	}
}

func createRepoTrans(db DB) (*repoTrans, error) {
	result := &repoTrans{}
	var err error
	result.db, err = db.Begin()
	if err != nil {
		return nil, err
	}
	return result, nil
}

type accountIDList []AccountID

func (a accountIDList) Len() int      { return len(a) }
func (a accountIDList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a accountIDList) Less(i, j int) bool {
	l := a[i]
	r := a[j]
	return l.Currency < r.Currency || (l.Currency == r.Currency && l.ID < r.ID)
}

func (t *repoTrans) load(trans Trans) error {
	accountIDs := make(accountIDList, 0)
	// To avoid database deadlock while locking records it has to take records for
	// an update in the same order as any another request, so it has to be sorted.
	{
		keys := map[AccountID]interface{}{}
		for _, action := range trans {
			if _, has := keys[action.Account]; !has {
				accountIDs = append(accountIDs, action.Account)
				keys[action.Account] = nil
			}
		}
		sort.Sort(accountIDs)
	}

	accounts := map[AccountID]struct {
		account *Account
		pk      int
	}{}
	for _, id := range accountIDs {
		account, pk, err := t.db.QueryAccount(id, true)
		if err != nil {
			return err
		}
		accounts[account.ID] = struct {
			account *Account
			pk      int
		}{account: account, pk: *pk}
	}

	t.accounts = accounts
	return nil
}

func (t *repoTrans) GetAccount(id AccountID) (*Account, error) {
	if result, ok := t.accounts[id]; ok {
		return result.account, nil
	}
	return nil,
		fmt.Errorf(`Account "%s" (%s) was not prefetched`, id.ID, id.Currency)
}

func (t *repoTrans) commit() error {
	for _, account := range t.accounts {
		if err := t.db.UpdateAccount(*account.account, account.pk); err != nil {
			return err
		}
	}
	return t.db.Commit()
}

func (t *repoTrans) rollback() { t.db.Rollback() }

func (t *repoTrans) storeTrans(trans Trans, author string) error {
	transPk, err := t.db.InsertTrans(time.Now().UTC(), author)
	if err != nil {
		return err
	}
	for _, action := range trans {
		account, ok := t.accounts[action.Account]
		if !ok {
			return fmt.Errorf(`Transaction account "%s" (%s) was not prefetched`,
				action.Account.ID, action.Account.Currency)
		}
		err = t.db.InsertAction(account.pk, *transPk, action.Volume)
		if err != nil {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type repo struct{ db DB }

// CreateRepo creates repository implementation instance.
func CreateRepo(db DB) Repo { return &repo{db: db} }

func (r *repo) Close() {}

func (r *repo) AddAccount(account Account) error {
	trans, err := createRepoTrans(r.db)
	if err != nil {
		return err
	}
	defer trans.rollback()
	if err = trans.db.AddAccount(account); err != nil {
		return err
	}
	return trans.commit()
}

func (r *repo) Modify(
	trans Trans, author string, f func(tans RepoTrans) error) error {

	dbTrans, err := createRepoTrans(r.db)
	if err != nil {
		return err
	}
	defer dbTrans.rollback()
	if err = dbTrans.load(trans); err != nil {
		return err
	}
	if err = f(dbTrans); err != nil {
		return err
	}
	if err = dbTrans.storeTrans(trans, author); err != nil {
		return err
	}
	return dbTrans.commit()
}

func (r *repo) GetAccounts() ([]Account, error) { return r.db.GetAccounts() }

func (r *repo) GetTransList() ([]Trans, error) { return r.db.GetTransList() }

////////////////////////////////////////////////////////////////////////////////
