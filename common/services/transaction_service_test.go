package services

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/csv"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransactionService_ProcessFile(t *testing.T) {
	testCases := []struct {
		name           string
		csvContent     string
		expectError    bool
		expectInsert   bool
		accountID      int64
		expectedArgs   []driver.Value
		expectedReport BalanceReport
	}{
		{"Invalid CSV", "", true, false, 0, []driver.Value{}, BalanceReport{}},
		{"Invalid row", "ID,DATE,AMOUNT\nA,B,C", false, false, 1, []driver.Value{}, BalanceReport{
			AccountID:        1,
			TotalCredit:      decimal.Zero,
			CountCredit:      0,
			TotalDebit:       decimal.Zero,
			CountDebit:       0,
			TotalBalance:     decimal.Zero,
			AvgDebitAmount:   decimal.Zero,
			AvgCreditAmount:  decimal.Zero,
			TransactionCount: make(map[int]int),
		}},
		{"Single debit", "ID,DATE,AMOUNT\n1,01/01,+1.5", false, true, 1,
			[]driver.Value{1, "credit", decimal.NewFromFloat(1.5), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()},
			BalanceReport{
				AccountID:        1,
				TotalCredit:      decimal.NewFromFloat(1.5),
				CountCredit:      1,
				TotalDebit:       decimal.Zero,
				CountDebit:       0,
				TotalBalance:     decimal.NewFromFloat(1.5),
				AvgDebitAmount:   decimal.Zero,
				AvgCreditAmount:  decimal.NewFromFloat(1.5),
				TransactionCount: map[int]int{1: 1},
			},
		},
		{"Single credit", "ID,DATE,AMOUNT\n1,01/01,-1.5", false, true, 1,
			[]driver.Value{1, "debit", decimal.NewFromFloat(1.5), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()},
			BalanceReport{
				AccountID:        1,
				TotalCredit:      decimal.Zero,
				CountCredit:      0,
				TotalDebit:       decimal.NewFromFloat(1.5),
				CountDebit:       1,
				TotalBalance:     decimal.NewFromFloat(-1.5),
				AvgDebitAmount:   decimal.NewFromFloat(1.5),
				AvgCreditAmount:  decimal.Zero,
				TransactionCount: map[int]int{1: 1},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			service := TransactionService{
				Database:  db,
				Workers:   1,
				BatchSize: 1,
			}

			if tc.expectInsert {
				mock.ExpectQuery(`INSERT INTO transactions`).WithArgs(tc.expectedArgs...).
					WillReturnRows(sqlmock.NewRows([]string{"transaction_id"}).AddRow(1))
			}

			reader := csv.NewReader(bytes.NewBufferString(tc.csvContent))
			report, err := service.ProcessFile(context.Background(), tc.accountID, reader)
			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
			require.Equal(t, tc.expectedReport.AccountID, report.AccountID)
			require.Equal(t, tc.expectedReport.CountDebit, report.CountDebit)
			require.Equal(t, tc.expectedReport.CountCredit, report.CountCredit)
			require.Equal(t, tc.expectedReport.TransactionCount, report.TransactionCount)
			require.Equal(t, tc.expectedReport.TotalBalance.StringFixed(2), report.TotalBalance.StringFixed(2))
			require.Equal(t, tc.expectedReport.TotalDebit.StringFixed(2), report.TotalDebit.StringFixed(2))
			require.Equal(t, tc.expectedReport.TotalCredit.StringFixed(2), report.TotalCredit.StringFixed(2))
			require.Equal(t, tc.expectedReport.AvgDebitAmount.StringFixed(2), report.AvgDebitAmount.StringFixed(2))
			require.Equal(t, tc.expectedReport.AvgCreditAmount.StringFixed(2), report.AvgCreditAmount.StringFixed(2))
		})
	}
}
