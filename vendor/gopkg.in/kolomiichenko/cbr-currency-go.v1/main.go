package cbr

import (
	"encoding/xml"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html/charset"
)

// Original URL - http://www.cbr.ru/scripts/XML_daily.asp
// Unofficial mirror - https://www.cbr-xml-daily.ru/daily.xml, but without past dates
const URL = "http://www.cbr.ru/scripts/XML_daily.asp"

// date_req = Date of query (dd/mm/yyyy)
const DF = "02/01/2006"

type xmlResult struct {
	ValCurs xml.Name `xml:"ValCurs"`
	Date    string   `xml:"Date,attr"`
	Name    string   `xml:"name,attr"`
	Valute  []valute `xml:"Valute"`
}

type valute struct {
	ID       string  `xml:"ID,attr"`
	NumCode  int64   `xml:"NumCode"`
	CharCode string  `xml:"CharCode"`
	Nominal  float64 `xml:"Nominal"`
	Name     string  `xml:"Name"`
	Value    string  `xml:"Value"`
}

type currencyRate struct {
	ID      string
	NumCode int64
	ISOCode string
	Name    string
	Value   float64
}

var (
	currencyRates map[string]currencyRate
	mu            sync.Mutex
)

func init() {
	UpdateCurrencyRates()
	go doEvery(1*time.Hour, UpdateCurrencyRates)
}

func doEvery(d time.Duration, f func()) {
	for range time.Tick(d) {
		f()
	}
}

// GetCurrencyRates Return cached map of rates
func GetCurrencyRates() map[string]currencyRate {
	return currencyRates
}

// UpdateCurrencyRates Sync today currency rates with CBR
func UpdateCurrencyRates() {
	mu.Lock()
	currencyRates = FetchCurrencyRates(time.Time{})
	defer mu.Unlock()
}

// FetchCurrencyRates Fetch map of rates for specified date
func FetchCurrencyRates(d time.Time) map[string]currencyRate {
	url := URL
	if !d.IsZero() {
		url = url + "?date_req=" + d.Format(DF)
	}
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error of get currency: %v", err.Error())
		return nil
	}

	var data xmlResult

	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&data)
	if err != nil {
		log.Printf("error: %v", err)
		return nil
	}

	rates := make(map[string]currencyRate)

	for _, el := range data.Valute {
		value, _ := strconv.ParseFloat(strings.Replace(el.Value, ",", ".", -1), 64)

		rates[el.CharCode] = currencyRate{
			ID:      el.ID,
			NumCode: el.NumCode,
			ISOCode: el.CharCode,
			Name:    el.Name,
			Value:   value / el.Nominal,
		}
	}

	return rates
}
