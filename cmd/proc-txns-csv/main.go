package main

import "flag"

var (
	pFile         = flag.String("file", "", "File to read transactions from")
	pAccountId    = flag.Int("account-id", 0, "Account ID to use (leave zero to create one)")
	pAccountName  = flag.String("account-name", "", "Account Name to use when creating accounts (leave blank to random)")
	pAccountEmail = flag.String("account-email", "", "Account Email to use when creating accounts (leave blank to random)")
)

func main() {

}
