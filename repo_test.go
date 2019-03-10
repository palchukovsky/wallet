package wallet_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	w "github.com/palchukovsky/wallet"
	mw "github.com/palchukovsky/wallet/mock"
)

// Test_Repo_AddAccount tests new account adding.
func Test_Repo_AddAccount(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	account := w.Account{
		ID:      w.AccountID{ID: "qwerty", Currency: "123456"},
		Balance: 12345.6789}
	{
		errText := "Test error"
		db := mw.NewMockDB(ctrl)
		db.EXPECT().Begin().Return(nil, errors.New(errText))
		repo := w.CreateRepo(db)
		err := repo.AddAccount(account)
		if err == nil || err.Error() != errText {
			test.Errorf(`Wrong error status: "%v".`, err)
		}
	}
	{
		errText := "Test error"
		trans := mw.NewMockDBTrans(ctrl)
		db := mw.NewMockDB(ctrl)
		trans.EXPECT().Rollback().
			After(trans.EXPECT().AddAccount(account).Return(errors.New(errText)).
				After(db.EXPECT().Begin().Return(trans, nil)))
		repo := w.CreateRepo(db)
		err := repo.AddAccount(account)
		if err == nil || err.Error() != errText {
			test.Errorf(`Wrong error status: "%v".`, err)
		}
	}
	{
		errText := "Test error"
		trans := mw.NewMockDBTrans(ctrl)
		db := mw.NewMockDB(ctrl)
		trans.EXPECT().Rollback().
			After(trans.EXPECT().Commit().Return(errors.New(errText)).
				After(trans.EXPECT().AddAccount(account).Return(nil).
					After(db.EXPECT().Begin().Return(trans, nil))))
		repo := w.CreateRepo(db)
		err := repo.AddAccount(account)
		if err == nil || err.Error() != errText {
			test.Errorf(`Wrong error status: "%v".`, err)
		}
	}
	{
		trans := mw.NewMockDBTrans(ctrl)
		db := mw.NewMockDB(ctrl)
		trans.EXPECT().Rollback().
			After(trans.EXPECT().Commit().Return(nil).
				After(trans.EXPECT().AddAccount(account).Return(nil).
					After(db.EXPECT().Begin().Return(trans, nil))))
		repo := w.CreateRepo(db)
		err := repo.AddAccount(account)
		if err != nil {
			test.Errorf(`Failed to store account: "%s".`, err)
		}
	}
}

