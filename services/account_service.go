package services

import (
	"context"
	"database/sql"
	"github.com/cedmundo/account-balance/dao"
	"time"
)

// AccountService abstracts business logic over DAO.
type AccountService struct {
	Database *sql.DB
}

// FetchOrCreateAccount fetches an account by email, or creates a new one if none exists.
func (s *AccountService) FetchOrCreateAccount(ctx context.Context, email, firstName, lastName string) (dao.Account, error) {
	queries := dao.New(s.Database)
	account, err := queries.GetAccountByEmail(ctx, email)
	if err != nil {
		now := time.Now()
		account = dao.Account{
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			CreatedAt: sql.NullTime{Valid: true, Time: now},
			UpdatedAt: sql.NullTime{Valid: true, Time: now},
		}

		account.AccountID, err = queries.CreateAccount(ctx, dao.CreateAccountParams{
			FirstName: account.FirstName,
			LastName:  account.LastName,
			Email:     account.Email,
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
		})
		return account, err
	}

	return account, nil
}
