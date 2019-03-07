package wallet_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	w "github.com/palchukovsky/wallet"
)

// Test_Trans_AccountList tests GetTransAccounts function.
func Test_GetTransAccounts(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	trans := []w.BalanceAction{}
	addTrans := func(id, currency string, vol float64) {
		trans = append(
			trans,
			w.BalanceAction{
				Account: w.AccountID{ID: id, Currency: currency},
				Volume:  vol})
	}
	addTrans("123", "USD", 123123)    // +1
	addTrans("123", "EUR", 24532345)  // +1
	addTrans("123", "USD", 53254)     // +0
	addTrans("abc", "USD", 6563)      // +1
	addTrans("abc", "USD", 73234)     // +0
	addTrans("qwerty1", "USD", 4727)  // +1
	addTrans("qwerty2", "USD", 34434) // +1

	list := w.GetTransAccounts(trans)
	scores := 0
	if len(list) == 5 {
		scores++
	}
	if _, ok := list[w.AccountID{ID: "123", Currency: "USD"}]; ok {
		scores++
	}
	if _, ok := list[w.AccountID{ID: "123", Currency: "EUR"}]; ok {
		scores++
	}
	if _, ok := list[w.AccountID{ID: "abc", Currency: "USD"}]; ok {
		scores++
	}
	if _, ok := list[w.AccountID{ID: "qwerty1", Currency: "USD"}]; ok {
		scores++
	}
	if _, ok := list[w.AccountID{ID: "qwerty2", Currency: "USD"}]; ok {
		scores++
	}
	if scores != 6 {
		test.Errorf("List is wrong: %v", list)
	}
}