// Test_Repo_Modify_Success tests atomic repository successfull modification.
func Test_Repo_Modify_Success(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	transData := []w.BalanceAction{
		w.BalanceAction{Account: w.AccountID{ID: "bbb", Currency: "AAA"}, Volume: 2},
		w.BalanceAction{Account: w.AccountID{ID: "aaa", Currency: "AAA"}, Volume: 1112},
		w.BalanceAction{Account: w.AccountID{ID: "bbb", Currency: "AAA"}, Volume: -2},
		w.BalanceAction{Account: w.AccountID{ID: "bbb", Currency: "AAA"}, Volume: 45},
		w.BalanceAction{Account: w.AccountID{ID: "aaa", Currency: "AAA"}, Volume: -23},
		w.BalanceAction{Account: w.AccountID{ID: "bbb", Currency: "BBB"}, Volume: -2},
		w.BalanceAction{Account: w.AccountID{ID: "bbb", Currency: "AAA"}, Volume: 2342},
		w.BalanceAction{Account: w.AccountID{ID: "aaa", Currency: "BBB"}, Volume: 1112}}
	pk1 := 1
	pk2 := 2
	pk3 := 3
	pk4 := 4

	hasCommit := false
	checkCommit := func(...interface{}) {
		if hasCommit {
			test.Error("Update after commit.")
		}
	}

	db := mw.NewMockDB(ctrl)
	dbTrans := mw.NewMockDBTrans(ctrl)
	dbTrans.EXPECT().Rollback().
		After(dbTrans.EXPECT().Commit().Return(nil).Do(func(...interface{}) { hasCommit = true }).
			After(dbTrans.EXPECT().QueryAccount(w.AccountID{ID: "bbb", Currency: "BBB"}, true).
				Return(&w.Account{ID: w.AccountID{ID: "bbb", Currency: "BBB"}, Balance: 4}, &pk4, nil).
				Do(func(...interface{}) {
					dbTrans.EXPECT().UpdateAccount(w.Account{ID: w.AccountID{ID: "bbb", Currency: "BBB"}, Balance: 400}, 4).Return(nil).Do(checkCommit)
					dbTrans.EXPECT().UpdateAccount(w.Account{ID: w.AccountID{ID: "aaa", Currency: "BBB"}, Balance: 3}, 3).Return(nil).Do(checkCommit)
					dbTrans.EXPECT().UpdateAccount(w.Account{ID: w.AccountID{ID: "bbb", Currency: "AAA"}, Balance: 200}, 2).Return(nil).Do(checkCommit)
					dbTrans.EXPECT().UpdateAccount(w.Account{ID: w.AccountID{ID: "aaa", Currency: "AAA"}, Balance: 1}, 1).Return(nil).Do(checkCommit)
					transPk := 99
					dbTrans.EXPECT().InsertTrans(gomock.Any(), "tester").Return(&transPk, nil).Do(checkCommit)
					dbTrans.EXPECT().InsertAction(gomock.Any(), transPk, gomock.Any()).Times(len(transData)).Return(nil).Do(checkCommit)
				}).
				After(dbTrans.EXPECT().QueryAccount(w.AccountID{ID: "aaa", Currency: "BBB"}, true).
					Return(&w.Account{ID: w.AccountID{ID: "aaa", Currency: "BBB"}, Balance: 3}, &pk3, nil).
					After(dbTrans.EXPECT().QueryAccount(w.AccountID{ID: "bbb", Currency: "AAA"}, true).
						Return(&w.Account{ID: w.AccountID{ID: "bbb", Currency: "AAA"}, Balance: 2}, &pk2, nil).
						After(dbTrans.EXPECT().QueryAccount(w.AccountID{ID: "aaa", Currency: "AAA"}, true).
							Return(&w.Account{ID: w.AccountID{ID: "aaa", Currency: "AAA"}, Balance: 1}, &pk1, nil).
							After(db.EXPECT().Begin().Return(dbTrans, nil)))))))

	repo := w.CreateRepo(db)
	err := repo.Modify(
		transData,
		"tester",
		func(dbTrans w.RepoTrans) error {
			checkCommit()
			account, err := dbTrans.GetAccount(w.AccountID{ID: "bbb", Currency: "AAA"})
			if err != nil || account == nil || account.Balance != 2 {
				test.Errorf(`Failed to get account: "%s".`, err)
			}
			account.Balance = 200
			account, err = dbTrans.GetAccount(w.AccountID{ID: "bbb", Currency: "BBB"})
			if err != nil || account == nil || account.Balance != 4 {
				test.Errorf(`Failed to get account: "%s".`, err)
			}
			account.Balance = 400
			account, err = dbTrans.GetAccount(w.AccountID{ID: "5", Currency: "555"})
			if err == nil || err.Error() != `Account "5" (555) was not prefetched` ||
				account != nil {

				test.Errorf(`Error expected: %v, %v.`, err, account)
			}
			return nil
		})
	if err != nil {
		test.Errorf(`Failed to modify: "%s".`, err)
	}
}

