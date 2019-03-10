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
