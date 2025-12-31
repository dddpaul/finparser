package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dddpaul/cbr-currency-go"
	"github.com/soniah/evaler"
)

const DEFAULT_PERSON = "Общие"

var CATEGORY_REPLACES = map[string]string{
	"автобус":    "транспорт",
	"трамвай":    "транспорт",
	"троллейбус": "транспорт",
	"маршрутка":  "транспорт",
	"метро":      "транспорт",
	"электричка": "транспорт",
	"такси":      "транспорт",
	"интернет":   "связь",
}

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
	var c [][]string
	for _, purchase := range pp {
		c = append(c, purchase.toArray())
	}
	return c
}

var (
	l               *log.Logger
	df              string
	re1, re2, re3   *regexp.Regexp
	currencySymbols = map[string]string{"$": "USD", "€": "EUR", "Br": "BYN", "֏": "AMD"}
)

func init() {
	var err error
	re1, err = regexp.Compile("^\\d+$")
	panicIfNotNil(err)
	re2, err = regexp.Compile("^([$€]|Br|֏)\\d+(\\.\\d+)*=(\\d+)$")
	panicIfNotNil(err)
	re3, err = regexp.Compile("^([$€]|Br|֏)(\\d+)$")
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
		return "", "", "", fmt.Errorf("invalid description format: %s", s)
	}

	// Get ["Mary", "School"] from "Mary/School"
	subItems := strings.FieldsFunc(items[0], func(r rune) bool {
		return r == '/' || r == '|'
	})
	if len(subItems) == 0 {
		return "", "", "", fmt.Errorf("invalid person/category format: %s", items[0])
	}

	if len(subItems) >= 2 {
		person = strings.TrimSpace(subItems[0])
		category = strings.TrimSpace(subItems[1])
	} else {
		person = DEFAULT_PERSON
		category = strings.TrimSpace(subItems[0])
	}

	if len(items) == 2 {
		name = strings.TrimSpace(items[1])
	} else {
		if len(subItems) > 1 {
			name = strings.TrimSpace(subItems[1])
		} else {
			name = strings.TrimSpace(subItems[0])
		}
	}

	person = strings.ToLower(person)
	category = strings.ToLower(category)
	name = strings.ToLower(name)

	if v, ok := CATEGORY_REPLACES[category]; ok {
		category = v
	}

	return person, category, name, nil
}

// Parse strings like "123+456+789", "2*400", "$5=338" or "€17" and return sum in roubles
func parsePriceExpr(s string, date time.Time) (int, error) {
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
	} else if re3.MatchString(s) {
		if tokens := re3.FindStringSubmatch(s); tokens != nil {
			if sum, err = strconv.Atoi(tokens[2]); err != nil {
				return 0, err
			}
			code := currencySymbols[tokens[1]]
			sum = int(math.Round(float64(sum) * getCurrencyRate(code, date)))
			return sum, nil
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

func getCurrencyRate(code string, d time.Time) float64 {
	if d.IsZero() {
		return cbr.GetCurrencyRates()[code].Value
	} else {
		rates, err := cbr.FetchCurrencyRates(d)
		if err != nil {
			return 0
		}
		return rates[code].Value
	}
}

func newCommodity(s string, date time.Time) (*Commodity, error) {
	tokens := strings.Split(s, "(")
	if len(tokens) < 2 {
		return nil, fmt.Errorf("can't parse: %s", s)
	}
	desc := strings.TrimSpace(tokens[0])
	strPrice := strings.TrimRight(strings.TrimSpace(tokens[1]), ")")
	person, category, name, err := parseDesc(desc)
	if err != nil {
		return nil, err
	}
	price, err := parsePriceExpr(strPrice, date)
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
			commodity, err := newCommodity(s, date)
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
	flag.StringVar(&df, "df", "02.01.2006", "Golang date format")
	flag.Parse()

	l = log.New(os.Stderr, "", log.LstdFlags)

	r := csv.NewReader(bufio.NewReader(os.Stdin))
	records, err := r.ReadAll()
	panicIfNotNil(err)

	purchases, errors := getPurchases(records)
	l.Printf("Records total: %d, purchases: %d, errors: %d\n", len(records), len(purchases), len(errors))
	if len(errors) > 0 {
		l.Printf("Errors are: %s\n", errors)
	}

	w := csv.NewWriter(bufio.NewWriter(os.Stdout))
	panicIfNotNil(w.WriteAll(purchases.toCsv()))
	panicIfNotNil(os.Stdout.Close())
}
