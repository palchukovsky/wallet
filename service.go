package wallet

import "log"

// Service describes the interface to access to the wallets service to request
// data and process payments.
type Service interface {
	// Close closes the service and frees resources.
	Close()
	// GetPayments returns information about all known payments for all accounts.
	GetPayments() []Trans
	// GetAccounts returns information about all known accounts.
	GetAccounts() []Account
	// CreateAccount creates new account with zero balance.
	CreateAccount(AccountID) error
	// SetupAccount modifies account balance by the manager.
	SetupAccount(BalanceAction) error
	// MakePayment executes funds transfer between two accounts.
	MakePayment(src BalanceAction, dst BalanceAction) error
}

type service struct {
	repo            Repo
	clientExecutor  Executor
	managerExecutor Executor
}

// CreateService crates wallet service to access to the wallets service to
// request data and process payments.
func CreateService(
	repo Repo, clientExecutor Executor, managerExecutor Executor) Service {

	return &service{
		repo:            repo,
		clientExecutor:  clientExecutor,
		managerExecutor: managerExecutor}
}

func (s *service) Close() {}

func (s *service) GetPayments() []Trans {
	result, err := s.repo.GetTransList()
	if err != nil {
		log.Printf(`Failed to query transaction list: "%s".`, err)
		return []Trans{}
	}
	return result
}

func (s *service) GetAccounts() []Account {
	result, err := s.repo.GetAccounts()
	if err != nil {
		log.Printf(`Failed to query account list: "%s".`, err)
		return []Account{}
	}
	return result
}

func (s *service) CreateAccount(id AccountID) error {
	return s.repo.AddAccount(Account{ID: id, Balance: 0})
}

func (s *service) SetupAccount(action BalanceAction) error {
	_, err := s.managerExecutor.Execute(Trans{action}, s.repo)
	return err
}

func (s *service) MakePayment(src BalanceAction, dst BalanceAction) error {
	_, err := s.clientExecutor.Execute(Trans{src, dst}, s.repo)
	return err
}
