package services

import (
	"bytes"
	"common/dao"
	"crypto/tls"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"gopkg.in/gomail.v2"
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

// EmailSender interface abstracts only the part to send actual HTML to an email
type EmailSender interface {
	SendHTML(email string, subject string, html string) error
}

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
	Sender    EmailSender
	Messages  map[string]map[string]string
}

// SMTPSender implements the EmailSender interface to send messages through SMTP
type SMTPSender struct {
	FromEmail string
	SMTPHost  string
	SMTPPort  int
	SMTPUser  string
	SMTPPass  string
}

func (s *SMTPSender) SendHTML(email string, subject string, html string) error {
	if s.SMTPHost != "" {
		m := gomail.NewMessage()
		m.SetHeader("From", s.FromEmail)
		m.SetHeader("To", email)
		m.SetHeader("Subject", subject)
		m.SetBody("text/html", html)

		// Send email via SMTP
		d := gomail.NewDialer(s.SMTPHost, s.SMTPPort, s.SMTPUser, s.SMTPPass)
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		if err := d.DialAndSend(m); err != nil {
			return err
		}

		log.Println("Sent email to", email)
	} else if s.SMTPHost == "" {
		log.Println("Skipping email send because SMTPHost is empty")
	}

	return nil
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
	var buffer *bytes.Buffer
	accountID := account.AccountID
	email := account.Email
	locale := account.Locale

	if strings.HasPrefix(email, "fake+") {
		log.Println("Sending fake email to", email, "rendering to files/fake_email.html")
		writer, err := os.Create("support/files/fake_email.html")
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
		buffer = bytes.NewBufferString("")
		output = buffer
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

	// It is probably a good idea to pub/sub this process, since they are only limited email it is ok for now.
	if buffer != nil {
		return s.Sender.SendHTML(email, loc["balance_email.title"], buffer.String())
	}
	return nil
}
