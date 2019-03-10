package main_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	w "github.com/palchukovsky/wallet"
	rs "github.com/palchukovsky/wallet/cmd/rest-server"
)

// Test_Protocol tests JSON protocol implementation.
func Test_Protocol(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	protocol := rs.CreateProtocol()

	if protocol.GetContentType() != "application/json; charset=utf-8" {
		test.Error("Wrong content type.")
	}

	{
		source := []w.Trans{
			{
				w.BalanceAction{
					Account: w.AccountID{ID: "accId1", Currency: "currencyCode1"},
					Volume:  123.123},
				w.BalanceAction{
					Account: w.AccountID{ID: "accId2", Currency: "currencyCode2"},
					Volume:  234.234}},
			{
				w.BalanceAction{
					Account: w.AccountID{ID: "accId3", Currency: "currencyCode3"},
					Volume:  567.567},
				w.BalanceAction{
					Account: w.AccountID{ID: "accId4", Currency: "currencyCode4"},
					Volume:  890.89}}}
		result := protocol.SerializeTransList(source)
		template := `[[{"account":{"id":"accId1","currency":"currencyCode1"},"volume":123.123},{"account":{"id":"accId2","currency":"currencyCode2"},"volume":234.234}],[{"account":{"id":"accId3","currency":"currencyCode3"},"volume":567.567},{"account":{"id":"accId4","currency":"currencyCode4"},"volume":890.89}]]`
		if template != string(result) {
			test.Errorf("Wrong JSON: %s.", string(result))
		}
	}

	{
		source := []w.Account{
			{
				ID: w.AccountID{ID: "accId1", Currency: "currencyCode1"},
				Balance: 123.123},
			{
				ID: w.AccountID{ID: "accId2", Currency: "currencyCode2"},
				Balance: 2222.3333}}
		result := protocol.SerializeAccounts(source)
		template := `[{"id":{"id":"accId1","currency":"currencyCode1"},"balance":123.123},{"id":{"id":"accId2","currency":"currencyCode2"},"balance":2222.3333}]`
		if template != string(result) {
			test.Errorf("Wrong JSON: %s.", string(result))
		}
	}
}
