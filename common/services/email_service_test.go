package services

import (
	"common/dao"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

type MockSender struct {
	SentTo      string
	SentSubject string
	SentHTML    string
}

func (m *MockSender) SendHTML(email string, subject string, html string) error {
	m.SentTo = email
	m.SentSubject = subject
	m.SentHTML = html
	return nil
}

func TestEmailService_SendReport(t *testing.T) {
	mockSender := &MockSender{}
	service := &EmailService{Sender: mockSender, PublicURL: "http://localhost:3000"}
	require.NoError(t, service.LoadMessages())

	account := dao.Account{
		AccountID: 1,
		Email:     "test@example.com",
		Locale:    "en-US",
	}

	report := BalanceReport{
		AccountID:   1,
		TotalCredit: decimal.NewFromInt(100),
		CountCredit: 10,
		TotalDebit:  decimal.NewFromInt(50),
		CountDebit:  5,
	}

	err := service.SendReport(account, report)
	require.NoError(t, err)
	require.Equal(t, account.Email, mockSender.SentTo)
	require.Equal(t, "Balance Report", mockSender.SentSubject)
}
