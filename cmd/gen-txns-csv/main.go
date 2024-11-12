package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	dayMonthLayout      = "01/02"
	maxRandomGeneration = 10000
)

var (
	pFile      = flag.String("file", "", "File to output transactions to")
	pSeed      = flag.Int64("seed", 0, "Seed for random number generator (leave empty for current time)")
	pGenerate  = flag.Uint("gen", 1000, "Number of transactions to generate (leave empty for random)")
	pDateMin   = flag.String("date-min", "", "Minimum date to randomly pick from, YYYY-MM-DD format (leave empty for last year)")
	pDateMax   = flag.String("date-max", "", "Maximum date to randomly pick from, YYYY-MM-DD format (leave empty for today)")
	pAmountMin = flag.Float64("amount-min", 0.0, "Minimum amount to randomly pick from (fix to two decimals, MXN)")
	pAmountMax = flag.Float64("amount-max", 10000.0, "Maximum amount to randomly pick from (fix to two decimals, MXN)")
	rng        *rand.Rand
)

func flagSeed() int64 {
	if *pSeed == 0 {
		return time.Now().UnixNano()
	}
	return *pSeed
}

func flagGenerate() uint {
	if *pGenerate == 0 {
		return uint(rng.Intn(maxRandomGeneration))
	}

	return *pGenerate
}

func flagDates() (dateMin, dateMax time.Time) {
	var err error
	dateMin = time.Now().AddDate(-1, 0, 0)
	if *pDateMin != "" {
		if dateMin, err = time.Parse(time.DateOnly, *pDateMin); err != nil {
			log.Fatalf("Error parsing date min: %v", err)
		}
	}

	dateMax = time.Now()
	if *pDateMax != "" {
		if dateMax, err = time.Parse(time.DateOnly, *pDateMax); err != nil {
			log.Fatalf("Error parsing date max: %v", err)
		}
	}
	return
}

func flagAmounts() (amountMin, amountMax float64) {
	return *pAmountMin, *pAmountMax
}

func randomDebitOrCredit() string {
	if rng.Intn(2) == 0 {
		return "+"
	}

	return "-"
}

func randomDateBetween(minDate, maxDate time.Time) time.Time {
	return minDate.Add(time.Duration(rng.Intn(int(maxDate.Sub(minDate).Seconds()))) * time.Second)
}

func randomAmountBetween(minAmount, maxAmount float64) float64 {
	return minAmount + float64(rng.Intn(int(maxAmount-minAmount)))
}

func main() {
	flag.Parse()
	seed := flagSeed()
	rng = rand.New(rand.NewSource(seed))

	if *pFile == "" {
		log.Println("Doing nothing, no file specified")
		os.Exit(0)
		return
	}
	file, err := os.Create(*pFile)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Fatalf("Error closing file: %v", err)
		}
	}(file)

	bufOutput := bufio.NewWriter(file)
	writer := csv.NewWriter(bufOutput)
	if err = writer.Write([]string{"id", "date", "amount"}); err != nil {
		log.Fatalf("Error writing header to file: %v", err)
	}

	minDate, maxDate := flagDates()
	minAmount, maxAmount := flagAmounts()
	for i := uint(0); i < flagGenerate(); i++ {
		txnId := fmt.Sprintf("%d", i)
		txnDate := randomDateBetween(minDate, maxDate)
		txnAmount := randomAmountBetween(minAmount, maxAmount)
		err = writer.Write([]string{txnId, txnDate.Format(dayMonthLayout), fmt.Sprintf("%s%.2f", randomDebitOrCredit(), txnAmount)})
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}
	}

	err = bufOutput.Flush()
	if err != nil {
		log.Fatalf("Error flushing buffer: %v", err)
	}
	log.Printf("Wrote %d transactions to %s", flagGenerate(), *pFile)
}