// Test_Repo_Modify_Success tests repository modification commit error.
func Test_Repo_Modify_CommitError(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	transData := []w.BalanceAction{
		w.BalanceAction{Account: w.AccountID{ID: "bbb", Currency: "AAA"}, Volume: 2}}
	pk1 := 1

	hasCommit := false
	checkCommit := func(...interface{}) {
		if hasCommit {
			test.Error("Update after commit.")
		}
	}

	errText := "Test error"
	db := mw.NewMockDB(ctrl)
	dbTrans := mw.NewMockDBTrans(ctrl)
	dbTrans.EXPECT().Rollback().
		After(dbTrans.EXPECT().Commit().Return(errors.New(errText)).Do(func(...interface{}) { hasCommit = true }).
			After(dbTrans.EXPECT().QueryAccount(w.AccountID{ID: "bbb", Currency: "AAA"}, true).
				Return(&w.Account{ID: w.AccountID{ID: "bbb", Currency: "AAA"}, Balance: 1}, &pk1, nil).
				Do(func(...interface{}) {
					dbTrans.EXPECT().UpdateAccount(w.Account{ID: w.AccountID{ID: "bbb", Currency: "AAA"}, Balance: 1}, 1).Return(nil).Do(checkCommit)
					transPk := 99
					dbTrans.EXPECT().InsertTrans(gomock.Any(), "tester").Return(&transPk, nil).Do(checkCommit)
					dbTrans.EXPECT().InsertAction(gomock.Any(), transPk, gomock.Any()).Times(len(transData)).Return(nil).Do(checkCommit)
				}).
				After(db.EXPECT().Begin().Return(dbTrans, nil))))

	repo := w.CreateRepo(db)
	err := repo.Modify(
		transData,
		"tester",
		func(dbTrans w.RepoTrans) error {
			checkCommit()
			return nil
		})
	if err == nil || err.Error() != errText {
		test.Errorf(`Wrong error status: "%v".`, err)
	}
}

// Test_Repo_Modify_UpdateError tests repository account update error.
func Test_Repo_Modify_UpdateError(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	transData := []w.BalanceAction{
		w.BalanceAction{Account: w.AccountID{ID: "aaa", Currency: "AAA"}, Volume: 1112}}
	pk1 := 1

	hasRollback := false
	checkRollback := func(...interface{}) {
		if hasRollback {
			test.Error("Update after rollback.")
		}
	}

	errText := "Test error"
	db := mw.NewMockDB(ctrl)
	dbTrans := mw.NewMockDBTrans(ctrl)
	dbTrans.EXPECT().Rollback().Do(func(...interface{}) { hasRollback = true }).
		After(dbTrans.EXPECT().QueryAccount(w.AccountID{ID: "aaa", Currency: "AAA"}, true).
			Return(&w.Account{ID: w.AccountID{ID: "aaa", Currency: "AAA"}, Balance: 2}, &pk1, nil).
			Do(func(...interface{}) {
				dbTrans.EXPECT().UpdateAccount(w.Account{ID: w.AccountID{ID: "aaa", Currency: "AAA"}, Balance: 2}, 1).
					Return(errors.New(errText)).Do(checkRollback)
				transPk := 99
				dbTrans.EXPECT().InsertTrans(gomock.Any(), "tester").Return(&transPk, nil).Do(checkRollback)
				dbTrans.EXPECT().InsertAction(gomock.Any(), transPk, gomock.Any()).Times(len(transData)).Return(nil).Do(checkRollback)
			}).
			After(db.EXPECT().Begin().Return(dbTrans, nil)))

	repo := w.CreateRepo(db)
	err := repo.Modify(
		transData,
		"tester",
		func(dbTrans w.RepoTrans) error {
			checkRollback()
			return nil
		})
	if err == nil || err.Error() != errText {
		test.Errorf(`Wrong error status: "%v".`, err)
	}
}

