package services

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/cedmundo/account-balance/dao"
	"github.com/shopspring/decimal"
	"html/template"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

//go:embed all:static
var content embed.FS
var emailTemplate = template.Must(template.New("balance_report.html").Funcs(template.FuncMap{
	"moneyFmt": func(decimal decimal.Decimal) string {
		return "$" + decimal.StringFixedBank(2)
	},
	"monthToString": func(loc map[string]string, m int) string {
		return loc["month."+strconv.Itoa(m)]
	},
}).ParseFS(content, "static/email/balance_report.html"))

// EmailData represents the data required to generate an email report for an account.
type EmailData struct {
	Account                dao.Account
	TitleMsg               string
	SubtitleMsg            string
	CheckBalanceMsg        string
	CheckBalanceLink       string
	FooterMsg              string
	TotalBalanceMsg        string
	AvgCreditAmountMsg     string
	AvgDebitAmountMsg      string
	TransactionsInMonthMsg string
	Locale                 map[string]string
	Report                 BalanceReport
}

// EmailService manages sending emails and loading localization messages for emails.
type EmailService struct {
	PublicURL string
	Messages  map[string]map[string]string
}

// LoadMessages loads localization messages
func (s *EmailService) LoadMessages() error {
	s.Messages = make(map[string]map[string]string)
	jsonData, err := content.ReadFile("static/email/messages.json")
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, &s.Messages)
}

// SendReport sends a balance report to the specified account's email.
func (s *EmailService) SendReport(account dao.Account, report BalanceReport) error {
	var output io.Writer
	accountID := account.AccountID
	email := account.Email
	locale := account.Locale

	if strings.HasPrefix(email, "fake+") {
		log.Println("Sending fake email to", email, "rendering to files/fake_email.html")
		writer, err := os.Create("files/fake_email.html")
		if err != nil {
			return err
		}
		defer func(writer *os.File) {
			if err := writer.Close(); err != nil {
				log.Println("Error closing fake email file", err)
			}
		}(writer)
		output = writer
	} else {
		output = bytes.NewBufferString("")
	}

	loc := s.Messages[locale]
	err := emailTemplate.Execute(output, EmailData{
		Account:                account,
		Locale:                 loc,
		TitleMsg:               loc["balance_email.title"],
		SubtitleMsg:            loc["balance_email.subtitle"],
		CheckBalanceMsg:        loc["balance_email.check"],
		CheckBalanceLink:       fmt.Sprintf("%s/balance?id=%d", s.PublicURL, accountID),
		FooterMsg:              loc["balance_email.footer"],
		TotalBalanceMsg:        loc["balance_email.total_balance"],
		AvgCreditAmountMsg:     loc["balance_email.avg_credit_amount"],
		AvgDebitAmountMsg:      loc["balance_email.avg_debit_amount"],
		TransactionsInMonthMsg: loc["balance_email.transactions_in_month"],
		Report:                 report,
	})
	if err != nil {
		return err
	}

	return nil
}
