# Account balance

This is a take home challenge for a software engineer position. Requirements are:

* Process account balances from a file (`files/transactions.csv`).
  * Calculate total balance.
  * Calculate number of transactions monthly.
  * Calculate average debit amount.
  * Calculate average credit amount.
* Format report in HTML for email.
* Optionally save account balance into database.
* Package and run code in a could platform provider.

## Generating transactions

This exercise include random generator for the transaction data. The utility is in `cmd/gen-txns-csv` and it can be built
independently of the project. However, a docker version is available:
```sh
docker compose run gen-txns-csv -file files/transactions.csv
```

There are additional parameters that change amount of transactions, dates and amount:
```sh
docker compose run gen-txns-csv -help
```

## Processing transactions

TODO