// Test_Repo_Modify_UpdateError tests repository action insert error.
func Test_Repo_Modify_ActionError(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	transData := []w.BalanceAction{
		w.BalanceAction{Account: w.AccountID{ID: "aaa", Currency: "AAA"}, Volume: 1112}}
	pk1 := 1

	hasRollback := false
	checkRollback := func(...interface{}) {
		if hasRollback {
			test.Error("Update after rollback.")
		}
	}

	errText := "Test error"
	db := mw.NewMockDB(ctrl)
	dbTrans := mw.NewMockDBTrans(ctrl)
	dbTrans.EXPECT().Rollback().Do(func(...interface{}) { hasRollback = true }).
		After(dbTrans.EXPECT().QueryAccount(w.AccountID{ID: "aaa", Currency: "AAA"}, true).
			Return(&w.Account{ID: w.AccountID{ID: "aaa", Currency: "AAA"}, Balance: 2}, &pk1, nil).
			Do(func(...interface{}) {
				transPk := 99
				dbTrans.EXPECT().InsertTrans(gomock.Any(), "tester").Return(&transPk, nil).Do(checkRollback)
				dbTrans.EXPECT().InsertAction(gomock.Any(), transPk, gomock.Any()).Times(len(transData)).
					Return(errors.New(errText)).Do(checkRollback)
			}).
			After(db.EXPECT().Begin().Return(dbTrans, nil)))

	repo := w.CreateRepo(db)
	err := repo.Modify(
		transData,
		"tester",
		func(dbTrans w.RepoTrans) error {
			checkRollback()
			return nil
		})
	if err == nil || err.Error() != errText {
		test.Errorf(`Wrong error status: "%v".`, err)
	}
}

// Test_Repo_Modify_UpdateError tests repository transaction insert error.
func Test_Repo_Modify_TransError(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	transData := []w.BalanceAction{
		w.BalanceAction{Account: w.AccountID{ID: "aaa", Currency: "AAA"}, Volume: 1112}}
	pk1 := 1

	hasRollback := false
	checkRollback := func(...interface{}) {
		if hasRollback {
			test.Error("Update after rollback.")
		}
	}

	errText := "Test error"
	db := mw.NewMockDB(ctrl)
	dbTrans := mw.NewMockDBTrans(ctrl)
	dbTrans.EXPECT().Rollback().Do(func(...interface{}) { hasRollback = true }).
		After(dbTrans.EXPECT().QueryAccount(w.AccountID{ID: "aaa", Currency: "AAA"}, true).
			Return(&w.Account{ID: w.AccountID{ID: "aaa", Currency: "AAA"}, Balance: 2}, &pk1, nil).
			Do(func(...interface{}) {
				dbTrans.EXPECT().InsertTrans(gomock.Any(), "tester").
					Return(nil, errors.New(errText)).Do(checkRollback)
			}).
			After(db.EXPECT().Begin().Return(dbTrans, nil)))

	repo := w.CreateRepo(db)
	err := repo.Modify(
		transData,
		"tester",
		func(dbTrans w.RepoTrans) error {
			checkRollback()
			return nil
		})
	if err == nil || err.Error() != errText {
		test.Errorf(`Wrong error status: "%v".`, err)
	}
}

// Test_Repo_Modify_CallbackError tests modify-callback error.
func Test_Repo_Modify_CallbackError(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	transData := []w.BalanceAction{
		w.BalanceAction{Account: w.AccountID{ID: "aaa", Currency: "AAA"}, Volume: 1112}}
	pk1 := 1

	hasRollback := false
	db := mw.NewMockDB(ctrl)
	dbTrans := mw.NewMockDBTrans(ctrl)
	dbTrans.EXPECT().Rollback().Do(func(...interface{}) { hasRollback = true }).
		After(dbTrans.EXPECT().QueryAccount(w.AccountID{ID: "aaa", Currency: "AAA"}, true).
			Return(&w.Account{ID: w.AccountID{ID: "aaa", Currency: "AAA"}, Balance: 1}, &pk1, nil).
			After(db.EXPECT().Begin().Return(dbTrans, nil)))

	repo := w.CreateRepo(db)
	errText := "Test error"
	err := repo.Modify(
		transData, "tester",
		func(dbTrans w.RepoTrans) error {
			if hasRollback {
				test.Error("Update after rollback.")
			}
			return errors.New(errText)
		})
	if err == nil || err.Error() != errText {
		test.Errorf(`Wrong error status: "%v".`, err)
	}
}

