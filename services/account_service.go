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

		return queries.CreateAccount(ctx, dao.CreateAccountParams{
			FirstName: account.FirstName,
			LastName:  account.LastName,
			Email:     account.Email,
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
		})
	}

	return account, nil
}

func (s *AccountService) UpdateAccountBalance(ctx context.Context, account dao.Account, report BalanceReport) (dao.Account, error) {
	queries := dao.New(s.Database)
	return queries.UpdateAccountBalance(ctx, dao.UpdateAccountBalanceParams{
		LastBalanceAt:   sql.NullTime{Valid: true, Time: time.Now()},
		TotalBalance:    sql.NullString{Valid: true, String: report.TotalBalance.String()},
		AvgDebitAmount:  sql.NullString{Valid: true, String: report.AvgDebitAmount.String()},
		AvgCreditAmount: sql.NullString{Valid: true, String: report.AvgCreditAmount.String()},
		AccountID:       account.AccountID,
	})
}
