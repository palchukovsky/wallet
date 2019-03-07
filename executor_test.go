package wallet_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	w "github.com/palchukovsky/wallet"
	mw "github.com/palchukovsky/wallet/mock"
)

////////////////////////////////////////////////////////////////////////////////

func testExecutorExecutionRepoError(executor w.Executor, test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	trans := []w.BalanceAction{
		w.BalanceAction{
			Account: w.AccountID{ID: "qwerty", Currency: "USD"}, Volume: 1},
		w.BalanceAction{
			Account: w.AccountID{ID: "qwerty", Currency: "EUR"}, Volume: 2},
		w.BalanceAction{
			Account: w.AccountID{ID: "qwerty", Currency: "RUB"}, Volume: 3}}
	firstRequest := w.AccountID{ID: "qwerty", Currency: "USD"}

	repo := mw.NewMockAccountRepo(ctrl)
	repo.EXPECT().Modify(w.GetTransAccounts(trans), gomock.Any()).Do(
		func(
			_ map[w.AccountID]interface{},
			f func(repoTrans w.AccountRepoTrans) error) {

			secondRequest := w.AccountID{ID: "qwerty", Currency: "EUR"}
			// Second account retrieving attempt ends with a  predefined error.
			repoTrans := mw.NewMockAccountRepoTrans(ctrl)
			repoTrans.EXPECT().
				GetAccount(secondRequest).Return(nil, errors.New("Test error 1")).
				After(repoTrans.EXPECT().GetAccount(firstRequest).
					Return(&w.Account{ID: firstRequest, Balance: 100}, nil))
			err := f(repoTrans)
			if err == nil || err.Error() != "Test error 1" {
				test.Errorf(`Callback has returned wrong error: "%v".`, err)
			}
		}).Return(errors.New("Test error 2"))

	affected, err := executor.Execute(trans, repo)
	if err == nil || err.Error() != "Test error 2" {
		test.Errorf(`Error handling is wrong: "%v".`, err)
	}

	if len(affected) != 0 {
		test.Errorf(`Affected account list has to be empty: "%v".`, affected)
	}

}

////////////////////////////////////////////////////////////////////////////////

// Test_Executor_Manager_Execute_Success tests successful transaction execution
// by the manager.
func Test_Executor_Manager_Execute_Success(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	trans := []w.BalanceAction{
		w.BalanceAction{
			Account: w.AccountID{ID: "0", Currency: "USD"}, Volume: 1},
		w.BalanceAction{
			Account: w.AccountID{ID: "0", Currency: "EUR"}, Volume: -1},
		w.BalanceAction{
			Account: w.AccountID{ID: "-50", Currency: "RUB"}, Volume: 100}}

	repo := mw.NewMockAccountRepo(ctrl)
	repo.EXPECT().Modify(w.GetTransAccounts(trans), gomock.Any()).Do(
		func(
			_ map[w.AccountID]interface{},
			f func(repoTrans w.AccountRepoTrans) error) {

			repoTrans := mw.NewMockAccountRepoTrans(ctrl)
			for _, action := range trans {
				balance, err := strconv.ParseFloat(action.Account.ID, 64)
				if err != nil {
					test.Fatalf(`Test code has errors: "%s",`, err)
				}
				repoTrans.EXPECT().GetAccount(action.Account).Return(
					&w.Account{ID: action.Account, Balance: balance},
					nil)
			}
			f(repoTrans)
		})

	executor := w.CreateManagerExecutor()
	defer executor.Close()

	affected, err := executor.Execute(trans, repo)
	if err != nil {
		test.Fatalf(`Failed to execute: "%s".`, err)
	}
	if len(affected) != 3 {
		test.Fatalf(`Wrong affected list size: "%v".`, affected)
	}

	for _, action := range trans {
		ok := false
		for _, account := range affected {
			if action.Account != account.ID {
				continue
			}
			startBalance, err := strconv.ParseFloat(action.Account.ID, 64)
			if err != nil {
				test.Fatalf(`Test code has errors: "%s",`, err)
			}
			if account.Balance != startBalance+action.Volume {
				test.Errorf(`Wrong affected account: "%v".`, account)
			}
			ok = true
			break
		}
		if !ok {
			test.Errorf(`Affected has no required account: "%v".`, affected)
		}
	}

}

// Test_Executor_Manager_Execute_Error tests repository error handling while
// transaction execution by the manager.
func Test_Executor_Manager_Execute_RepoError(test *testing.T) {
	executor := w.CreateManagerExecutor()
	defer executor.Close()
	testExecutorExecutionRepoError(executor, test)
}

////////////////////////////////////////////////////////////////////////////////