// Test_Repo_Modify_QueryError tests repository account querying error.
func Test_Repo_Modify_QueryError(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	transData := []w.BalanceAction{
		w.BalanceAction{Account: w.AccountID{ID: "bbb", Currency: "AAA"}, Volume: 2},
		w.BalanceAction{Account: w.AccountID{ID: "aaa", Currency: "AAA"}, Volume: 1112},
		w.BalanceAction{Account: w.AccountID{ID: "ссс", Currency: "AAA"}, Volume: -2}}
	pk1 := 1

	errText := "Test error"
	db := mw.NewMockDB(ctrl)
	dbTrans := mw.NewMockDBTrans(ctrl)
	dbTrans.EXPECT().Rollback().
		After(dbTrans.EXPECT().QueryAccount(w.AccountID{ID: "bbb", Currency: "AAA"}, true).
			Return(nil, nil, errors.New(errText)).
			After(dbTrans.EXPECT().QueryAccount(w.AccountID{ID: "aaa", Currency: "AAA"}, true).
				Return(&w.Account{ID: w.AccountID{ID: "aaa", Currency: "AAA"}, Balance: 1}, &pk1, nil).
				After(db.EXPECT().Begin().Return(dbTrans, nil))))

	repo := w.CreateRepo(db)
	err := repo.Modify(
		transData,
		"tester",
		func(dbTrans w.RepoTrans) error {
			test.Error("Unexpected callback call.")
			return nil
		})
	if err == nil || err.Error() != errText {
		test.Errorf(`Wrong error status: "%v".`, err)
	}

}

// Test_Repo_Modify_BeginError tests repository modification beginning error.
func Test_Repo_Modify_BeginError(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	transData := []w.BalanceAction{
		w.BalanceAction{Account: w.AccountID{ID: "bbb", Currency: "AAA"}, Volume: 2},
		w.BalanceAction{Account: w.AccountID{ID: "aaa", Currency: "AAA"}, Volume: 1112}}

	errText := "Test error"
	db := mw.NewMockDB(ctrl)
	db.EXPECT().Begin().Return(nil, errors.New(errText))

	repo := w.CreateRepo(db)
	err := repo.Modify(
		transData,
		"tester",
		func(trans w.RepoTrans) error {
			test.Error("Unexpected callback call.")
			return nil
		})
	if err == nil || err.Error() != errText {
		test.Errorf(`Wrong error status: "%v".`, err)
	}

}

// Test_Repo_Modify_Lists tests repository lists requests.
func Test_Repo_Modify_Lists(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	db := mw.NewMockDB(ctrl)
	accounts := []w.Account{
		{ID: w.AccountID{ID: "123", Currency: "345"},
			Balance: 456.678},
		{ID: w.AccountID{ID: "678", Currency: "098"},
			Balance: 123.123}}
	transList := []w.Trans{
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
	errText := "Test error"
	db.EXPECT().GetAccounts().Return(accounts, errors.New(errText))
	db.EXPECT().GetTransList().Return(transList, errors.New(errText))

	repo := w.CreateRepo(db)

	{
		result, err := repo.GetAccounts()
		if len(accounts) != len(result) {
			test.Errorf("Wrong result: %v.", result)
		} else {
			for i, account := range result {
				if account != accounts[i] {
					test.Errorf("Wrong result: %v.", result)
				}
			}
		}
		if err == nil || err.Error() != errText {
			test.Errorf("Wrong error: %v.", err)
		}
	}
	{
		result, err := repo.GetTransList()
		if len(transList) != len(result) {
			test.Errorf("Wrong result: %v.", result)
		} else {
			for i, trans := range result {
				for j, action := range trans {
					if action != transList[i][j] {
						test.Errorf("Wrong result: %v.", result)
					}
				}
			}
		}
		if err == nil || err.Error() != errText {
			test.Errorf("Wrong error: %v.", err)
		}
	}
}
