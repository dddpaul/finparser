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

	"github.com/soniah/evaler"
	cbr "gopkg.in/kolomiichenko/cbr-currency-go.v1"
)

const df = "02.01.2006"

type ParseError struct {
	s   string
	row int
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s, row: %d", e.s, e.row)
}

type Commodity struct {
	person   string
	category string
	name     string
	price    int
}

type Purchase struct {
	date      time.Time
	commodity *Commodity
}

func (p Purchase) toArray() []string {
	return []string{
		p.date.Format(df),
		p.commodity.person,
		p.commodity.category,
		p.commodity.name,
		strconv.Itoa(p.commodity.price),
	}
}

type Purchases []*Purchase

func (pp Purchases) toCsv() [][]string {
	var csv [][]string
	for _, purchase := range pp {
		csv = append(csv, purchase.toArray())
	}
	return csv
}

var re1 *regexp.Regexp
var re2 *regexp.Regexp

func init() {
	var err error
	re1, err = regexp.Compile("^\\d+$")
	panicIfNotNil(err)
	re2, err = regexp.Compile("^[^\\d]\\d+=(\\d+)$")
	panicIfNotNil(err)
	cbr.UpdateCurrencyRates()
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

// Input string formats:
// - "person/category - name" - it's all clear;
// - "person/category" - name=category;
// - "category - name" - person is empty;
// - "name" - person is empty, category=name.
// Returns person, category, name, error.
func parseDesc(s string) (string, string, string, error) {
	var person, category, name string
	items := strings.Split(s, " - ")
	if len(items) < 1 && len(items) > 2 {
		return "", "", "", fmt.Errorf("Invalid description format: %s", s)
	}

	// Get ["Mary", "School"] from "Mary/School"
	subItems := strings.FieldsFunc(items[0], func(r rune) bool {
		return r == '/' || r == '|'
	})
	if len(subItems) == 0 {
		return "", "", "", fmt.Errorf("Invalid person/category format: %s", items[0])
	}

	if len(subItems) >= 2 {
		person = subItems[0]
		category = subItems[1]
	} else {
		category = subItems[0]
	}

	if len(items) == 2 {
		name = items[1]
	} else {
		if len(subItems) > 1 {
			name = subItems[1]
		} else {
			name = subItems[0]
		}
	}

	return person, category, name, nil
}

// Parse strings like "123+456+789", "2*400", "$5=338" and return sum in roubles
func parseSum(s string) (int, error) {
	var sum int
	var err error
	if re1.MatchString(s) {
		if sum, err = strconv.Atoi(s); err != nil {
			return 0, err
		}
	} else if re2.MatchString(s) {
		strItems := strings.Split(s, "=")
		if sum, err = strconv.Atoi(strItems[1]); err != nil {
			return 0, err
		}
	} else {
		rat, err := evaler.Eval(s)
		if err != nil {
			return 0, err
		}
		sum = int(rat.Num().Int64())
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
	person, category, name, err := parseDesc(desc)
	if err != nil {
		return nil, err
	}
	price, err := parseSum(strPrice)
	if err != nil {
		return nil, err
	}
	return &Commodity{person, category, name, price}, nil
}

func getPurchases(records [][]string) (Purchases, []*ParseError) {
	var purchases []*Purchase
	var errors []*ParseError
	for row, record := range records {
		if row == 0 {
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
				errors = append(errors, &ParseError{err.Error(), row + 1})
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
	if len(errors) > 0 {
		fmt.Printf("Errors are: %s\n", errors)
	}

	out, err := os.Create(os.Args[2])
	panicIfNotNil(err)

	w := csv.NewWriter(bufio.NewWriter(out))
	w.WriteAll(purchases.toCsv())
	out.Close()
}
