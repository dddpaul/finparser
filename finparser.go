package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const df = "02.01.2006"

type Commodity struct {
	desc  string
	price int
}

type Purchase struct {
	date      time.Time
	commodity *Commodity
}

func (purchase Purchase) toArray() []string {
	return []string{
		purchase.date.Format(df),
		purchase.commodity.desc,
		strconv.Itoa(purchase.commodity.price),
	}
}

type Purchases []*Purchase

func (purchases Purchases) toCsv() [][]string {
	var csv [][]string
	for _, purchase := range purchases {
		csv = append(csv, purchase.toArray())
	}
	return csv
}

var re1 *regexp.Regexp
var re2 *regexp.Regexp

func init() {
	var err error
	re1, err = regexp.Compile("^(\\d+\\+)+(\\d+)$")
	panicIfNotNil(err)
	re2, err = regexp.Compile("^[^\\d]\\d+=(\\d+)$")
	panicIfNotNil(err)
}

func panicIfNotNil(err error) {
	if err != nil {
		panic(err)
	}
}

func isEmpty(records []string) bool {
	for _, token := range records {
		if len(token) > 0 {
			return false
		}
	}
	return true
}

// Parse strings like "123+456+789" or "$5=338" and return sum in roubles
func parseAndSum(s string) (int, error) {
	var sum int
	if re1.MatchString(s) {
		strItems := strings.Split(s, "+")
		for _, strItem := range strItems {
			item, err := strconv.Atoi(strItem)
			if err != nil {
				return 0, err
			}
			sum += item
		}
	} else if re2.MatchString(s) {
		strItems := strings.Split(s, "=")
		item, err := strconv.Atoi(strItems[1])
		if err != nil {
			return 0, err
		}
		sum = item
	} else {
		return 0, fmt.Errorf("Invalid string sum value: %s", s)
	}
	return sum, nil
}

func newCommodity(s string) (*Commodity, error) {
	tokens := strings.Split(s, "(")
	if len(tokens) != 2 {
		return nil, fmt.Errorf("Can't parse: %s", s)
	}
	desc := strings.Trim(tokens[0], " ")
	strPrice := strings.TrimRight(tokens[1], ")")

	isDigit, err := regexp.MatchString("^\\d+$", strPrice)
	if err != nil {
		return nil, err
	}

	isSum, err := regexp.MatchString("^(\\d+\\+)+\\d+$", strPrice)
	if err != nil {
		return nil, err
	}

	var price int
	if isDigit {
		price, err = strconv.Atoi(strPrice)
	} else if isSum {
		price, err = parseAndSum(strPrice)
	}
	if err != nil {
		return nil, err
	}

	return &Commodity{desc, price}, nil
}

func getPurchases(records [][]string) (Purchases, []error) {
	var purchases []*Purchase
	var errors []error
	for i, record := range records {
		if i == 0 {
			continue
		}
		if isEmpty(record) {
			continue
		}

		// First field of record is a date, but if it's not a date - it's ok
		date, err := time.Parse(df, record[0])
		if err != nil {
			continue
		}

		// Second field of record is commodity list in text format
		commodities := strings.Split(record[1], ",")
		for _, s := range commodities {
			commodity, err := newCommodity(s)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			purchase := &Purchase{
				date:      date,
				commodity: commodity,
			}
			purchases = append(purchases, purchase)
		}
	}
	return purchases, errors
}

func main() {
	if len(os.Args) < 3 {
		panic(fmt.Errorf("Usage: %s <input-file> <output-file>", os.Args[0]))
	}

	in, err := os.Open(os.Args[1])
	panicIfNotNil(err)

	r := csv.NewReader(bufio.NewReader(in))
	records, err := r.ReadAll()
	panicIfNotNil(err)

	purchases, errors := getPurchases(records)
	fmt.Printf("Records total: %d, purchases: %d, errors: %d\n", len(records), len(purchases), len(errors))

	out, err := os.Create(os.Args[2])
	panicIfNotNil(err)

	w := csv.NewWriter(bufio.NewWriter(out))
	w.WriteAll(purchases.toCsv())
}
