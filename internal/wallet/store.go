package wallet

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/aboronilov/go-payments/internal/wallet/walletdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Account        = walletdb.Account
	AccountBalance = walletdb.AccountBalance
	Entry          = walletdb.Entry
	NewAccount     = walletdb.CreateAccountParams
	Direction      = walletdb.EntryDirection
)

const (
	Debit  = walletdb.EntryDirectionDebit
	Credit = walletdb.EntryDirectionCredit
)

var ErrAccountNotFound = errors.New("wallet: account not found")

// store is the wallet's data access layer (repo) - implements queries and transactions
type Store struct {
	pool *pgxpool.Pool
	q    *walletdb.Queries
}

// new store returns a store that runs queries on pool
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool: pool,
		q:    walletdb.New(pool),
	}
}

// CreateAccount opens a new account. Pwner id is nil for system accounts
func (s *Store) CreateAccount(ctx context.Context, a NewAccount) (Account, error) {
	acc, err := s.q.CreateAccount(ctx, a)
	if err != nil {
		return Account{}, fmt.Errorf("create account: %w", err)
	}
	return acc, nil
}

// Account returns the account with given ID
func (s *Store) Account(ctx context.Context, id uuid.UUID) (Account, error) {
	acc, err := s.q.GetAccount(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrAccountNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("get account %s: %w", id, err)
	}
	return acc, nil
}

func (s *Store) Balance(ctx context.Context, id uuid.UUID) (AccountBalance, error) {
	b, err := s.q.GetAccountBalance(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return AccountBalance{}, ErrAccountNotFound
	}
	if err != nil {
		return AccountBalance{}, fmt.Errorf("get account balance %s: %w", id, err)
	}
	return b, nil
}

// Entries return the page of account's entries, newest first. Pass 0 as before
// the for the first page, then the id of the last entry you received
func (s *Store) Entries(ctx context.Context, id uuid.UUID, before int64, limit int32) ([]Entry, error) {
	if before < 0 {
		before = math.MaxInt64
	}
	entries, err := s.q.ListEntriesByAccount(ctx, walletdb.ListEntriesByAccountParams{
		AccountID: id,
		BeforeID:  before,
		PageSize:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get entries %s: %w", id, err)
	}
	return entries, nil
}
