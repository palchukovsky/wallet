package wallet

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

////////////////////////////////////////////////////////////////////////////////

// DBTrans represents an active interface to a database in one transaction. All
// actions are collected by transaction and has to be written by commit
// operations.
type DBTrans interface {
	// Commit commits all changes since the start if the rollback was not called
	// before.
	Commit() error
	// Rollback rollbacks all changes since the start if the commit was not called
	// before.
	Rollback()

	// QueryAccount returns state of account and account primary key from a
	// database.
	QueryAccount(request AccountID, lock bool) (*Account, *int, error)
	// StoreAccount stores state of account into a database.
	StoreAccount(account Account) error
	// UpdateAccount updates state of account only if it existent.
	UpdateAccount(account Account, pk int) error

	// InsertTrans inserts record about transaction into a database and
	// return transaction primary key.
	InsertTrans(time time.Time, author string) (*int, error)
	// InsertAction inserts record about action into a database.
	InsertAction(accountPk, transPk int, actionVolume float64) error
}

////////////////////////////////////////////////////////////////////////////////

// DB represents a database connection.
type DB interface {
	// Close closes database connection and frees resources.
	Close()

	// Begin starts a new database transaction to execute database IO operations.
	Begin() (DBTrans, error)
}

////////////////////////////////////////////////////////////////////////////////

type dbTrans struct{ tx *sql.Tx }

func (t *dbTrans) Commit() error {
	if t.tx == nil {
		return nil
	}
	if err := t.tx.Commit(); err != nil {
		return err
	}
	t.tx = nil
	return nil
}

func (t *dbTrans) Rollback() {
	if t.tx == nil {
		return
	}
	if err := t.tx.Rollback(); err != nil {
		// There is no way to restore application state at error at rollback, the
		// behavior is undefined, so the application must be stopped.
		log.Panicf(`Failed to commit database transaction: "%s".`, err)
	}
	t.tx = nil
}

func (t *dbTrans) QueryAccount(
	request AccountID, lock bool) (*Account, *int, error) {

	query := "SELECT id, balance FROM account WHERE currency = $2 AND name = $1"
	if lock {
		query += " FOR UPDATE"
	}
	row := t.tx.QueryRow(query, request.ID, request.Currency)
	var primaryKey int
	account := &Account{ID: request}
	if err := row.Scan(&primaryKey, &account.Balance); err != nil {
		return nil, nil, err
	}
	return account, &primaryKey, nil
}

func (t *dbTrans) StoreAccount(account Account) error {
	_, err := t.tx.Exec(
		"INSERT INTO account(name, currency, balance) VALUES($1, $2, $3)"+
			" ON CONFLICT ON CONSTRAINT account_unique DO UPDATE SET balance = $3",
		account.ID.ID, account.ID.Currency, account.Balance)
	return err
}

func (t *dbTrans) UpdateAccount(account Account, pk int) error {
	_, err := t.tx.Exec(
		"UPDATE account SET name = $1, currency = $2, balance = $3"+
			" WHERE id = $4",
		account.ID.ID, account.ID.Currency, account.Balance, pk)
	return err
}

func (t *dbTrans) InsertTrans(time time.Time, author string) (*int, error) {
	row := t.tx.QueryRow(
		"INSERT INTO trans(time, author) VALUES($1, $2) RETURNING id",
		time, author)
	result := 0
	if err := row.Scan(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (t *dbTrans) InsertAction(
	accountPk, transPk int, actionVolume float64) error {

	_, err := t.tx.Exec(
		"INSERT INTO action(account, trans, volume) VALUES($1, $2, $3)",
		accountPk, transPk, actionVolume)
	return err
}

////////////////////////////////////////////////////////////////////////////////

type pgDB struct {
	conn *sql.DB
}

// CreateDB creates a wallet database connection with the provided data source
// access parameters.
func CreateDB(host, name, login, password string) (DB, error) {
	result := &pgDB{}
	var err error
	result.conn, err = sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
			login, password, host, name))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *pgDB) Close() { db.conn.Close() }

func (db *pgDB) Begin() (DBTrans, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, err
	}
	return &dbTrans{tx: tx}, nil
}

////////////////////////////////////////////////////////////////////////////////
