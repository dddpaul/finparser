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
	categories Categories
	name       string
	price      int
}

type Purchase struct {
	date      time.Time
	commodity *Commodity
}

func (p Purchase) toArray() []string {
	return []string{
		p.date.Format(df),
		strings.Join(p.commodity.categories, "|"),
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

type Categories []string

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
func parseSum(s string) (int, error) {
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

// Parse strings like "Продукты/Глобус" or "Кошка - витамины" or "Маша|обувь - кроссовки" or "пиво"
// and return list of categories and commodity name
func parseDesc(s string) (Categories, string, error) {
	items := strings.Split(s, " - ")
	if len(items) < 1 && len(items) > 2 {
		return nil, "", fmt.Errorf("Invalid description format: %s", s)
	}
	categories := strings.FieldsFunc(items[0], func(r rune) bool {
		return r == '/' || r == '|'
	})
	if len(categories) == 0 {
		return nil, "", fmt.Errorf("Invalid categories format: %s", items[0])
	}
	var name string
	if len(items) == 2 {
		name = items[1]
	} else {
		name = categories[0]
	}
	return categories, name, nil
}

func newCommodity(s string) (*Commodity, error) {
	tokens := strings.Split(s, "(")
	if len(tokens) != 2 {
		return nil, fmt.Errorf("Can't parse: %s", s)
	}
	desc := strings.Trim(tokens[0], " ")
	strPrice := strings.TrimRight(tokens[1], ")")
	categories, name, err := parseDesc(desc)

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
		price, err = parseSum(strPrice)
	}
	if err != nil {
		return nil, err
	}

	return &Commodity{categories, name, price}, nil
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
