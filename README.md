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

## Using locally with docker-compose

Make sure `data` directory exists (to persist database):

```
mkdir -p data
```

Start database service running:
```
docker compose up --remove-orphans
```

Any other command can be run afterward, i.e. [process transactions](#processing-transactions).

Note: You might need to run commands as root, even when your user belongs to `docker` group, since it will mount `data` directory
as storage for postgres and sometimes it brakes.

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

To generate a balance report of an account, run the following command:
```sh
docker compose run proc-txns-csv -file files/transactions.csv
```

Optional parameters are:
```
-account-email <email>
-account-first-name <name>
-account-last-name <name>
```

If none of them are given then a new account will be created, if `account-email` is given but
the account does not exist, then it will be also created with fake names, if the `account-email` exist
then that account will be used to attach all processed transactions.

Additional options are available, for example, to specify database URL or worker pool size:
```sh
docker compose run proc-txns-csv -help
```

## Project structure

The project is an entire golang package, the stand-alone commands are located at `cmd`, however, there are other support
files and directories to work alongside docker-compose:
```
cmd/
  gen-txns-csv/   <- CSV Generator command
  proc-txns-csv/  <- CSV Processor command
  
dao/        <- Go package, (github.com/cedmundo/account-balance/dao) generated SQL utilites (from sqlc)
data/       <- Persistent data of PostgreSQL (requires to be empty)
db/         <- SQL Files (required for sqlc)
files/      <- CSV and support files (automatically mounted)
services/   <- Go package, (github.com/cedmundo/account-balance/services) that manages the business logic
```

