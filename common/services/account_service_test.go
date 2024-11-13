package services

import (
	"common/dao"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAccountService_FetchOrCreateAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	accSrv := &AccountService{
		Database: db,
	}

	now := time.Now()

	acc := dao.Account{
		AccountID:       1,
		FirstName:       "John",
		LastName:        "Doe",
		Email:           "john.doe@example.com",
		Locale:          "es-MX",
		TotalBalance:    sql.NullString{},
		AvgDebitAmount:  sql.NullString{},
		AvgCreditAmount: sql.NullString{},
		LastBalanceAt:   sql.NullTime{},
		CreatedAt:       sql.NullTime{Valid: true, Time: now},
		UpdatedAt:       sql.NullTime{Valid: true, Time: now},
	}

	accountColumns := []string{
		"account_id",
		"first_name",
		"last_name",
		"email",
		"locale",
		"total_balance",
		"avg_debit_amount",
		"avg_credit_amount",
		"last_balance_at",
		"created_at",
		"updated_at",
	}
	accountRow := []driver.Value{
		acc.AccountID,
		acc.FirstName,
		acc.LastName,
		acc.Email,
		acc.Locale,
		acc.TotalBalance,
		acc.AvgDebitAmount,
		acc.AvgCreditAmount,
		acc.LastBalanceAt,
		acc.CreatedAt,
		acc.UpdatedAt,
	}

	fetchAccountQuery := `SELECT 
    	account_id, 
    	first_name, 
    	last_name, 
    	email, 
    	locale, 
    	total_balance, 
    	avg_debit_amount, 
    	avg_credit_amount, 
    	last_balance_at, 
    	created_at, 
    	updated_at 
	FROM accounts WHERE email = \$1 LIMIT 1`
	insertAccountQuery := `INSERT INTO accounts`

	t.Run("ExistingAccount", func(t *testing.T) {
		mock.ExpectQuery(fetchAccountQuery).WithArgs("john.doe@example.com").WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(accountRow...))
		res, err := accSrv.FetchOrCreateAccount(context.Background(), "john.doe@example.com", "John", "Doe")

		require.NoError(t, err)
		require.Equal(t, acc, res)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NewAccount", func(t *testing.T) {
		mock.ExpectQuery(fetchAccountQuery).WithArgs("john.doe@example.com").WillReturnRows(sqlmock.NewRows(accountColumns)) // Simulating account creation
		mock.ExpectQuery(insertAccountQuery).WithArgs("John", "Doe", "john.doe@example.com", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(accountRow...))

		res, err := accSrv.FetchOrCreateAccount(context.Background(), "john.doe@example.com", "John", "Doe")

		require.NoError(t, err)
		require.Equal(t, acc, res)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(fetchAccountQuery).WithArgs("test@example.com").WillReturnError(errors.New("db_error"))

		_, err := accSrv.FetchOrCreateAccount(context.Background(), "test@example.com", "Test", "User")
		require.Error(t, err)
		require.Contains(t, err.Error(), "db_error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAccountService_UpdateAccountBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	accSrv := &AccountService{
		Database: db,
	}

	now := time.Now()

	acc := dao.Account{
		AccountID:       1,
		FirstName:       "John",
		LastName:        "Doe",
		Email:           "john.doe@example.com",
		Locale:          "es-MX",
		TotalBalance:    sql.NullString{},
		AvgDebitAmount:  sql.NullString{},
		AvgCreditAmount: sql.NullString{},
		LastBalanceAt:   sql.NullTime{},
		CreatedAt:       sql.NullTime{Valid: true, Time: now},
		UpdatedAt:       sql.NullTime{Valid: true, Time: now},
	}

	report := BalanceReport{
		AccountID:        1,
		TotalCredit:      decimal.Decimal{},
		CountCredit:      1,
		TotalDebit:       decimal.Decimal{},
		CountDebit:       1,
		TotalBalance:     decimal.NewFromFloat32(10.0),
		AvgDebitAmount:   decimal.NewFromFloat32(11.1),
		AvgCreditAmount:  decimal.NewFromFloat32(12.2),
		TransactionCount: map[int]int{1: 1},
	}

	accountRow := []driver.Value{
		acc.AccountID,
		acc.FirstName,
		acc.LastName,
		acc.Email,
		acc.Locale,
		acc.TotalBalance,
		acc.AvgDebitAmount,
		acc.AvgCreditAmount,
		acc.LastBalanceAt,
		acc.CreatedAt,
		acc.UpdatedAt,
	}

	updateAccountBalanceQuery := `
		UPDATE accounts
			SET last_balance_at = \$1, total_balance = \$2, avg_debit_amount = \$3, avg_credit_amount = \$4
		WHERE account_id = \$5
		RETURNING 
			account_id, 
			first_name, 
			last_name,
			email, 
			locale, 
			total_balance, 
			avg_debit_amount, 
			avg_credit_amount, 
			last_balance_at, 
			created_at, 
			updated_at`

	accountColumns := []string{
		"account_id",
		"first_name",
		"last_name",
		"email",
		"locale",
		"total_balance",
		"avg_debit_amount",
		"avg_credit_amount",
		"last_balance_at",
		"created_at",
		"updated_at",
	}

	t.Run("UpdateBalance", func(t *testing.T) {
		args := []driver.Value{
			sqlmock.AnyArg(),
			report.TotalBalance,
			report.AvgDebitAmount,
			report.AvgCreditAmount,
			report.AccountID,
		}
		mock.ExpectQuery(updateAccountBalanceQuery).WithArgs(args...).WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(accountRow...))
		res, err := accSrv.UpdateAccountBalance(context.Background(), acc, report)

		require.NoError(t, err)
		require.Equal(t, acc, res)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DBError", func(t *testing.T) {
		args := []driver.Value{
			sqlmock.AnyArg(),
			report.TotalBalance,
			report.AvgDebitAmount,
			report.AvgCreditAmount,
			report.AccountID,
		}
		mock.ExpectQuery(updateAccountBalanceQuery).WithArgs(args...).WillReturnError(errors.New("db_error"))

		_, err := accSrv.UpdateAccountBalance(context.Background(), acc, report)
		require.Error(t, err)
		require.Contains(t, err.Error(), "db_error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
