package wallet

import "fmt"

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
func CreateClientExecutor() Executor { return &clientExecutor{} }

func (c *clientExecutor) Close() {}

func (c *clientExecutor) Execute(
	trans Trans, repo Repo) ([]Account, error) {

	result := []Account{}
	err := repo.Modify(
		trans,
		"client",
		func(repoTrans RepoTrans) error {
			for _, action := range trans {
				account, err := repoTrans.GetAccount(action.Account)
				if err != nil {
					return err
				}
				account.Balance += action.Volume
				// The policy does not allow to decrease account balance to the negative
				// value but allows to add funds, even if the final result still is
				// negative.
				if account.Balance < 0 && action.Volume < 0 {
					return fmt.Errorf(`Account "%s" (%s) does not have enough funds`,
						account.ID.ID, action.Account.Currency)
				}
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

type managerExecutor struct{}

// CreateManagerExecutor creates executor with policy for manager. The manager
// policy allows to set any account balance without limitation.
func CreateManagerExecutor() Executor { return &managerExecutor{} }

func (c *managerExecutor) Close() {}

func (c *managerExecutor) Execute(
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
