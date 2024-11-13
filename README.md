# Account balance

This is a take home challenge for a software engineer position. Requirements are:

* Process account balances from a file (`support/files/transactions.csv`).
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
mkdir -p support/data
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
docker compose run gen-txns-csv -file support/files/transactions.csv
```

There are additional parameters that change amount of transactions, dates and amount:
```sh
docker compose run gen-txns-csv -help
```

## Processing transactions

To generate a balance report of an account, run the following command:
```sh
docker compose run proc-txns-csv -file support/files/transactions.csv
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

This project has a workspace with three different main modules:

* cmd - Command line tools to handle generation and processing of transactions
* common - Data Access Objects and Services that cmd and lambdas use.
* lambda - AWS Lambda version of the processing command.

```
cmd/              <- Command line module.
  gen-txns-csv/   <- CSV Generator command.
  proc-txns-csv/  <- CSV Processor command.
  
common/       <- Common libraries and utilities.
  dao/        <- Go package, (github.com/cedmundo/account-balance/dao) generated SQL utilites (from sqlc).
  services/   <- Go package, (github.com/cedmundo/account-balance/services) that manages the business logic.
    static/   <- Email templates and static content.

lambda/       <- Lambda version of processor command.

support/      <- Misc files.
  data/       <- Persistent data of PostgreSQL (requires to be empty).
  db/         <- SQL Files (required for sqlc).
  files/      <- CSV and support files (automatically mounted).

```

## Decimals

It is well known to [never use floats for money](https://husobee.github.io/money/float/2016/09/23/never-use-floats-for-currency.html), that's
why I use `decimal.Decimal` in Go and use `DECIMAL(16, 2)` in PSQL for the money fields, Although I would still recommend
to use a more specialized monetary type instead. This project doesn't include one because the currency is only MXN and
was not actually in the specification, so I didn't want to add unnecessary complexity.

## SQLc Generation

This project uses [sqlc](https://docs.sqlc.dev/en/latest/) for code generation, queries are located at `db/query.sql` 
and schema is located at `db/schema.sql` in a larger project I would also add migrations, however, this time it is only 
needed to set the database once, that's why it is configured to run when creating the postgres node in the composer file.

## Database Querying

It is possible to plug a PostgresSQL console into the server and explore the database, to get the balance numbers just use the
following query:

```sql
WITH balance AS (
    SELECT
        SUM(CASE WHEN operation = 'debit' THEN amount ELSE 0 END) AS total_debit,
        SUM(CASE WHEN operation = 'debit' THEN 1 ELSE 0 END) AS debit_count,
        SUM(CASE WHEN operation = 'credit' THEN amount ELSE 0 END) AS total_credit,
        SUM(CASE WHEN operation = 'credit' THEN 1 ELSE 0 END) AS credit_count
    FROM transactions WHERE account_id = 20
) SELECT
    balance.total_credit,
    balance.total_debit,
    balance.credit_count,
    balance.debit_count,
    balance.total_credit - balance.total_debit AS total_balance,
    balance.total_credit / balance.credit_count AS avg_credit,
    balance.total_debit / balance.debit_count AS avg_debit
FROM balance;
```

Note that you might need to change the `account_id` from `20` to the ID you want to query, or remove it to perform a general
report.

## Batch processing

The transaction processing is performed in a producer/consumer manner, each `worker` has its own database connection and
pulls CSV records from a channel, then it returns its results into a reports channel which are reduced by the worker scheduler.

The workflow is described as follows:
```
                    /- (<-CSVRecord) Worker 0 (->Report) -\
                    |                                     |
Transaction Service +- (<-CSVRecord) Worker 1 (->Report) -+- [ Worker Report Reduce ] - [ Balance Report ]  
                    |                                     |
                    \- (<-CSVRecord) Worker n (->Report) -/
```

The database insertion is also handled by the worker so they pull jobs with more or less same recurrence.

## Content embed

HTML email template and JSON localization messages are both [embed](https://pkg.go.dev/embed) into `proc-txns-csv` 
executable so it is even easier to deploy because there is no need to drag `static` folder around.