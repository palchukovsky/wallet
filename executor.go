package wallet

import (
	"fmt"
)

// Executor executes account modifications by implementation rules.
type Executor interface {
	// Close closes executor and frees resources.
	Close()
	// Execute accepts and executes a business transaction. Returns the actual
	// state of accounts that were affected at success.
	Execute(trans Trans, repo Repo) ([]Account, error)
}

////////////////////////////////////////////////////////////////////////////////

type clientExecutor struct{}

// CreateClientExecutor creates executor with policy for normal client. The
// client policy does not allow to decrease account balance to the negative
// value but allows to add funds, even if the final result still is negative.
// It also does not allow to execute a transaction with various currencies in
// the action list - currency must be only one.
func CreateClientExecutor() Executor { return &clientExecutor{} }

func (e clientExecutor) Close() {}

func (e *clientExecutor) Execute(trans Trans, repo Repo) ([]Account, error) {
	var result []Account
	return result,
		repo.Modify(trans, "client",
			func(repoTrans RepoTrans) error {
				var err error
				result, err = e.execTrans(trans, repoTrans)
				return err
			})
}

func (*clientExecutor) execTrans(
	transData Trans, repoTrans RepoTrans) ([]Account, error) {

	result := []Account{}
	var currency *string
	for _, action := range transData {

		if currency == nil {
			currency = &action.Account.Currency
		} else if *currency != action.Account.Currency {
			return nil,
				fmt.Errorf(`Account "%s" (%s) has a different currency from "%s"`,
					action.Account.ID, action.Account.Currency, *currency)
		}

		account, err := repoTrans.GetAccount(action.Account)
		if err != nil {
			return nil, err
		}

		account.Balance += action.Volume

		// The policy does not allow to decrease account balance to the negative
		// value but allows to add funds, even if the final result still is
		// negative.
		if account.Balance < 0 && action.Volume < 0 {
			return nil,
				fmt.Errorf(`Account "%s" (%s) does not have enough funds`,
					account.ID.ID, action.Account.Currency)
		}

		result = append(result, *account)
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

type managerExecutor struct{}

// CreateManagerExecutor creates executor with policy for manager. The manager
// policy allows to set any account balance without limitation.
func CreateManagerExecutor() Executor { return &managerExecutor{} }

func (e managerExecutor) Close() {}

func (e *managerExecutor) Execute(
	trans Trans, repo Repo) ([]Account, error) {

	result := []Account{}
	err := repo.Modify(
		trans,
		"manager",
		func(repoTrans RepoTrans) error {
			for _, action := range trans {
				account, err := repoTrans.GetAccount(action.Account)
				if err != nil {
					return err
				}
				account.Balance += action.Volume
				result = append(result, *account)
			}
			return nil
		})
	if err != nil {
		return []Account{}, err
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