// Test_Executor_Client_Execute_Success tests normal transaction execution by
// client.
func Test_Executor_Client_Execute_Success(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	trans := []w.BalanceAction{
		w.BalanceAction{
			Account: w.AccountID{ID: "0", Currency: "USD"}, Volume: 1},
		w.BalanceAction{
			Account: w.AccountID{ID: "1", Currency: "RUB"}, Volume: -1},
		w.BalanceAction{
			Account: w.AccountID{ID: "-1", Currency: "EUR"}, Volume: 1},
		w.BalanceAction{
			Account: w.AccountID{ID: "-2", Currency: "EUR"}, Volume: 1},
		w.BalanceAction{
			Account: w.AccountID{ID: "-1", Currency: "RUB"}, Volume: 2},
		w.BalanceAction{
			Account: w.AccountID{ID: "0", Currency: "RUB"}, Volume: 0}}

	repo := mw.NewMockAccountRepo(ctrl)
	repo.EXPECT().Modify(w.GetTransAccounts(trans), gomock.Any()).Do(
		func(
			_ map[w.AccountID]interface{},
			f func(repoTrans w.AccountRepoTrans) error) {

			repoTrans := mw.NewMockAccountRepoTrans(ctrl)
			for _, action := range trans {
				balance, err := strconv.ParseFloat(action.Account.ID, 64)
				if err != nil {
					test.Fatalf(`Test code has errors: "%s",`, err)
				}
				repoTrans.EXPECT().GetAccount(action.Account).Return(
					&w.Account{ID: action.Account, Balance: balance},
					nil)
			}
			f(repoTrans)
		})

	executor := w.CreateClientExecutor()
	defer executor.Close()

	affected, err := executor.Execute(trans, repo)
	if err != nil {
		test.Fatalf(`Failed to execute: "%s".`, err)
	}

	for _, account := range affected {
		ok := false
		for _, action := range trans {
			if action.Account != account.ID {
				continue
			}
			ok = true
			break
		}
		if !ok {
			test.Fatalf(`Affected account is unknown: "%v".`, affected)
		}
	}

	for _, action := range trans {
		ok := false
		for _, account := range affected {
			if action.Account != account.ID {
				continue
			}
			startBalance, err := strconv.ParseFloat(action.Account.ID, 64)
			if err != nil {
				test.Fatalf(`Test code has errors: "%s",`, err)
			}
			if account.Balance != startBalance+action.Volume {
				test.Errorf(`Wrong affected account: "%v".`, account)
			}
			ok = true
			break
		}
		if !ok {
			test.Errorf(`Account was not affected: "%v".`, action)
		}
	}

}

// Test_Executor_Client_Execute_DoesNotHaveEnoughFundsError tests "account does
// not have enough funds" error handling while transaction execution by client.
func Test_Executor_Client_Execute_DoesNotHaveEnoughFundsError(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	transList := [][]w.BalanceAction{
		{
			w.BalanceAction{
				Account: w.AccountID{ID: "1", Currency: "USD"}, Volume: 2},
			w.BalanceAction{
				Account: w.AccountID{ID: "1", Currency: "EUR"}, Volume: -2}},
		{
			w.BalanceAction{
				Account: w.AccountID{ID: "1", Currency: "USD"}, Volume: 2},
			w.BalanceAction{
				Account: w.AccountID{ID: "0", Currency: "EUR"}, Volume: -1}},
		{
			w.BalanceAction{
				Account: w.AccountID{ID: "1", Currency: "USD"}, Volume: 2},
			w.BalanceAction{
				Account: w.AccountID{ID: "-1", Currency: "EUR"}, Volume: -1}}}

	for _, trans := range transList {
		var errorAccount w.AccountID
		for _, action := range trans {
			errorAccount = action.Account
		}

		repo := mw.NewMockAccountRepo(ctrl)
		repo.EXPECT().Modify(w.GetTransAccounts(trans), gomock.Any()).DoAndReturn(
			func(
				_ map[w.AccountID]interface{},
				f func(repoTrans w.AccountRepoTrans) error) error {

				repoTrans := mw.NewMockAccountRepoTrans(ctrl)
				for _, action := range trans {
					balance, err := strconv.ParseFloat(action.Account.ID, 64)
					if err != nil {
						test.Fatalf(`Test code has errors: "%s",`, err)
					}
					repoTrans.EXPECT().GetAccount(action.Account).Return(
						&w.Account{ID: action.Account, Balance: balance},
						nil)
				}
				return f(repoTrans)
			})

		executor := w.CreateClientExecutor()
		defer executor.Close()

		affected, err := executor.Execute(trans, repo)
		if err == nil ||
			err.Error() != fmt.Sprintf(
				`Account "%s" (%s) does not have enough funds`,
				errorAccount.ID, errorAccount.Currency) {

			test.Errorf(`Error handling is wrong: "%v".`, err)
		}

		if len(affected) != 0 {
			test.Errorf(`Affected account list has to be empty: "%v".`, affected)
		}
	}
}

// Test_Executor_Client_Execute_Error tests repository error handling while
// transaction execution by client.
func Test_Executor_Client_Execute_RepoError(test *testing.T) {
	executor := w.CreateClientExecutor()
	defer executor.Close()
	testExecutorExecutionRepoError(executor, test)
}

////////////////////////////////////////////////////////////////////////////////
