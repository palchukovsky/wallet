package wallet_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	w "github.com/palchukovsky/wallet"
	mw "github.com/palchukovsky/wallet/mock"
)

// Test_Service tests wallet service implementation.
func Test_Service(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	repo := mw.NewMockRepo(ctrl)
	clientExec := mw.NewMockExecutor(ctrl)
	managerExec := mw.NewMockExecutor(ctrl)

	service := w.CreateService(repo, clientExec, managerExec)
	defer service.Close()

	{
		account := w.AccountID{ID: "123", Currency: "asd"}
		repo.EXPECT().AddAccount(w.Account{ID: account, Balance: 0}).
			Return(errors.New("AddAccount error"))
		err := service.CreateAccount(account)
		if err == nil || err.Error() != "AddAccount error" {
			test.Errorf("Wrong result: %v.", err)
		}
	}
	{
		list := []w.Account{
			{ID: w.AccountID{ID: "123", Currency: "345"},
				Balance: 456.678},
			{ID: w.AccountID{ID: "678", Currency: "098"},
				Balance: 123.123}}
		repo.EXPECT().GetAccounts().Return(list, nil)
		result := service.GetAccounts()
		if len(list) != len(result) {
			test.Errorf("Wrong result: %v.", result)
		} else {
			for i, account := range result {
				if account != list[i] {
					test.Errorf("Wrong result: %v.", result)
				}
			}
		}
	}
	{
		repo.EXPECT().GetAccounts().Return(nil, errors.New("Test error"))
		result := service.GetAccounts()
		if len(result) != 0 {
			test.Errorf("Wrong result: %v.", result)
		}
	}
	{
		list := []w.Trans{
			{
				{Account: w.AccountID{ID: "123", Currency: "345"},
					Volume: 456.678},
				{Account: w.AccountID{ID: "678", Currency: "098"},
					Volume: 123.123}},
			{
				{Account: w.AccountID{ID: "1231", Currency: "3451"},
					Volume: 456.678},
				{Account: w.AccountID{ID: "6781", Currency: "0981"},
					Volume: 123.123}}}
		repo.EXPECT().GetTransList().Return(list, nil)
		result := service.GetPayments()
		if len(list) != len(result) {
			test.Errorf("Wrong result: %v.", result)
		} else {
			for i, trans := range result {
				for j, action := range trans {
					if action != list[i][j] {
						test.Errorf("Wrong result: %v.", result)
					}
				}
			}
		}
	}
	{
		repo.EXPECT().GetTransList().Return(nil, errors.New("Test error"))
		result := service.GetPayments()
		if len(result) != 0 {
			test.Errorf("Wrong result: %v.", result)
		}
	}
	{
		action := w.BalanceAction{
			Account: w.AccountID{ID: "123", Currency: "asd"},
			Volume:  123123.123}
		managerExec.EXPECT().Execute(w.Trans{action}, repo).
			Return(nil, errors.New("manager error"))
		err := service.SetupAccount(action)
		if err == nil || err.Error() != "manager error" {
			test.Errorf("Wrong result: %v.", err)
		}
	}
	{
		src := w.BalanceAction{
			Account: w.AccountID{ID: "123", Currency: "asd"},
			Volume:  123123.32}
		dst := w.BalanceAction{
			Account: w.AccountID{ID: "45345", Currency: "123123"},
			Volume:  4534}
		clientExec.EXPECT().Execute(w.Trans{src, dst}, repo).
			Return(nil, errors.New("client error"))
		err := service.MakePayment(src, dst)
		if err == nil || err.Error() != "client error" {
			test.Errorf("Wrong result: %v.", err)
		}
	}
}
