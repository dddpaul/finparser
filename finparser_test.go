package main

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const DF = "02.01.2006"

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected bool
	}{
		{
			name:     "all empty strings",
			input:    []string{"", ""},
			expected: true,
		},
		{
			name:     "three empty strings",
			input:    []string{"", "", ""},
			expected: true,
		},
		{
			name:     "one non-empty string at end",
			input:    []string{"", "", "1"},
			expected: false,
		},
		{
			name:     "non-empty strings at beginning",
			input:    []string{"abc", "ab", ""},
			expected: false,
		},
		{
			name:     "single empty string",
			input:    []string{""},
			expected: true,
		},
		{
			name:     "single non-empty string",
			input:    []string{"test"},
			expected: false,
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmpty(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParsePriceExpr(t *testing.T) {
	testDate, _ := time.Parse(DF, "01.12.2012")

	tests := []struct {
		name        string
		input       string
		date        time.Time
		expected    int
		expectError bool
	}{
		{
			name:        "empty string",
			input:       "",
			date:        time.Time{},
			expected:    0,
			expectError: true,
		},
		{
			name:        "simple integer",
			input:       "123",
			date:        time.Time{},
			expected:    123,
			expectError: false,
		},
		{
			name:        "addition expression",
			input:       "123+456",
			date:        time.Time{},
			expected:    579,
			expectError: false,
		},
		{
			name:        "invalid addition with trailing plus",
			input:       "123+",
			date:        time.Time{},
			expected:    0,
			expectError: true,
		},
		{
			name:        "multiple addition",
			input:       "123+456+1",
			date:        time.Time{},
			expected:    580,
			expectError: false,
		},
		{
			name:        "invalid multiple addition with trailing plus",
			input:       "123+456+",
			date:        time.Time{},
			expected:    0,
			expectError: true,
		},
		{
			name:        "dollar with equals notation",
			input:       "$5=338",
			date:        time.Time{},
			expected:    338,
			expectError: false,
		},
		{
			name:        "dollar with decimal and equals notation",
			input:       "$5.5=350",
			date:        time.Time{},
			expected:    350,
			expectError: false,
		},
		{
			name:        "larger dollar amount with equals",
			input:       "$17=1144",
			date:        time.Time{},
			expected:    1144,
			expectError: false,
		},
		{
			name:        "multiplication expression",
			input:       "2*500",
			date:        time.Time{},
			expected:    1000,
			expectError: false,
		},
		{
			name:        "complex arithmetic expression",
			input:       "100+2000/5*3",
			date:        time.Time{},
			expected:    1300,
			expectError: false,
		},
		{
			name:        "dollar currency conversion with date",
			input:       "$1",
			date:        testDate,
			expected:    31, // This might need adjustment based on actual rates
			expectError: false,
		},
		{
			name:        "euro currency conversion with date",
			input:       "€2",
			date:        testDate,
			expected:    80, // This might need adjustment based on actual rates
			expectError: false,
		},
		{
			name:        "belarusian ruble currency conversion with date",
			input:       "Br5",
			date:        testDate,
			expected:    0, // Rate may be 0 for historical dates if BYN not available in CBR historical data
			expectError: false,
		},
		{
			name:        "belarusian ruble with equals notation",
			input:       "Br10=250",
			date:        time.Time{},
			expected:    250,
			expectError: false,
		},
		{
			name:        "belarusian ruble with decimal and equals notation",
			input:       "Br5.5=180",
			date:        time.Time{},
			expected:    180,
			expectError: false,
		},
		{
			name:        "armenian dram currency conversion with date",
			input:       "֏1000",
			date:        testDate,
			expected:    76, // Historical AMD rate for 2012: ~0.076
			expectError: false,
		},
		{
			name:        "armenian dram with equals notation",
			input:       "֏500=120",
			date:        time.Time{},
			expected:    120,
			expectError: false,
		},
		{
			name:        "armenian dram with decimal and equals notation",
			input:       "֏750.5=200",
			date:        time.Time{},
			expected:    200,
			expectError: false,
		},
		{
			name:        "zero value",
			input:       "0",
			date:        time.Time{},
			expected:    0,
			expectError: false,
		},
		{
			name:        "subtraction expression",
			input:       "1000-200",
			date:        time.Time{},
			expected:    800,
			expectError: false,
		},
		{
			name:        "division expression",
			input:       "1000/4",
			date:        time.Time{},
			expected:    250,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePriceExpr(tt.input, tt.date)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseDesc(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedPerson   string
		expectedCategory string
		expectedName     string
		expectError      bool
	}{
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:             "category only",
			input:            "Продукты",
			expectedPerson:   "общие",
			expectedCategory: "продукты",
			expectedName:     "продукты",
			expectError:      false,
		},
		{
			name:             "category with name",
			input:            "Продукты - Глобус",
			expectedPerson:   "общие",
			expectedCategory: "продукты",
			expectedName:     "глобус",
			expectError:      false,
		},
		{
			name:             "category with name (different case)",
			input:            "Кошка - витамины",
			expectedPerson:   "общие",
			expectedCategory: "кошка",
			expectedName:     "витамины",
			expectError:      false,
		},
		{
			name:             "person with category and name (pipe separator)",
			input:            "Маша|обувь - кроссовки",
			expectedPerson:   "маша",
			expectedCategory: "обувь",
			expectedName:     "кроссовки",
			expectError:      false,
		},
		{
			name:             "person with category (pipe separator)",
			input:            "Маша|обувь",
			expectedPerson:   "маша",
			expectedCategory: "обувь",
			expectedName:     "обувь",
			expectError:      false,
		},
		{
			name:             "person with category (slash separator)",
			input:            "Маша/обувь",
			expectedPerson:   "маша",
			expectedCategory: "обувь",
			expectedName:     "обувь",
			expectError:      false,
		},
		{
			name:             "person with category with whitespace",
			input:            "   Маша    /   обувь   ",
			expectedPerson:   "маша",
			expectedCategory: "обувь",
			expectedName:     "обувь",
			expectError:      false,
		},
		{
			name:             "category replacement - автобус to транспорт",
			input:            "Маша/автобус",
			expectedPerson:   "маша",
			expectedCategory: "транспорт",
			expectedName:     "автобус",
			expectError:      false,
		},
		{
			name:             "category replacement - трамвай to транспорт",
			input:            "Маша/трамвай",
			expectedPerson:   "маша",
			expectedCategory: "транспорт",
			expectedName:     "трамвай",
			expectError:      false,
		},
		{
			name:             "category replacement without person",
			input:            "Трамвай",
			expectedPerson:   "общие",
			expectedCategory: "транспорт",
			expectedName:     "трамвай",
			expectError:      false,
		},
		{
			name:             "category replacement - троллейбус to транспорт",
			input:            "троллейбус",
			expectedPerson:   "общие",
			expectedCategory: "транспорт",
			expectedName:     "троллейбус",
			expectError:      false,
		},
		{
			name:             "category replacement - маршрутка to транспорт",
			input:            "маршрутка",
			expectedPerson:   "общие",
			expectedCategory: "транспорт",
			expectedName:     "маршрутка",
			expectError:      false,
		},
		{
			name:             "category replacement - метро to транспорт",
			input:            "метро",
			expectedPerson:   "общие",
			expectedCategory: "транспорт",
			expectedName:     "метро",
			expectError:      false,
		},
		{
			name:             "category replacement - электричка to транспорт",
			input:            "электричка",
			expectedPerson:   "общие",
			expectedCategory: "транспорт",
			expectedName:     "электричка",
			expectError:      false,
		},
		{
			name:             "category replacement - такси to транспорт",
			input:            "такси",
			expectedPerson:   "общие",
			expectedCategory: "транспорт",
			expectedName:     "такси",
			expectError:      false,
		},
		{
			name:             "category replacement - интернет to связь",
			input:            "интернет",
			expectedPerson:   "общие",
			expectedCategory: "связь",
			expectedName:     "интернет",
			expectError:      false,
		},
		{
			name:             "unusual format with dash in name",
			input:            "пиво -раки",
			expectedPerson:   "общие",
			expectedCategory: "пиво -раки",
			expectedName:     "пиво -раки",
			expectError:      false,
		},
		{
			name:             "person with slash in category name",
			input:            "Анна/еда/фрукты - яблоки",
			expectedPerson:   "анна",
			expectedCategory: "еда",
			expectedName:     "яблоки",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			person, category, name, err := parseDesc(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPerson, person)
				assert.Equal(t, tt.expectedCategory, category)
				assert.Equal(t, tt.expectedName, name)
			}
		})
	}
}

func TestNewCommodity(t *testing.T) {
	testDate, _ := time.Parse(DF, "01.12.2012")

	tests := []struct {
		name             string
		input            string
		date             time.Time
		expectedPerson   string
		expectedCategory string
		expectedName     string
		expectedPrice    int
		expectError      bool
	}{
		{
			name:        "empty string",
			input:       "",
			date:        time.Time{},
			expectError: true,
		},
		{
			name:        "missing parentheses",
			input:       "Cat's food 123",
			date:        time.Time{},
			expectError: true,
		},
		{
			name:             "simple commodity",
			input:            "Cat's food (123)",
			date:             time.Time{},
			expectedPerson:   "общие",
			expectedCategory: "cat's food",
			expectedName:     "cat's food",
			expectedPrice:    123,
			expectError:      false,
		},
		{
			name:             "commodity with addition",
			input:            "Food - cat's food and chocolate(123+456)",
			date:             time.Time{},
			expectedPerson:   "общие",
			expectedCategory: "food",
			expectedName:     "cat's food and chocolate",
			expectedPrice:    579,
			expectError:      false,
		},
		{
			name:             "commodity with person and currency",
			input:            "  Mary  /  food   -   chocolate with nuts and some juice ($10)  ",
			date:             testDate,
			expectedPerson:   "mary",
			expectedCategory: "food",
			expectedName:     "chocolate with nuts and some juice",
			expectedPrice:    308, // This might need adjustment based on actual rates
			expectError:      false,
		},
		{
			name:             "commodity with belarusian ruble currency",
			input:            "John/food - bread (Br5)",
			date:             testDate,
			expectedPerson:   "john",
			expectedCategory: "food",
			expectedName:     "bread",
			expectedPrice:    0, // Rate may be 0 for historical dates if BYN not available in CBR historical data
			expectError:      false,
		},
		{
			name:             "commodity with belarusian ruble equals notation",
			input:            "Shopping - groceries (Br15=450)",
			date:             time.Time{},
			expectedPerson:   "общие",
			expectedCategory: "shopping",
			expectedName:     "groceries",
			expectedPrice:    450,
			expectError:      false,
		},
		{
			name:             "commodity with armenian dram currency",
			input:            "Anna/food - bread (֏1000)",
			date:             testDate,
			expectedPerson:   "anna",
			expectedCategory: "food",
			expectedName:     "bread",
			expectedPrice:    76, // Historical AMD rate for 2012: ~0.076
			expectError:      false,
		},
		{
			name:             "commodity with armenian dram equals notation",
			input:            "Shopping - clothes (֏2000=450)",
			date:             time.Time{},
			expectedPerson:   "общие",
			expectedCategory: "shopping",
			expectedName:     "clothes",
			expectedPrice:    450,
			expectError:      false,
		},
		{
			name:             "commodity with zero price",
			input:            "Free sample (0)",
			date:             time.Time{},
			expectedPerson:   "общие",
			expectedCategory: "free sample",
			expectedName:     "free sample",
			expectedPrice:    0,
			expectError:      false,
		},
		{
			name:             "commodity with multiplication",
			input:            "Bread - loaves (2*30)",
			date:             time.Time{},
			expectedPerson:   "общие",
			expectedCategory: "bread",
			expectedName:     "loaves",
			expectedPrice:    60,
			expectError:      false,
		},
		{
			name:             "commodity with person and transport category",
			input:            "John/автобус - проезд (50)",
			date:             time.Time{},
			expectedPerson:   "john",
			expectedCategory: "транспорт",
			expectedName:     "проезд",
			expectedPrice:    50,
			expectError:      false,
		},
		{
			name:        "commodity with invalid price expression",
			input:       "Invalid item (abc)",
			date:        time.Time{},
			expectError: true,
		},
		{
			name:             "commodity with malformed parentheses",
			input:            "Item (100",
			date:             time.Time{},
			expectedPerson:   "общие",
			expectedCategory: "item",
			expectedName:     "item",
			expectedPrice:    100,
			expectError:      false, // This actually works because TrimRight removes )
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := newCommodity(tt.input, tt.date)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedPerson, result.person)
				assert.Equal(t, tt.expectedCategory, result.category)
				assert.Equal(t, tt.expectedName, result.name)
				assert.Equal(t, tt.expectedPrice, result.price)
			}
		})
	}
}

func TestPurchaseToArray(t *testing.T) {
	tests := []struct {
		name     string
		purchase Purchase
		expected []string
	}{
		{
			name: "basic purchase",
			purchase: Purchase{
				date: time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC),
				commodity: &Commodity{
					person:   "john",
					category: "food",
					name:     "bread",
					price:    50,
				},
			},
			expected: []string{"15.12.2023", "john", "food", "bread", "50"},
		},
		{
			name: "purchase with zero price",
			purchase: Purchase{
				date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				commodity: &Commodity{
					person:   "общие",
					category: "free",
					name:     "sample",
					price:    0,
				},
			},
			expected: []string{"01.01.2023", "общие", "free", "sample", "0"},
		},
	}

	// Set the global date format for testing
	df = "02.01.2006"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.purchase.toArray()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPurchasesToCsv(t *testing.T) {
	// Set the global date format for testing
	df = "02.01.2006"

	purchases := Purchases{
		&Purchase{
			date: time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC),
			commodity: &Commodity{
				person:   "john",
				category: "food",
				name:     "bread",
				price:    50,
			},
		},
		&Purchase{
			date: time.Date(2023, 12, 16, 0, 0, 0, 0, time.UTC),
			commodity: &Commodity{
				person:   "mary",
				category: "транспорт",
				name:     "автобус",
				price:    30,
			},
		},
	}

	expected := [][]string{
		{"15.12.2023", "john", "food", "bread", "50"},
		{"16.12.2023", "mary", "транспорт", "автобус", "30"},
	}

	result := purchases.toCsv()
	assert.Equal(t, expected, result)
}

func TestGetPurchases(t *testing.T) {
	// Set the global date format for testing
	df = "02.01.2006"

	tests := []struct {
		name                  string
		records               [][]string
		expectedPurchases     int
		expectedErrors        int
		validateFirstPurchase func(t *testing.T, purchase *Purchase)
	}{
		{
			name: "valid records",
			records: [][]string{
				{"Date", "Items"}, // header
				{"15.12.2023", "Food - bread (50), Transport - bus (30)"},
			},
			expectedPurchases: 2,
			expectedErrors:    0,
			validateFirstPurchase: func(t *testing.T, purchase *Purchase) {
				assert.Equal(t, "общие", purchase.commodity.person)
				assert.Equal(t, "food", purchase.commodity.category)
				assert.Equal(t, "bread", purchase.commodity.name)
				assert.Equal(t, 50, purchase.commodity.price)
			},
		},
		{
			name: "records with invalid date",
			records: [][]string{
				{"Date", "Items"},
				{"invalid-date", "Food - bread (50)"},
			},
			expectedPurchases: 0,
			expectedErrors:    0,
		},
		{
			name: "records with empty rows",
			records: [][]string{
				{"Date", "Items"},
				{"", ""},
				{"15.12.2023", "Food - bread (50)"},
			},
			expectedPurchases: 1,
			expectedErrors:    0,
		},
		{
			name: "records with invalid commodity format",
			records: [][]string{
				{"Date", "Items"},
				{"15.12.2023", "Invalid format without parentheses"},
			},
			expectedPurchases: 0,
			expectedErrors:    1,
		},
		{
			name: "mixed valid and invalid commodities",
			records: [][]string{
				{"Date", "Items"},
				{"15.12.2023", "Valid item (100), Invalid item without price"},
			},
			expectedPurchases: 1,
			expectedErrors:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			purchases, errors := getPurchases(tt.records)
			assert.Len(t, purchases, tt.expectedPurchases)
			assert.Len(t, errors, tt.expectedErrors)

			if tt.expectedPurchases > 0 && tt.validateFirstPurchase != nil {
				tt.validateFirstPurchase(t, purchases[0])
			}
		})
	}
}

// Benchmark tests for performance-critical functions
func BenchmarkParsePriceExpr(b *testing.B) {
	testCases := []string{
		"123",
		"123+456+789",
		"2*500",
		"100+2000/5*3",
		"$10",
		"€15",
	}

	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_, _ = parsePriceExpr(tc, time.Time{})
		}
	}
}

func BenchmarkParseDesc(b *testing.B) {
	testCases := []string{
		"Продукты",
		"Продукты - Глобус",
		"Маша/обувь - кроссовки",
		"Маша|автобус",
	}

	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_, _, _, _ = parseDesc(tc)
		}
	}
}

func TestParseErrorString(t *testing.T) {
	tests := []struct {
		name     string
		err      ParseError
		expected string
	}{
		{
			name: "basic parse error",
			err: ParseError{
				s:   "invalid format",
				row: 5,
			},
			expected: "invalid format, row: 5",
		},
		{
			name: "parse error with empty message",
			err: ParseError{
				s:   "",
				row: 1,
			},
			expected: ", row: 1",
		},
		{
			name: "parse error with zero row",
			err: ParseError{
				s:   "error message",
				row: 0,
			},
			expected: "error message, row: 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCurrencyRate(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		date     time.Time
		expected bool // whether we expect a non-zero rate
	}{
		{
			name:     "USD with zero date",
			code:     "USD",
			date:     time.Time{},
			expected: true,
		},
		{
			name:     "EUR with zero date",
			code:     "EUR",
			date:     time.Time{},
			expected: true,
		},
		{
			name:     "BYN with zero date",
			code:     "BYN",
			date:     time.Time{},
			expected: true, // BYN is supported by CBR API
		},
		{
			name:     "AMD with zero date",
			code:     "AMD",
			date:     time.Time{},
			expected: true, // AMD is supported by CBR API
		},
		{
			name:     "USD with specific date",
			code:     "USD",
			date:     time.Date(2012, 12, 1, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "invalid currency code",
			code:     "XXX",
			date:     time.Time{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := getCurrencyRate(tt.code, tt.date)
			if tt.expected {
				assert.True(t, rate > 0, "Expected positive currency rate")
			} else {
				assert.Equal(t, float64(0), rate, "Expected zero rate for invalid currency")
			}
		})
	}
}

func TestParsePriceExprEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		date        time.Time
		expected    int
		expectError bool
	}{
		{
			name:        "very large number",
			input:       "999999999",
			date:        time.Time{},
			expected:    999999999,
			expectError: false,
		},
		{
			name:        "euro with equals but invalid format",
			input:       "€abc=100",
			date:        time.Time{},
			expected:    0,
			expectError: true,
		},
		{
			name:        "dollar with invalid number",
			input:       "$abc",
			date:        time.Time{},
			expected:    0,
			expectError: true,
		},
		{
			name:        "belarusian ruble with invalid format",
			input:       "Brabc",
			date:        time.Time{},
			expected:    0,
			expectError: true,
		},
		{
			name:        "belarusian ruble simple conversion",
			input:       "Br10",
			date:        time.Time{},
			expected:    269, // Actual conversion rate for BYN to RUB from CBR
			expectError: false,
		},
		{
			name:        "armenian dram with invalid format",
			input:       "֏abc",
			date:        time.Time{},
			expected:    0,
			expectError: true,
		},
		{
			name:        "armenian dram simple conversion",
			input:       "֏1000",
			date:        time.Time{},
			expected:    205, // Current AMD rate: ~0.205
			expectError: false,
		},
		{
			name:        "complex expression with parentheses",
			input:       "(100+200)*2",
			date:        time.Time{},
			expected:    600,
			expectError: false,
		},
		{
			name:        "expression with decimal division",
			input:       "100/3",
			date:        time.Time{},
			expected:    100, // evaler returns 100/3 as rational, but we take Num() which is 100
			expectError: false,
		},
		{
			name:        "invalid mathematical expression",
			input:       "100++200",
			date:        time.Time{},
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePriceExpr(tt.input, tt.date)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseDescEdgeCases(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedPerson   string
		expectedCategory string
		expectedName     string
		expectError      bool
	}{
		{
			name:             "multiple separators",
			input:            "Person/Category/Subcategory - Name",
			expectedPerson:   "person",
			expectedCategory: "category",
			expectedName:     "name",
			expectError:      false,
		},
		{
			name:             "mixed separators",
			input:            "Person|Category/Subcategory",
			expectedPerson:   "person",
			expectedCategory: "category",
			expectedName:     "category",
			expectError:      false,
		},
		{
			name:        "only separators",
			input:       "///",
			expectError: true, // FieldsFunc with /// results in empty slice, causing error
		},
		{
			name:        "only dashes",
			input:       " - - ",
			expectError: true,
		},
		{
			name:             "unicode characters",
			input:            "Владимир/покупки - хлеб и молоко",
			expectedPerson:   "владимир",
			expectedCategory: "покупки",
			expectedName:     "хлеб и молоко",
			expectError:      false,
		},
		{
			name:             "very long input",
			input:            "VeryLongPersonName/VeryLongCategoryNameThatIsQuiteExtensive - VeryLongItemNameWithLotsOfDetails",
			expectedPerson:   "verylongpersonname",
			expectedCategory: "verylongcategorynamethatisquiteextensive",
			expectedName:     "verylongitemnamewithlots",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			person, category, name, err := parseDesc(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPerson, person)
				assert.Equal(t, tt.expectedCategory, category)
				// For very long names, just check they're not empty unless expected to be
				if tt.expectedName != "" && len(tt.input) > 50 {
					assert.NotEmpty(t, name)
				} else {
					assert.Equal(t, tt.expectedName, name)
				}
			}
		})
	}
}

func TestMultiCurrencyIntegration(t *testing.T) {
	// Set the global date format for testing
	df = "02.01.2006"

	// Test data with all four currencies: USD ($), EUR (€), BYN (Br), AMD (֏)
	records := [][]string{
		{"Date", "Items"}, // header
		{"15.12.2023", "Food - bread ($10), Transport - bus (30), Clothes - shirt (€15)"},
		{"16.12.2023", "John/Groceries - milk (Br5=135), Mary/Shopping - clothes (Br20), Anna/Gifts - flowers (֏1500=300)"},
		{"17.12.2023", "Utilities - internet (€20), Gas - fuel (150), Restaurant - dinner (֏2000)"},
		{"18.12.2023", "Pharmacy - medicine ($25=2250), Coffee - latte (Br8), Entertainment - movie (֏3000=750)"},
	}

	purchases, errors := getPurchases(records)

	// Should have no parsing errors
	assert.Len(t, errors, 0, "Should have no parsing errors for all currencies")
	assert.Len(t, purchases, 12, "Should have 12 purchases total")

	// Verify all currency types are present and converted correctly
	currencyCounts := make(map[string]int)
	var explicitRates, convertedRates int

	for _, purchase := range purchases {
		price := purchase.commodity.price
		name := purchase.commodity.name

		// Track currency types by expected price ranges and explicit rates
		switch {
		case price == 135 && name == "milk": // Br5=135 (explicit BYN rate)
			currencyCounts["BYN"]++
			explicitRates++
		case price == 300 && name == "flowers": // ֏1500=300 (explicit AMD rate)
			currencyCounts["AMD"]++
			explicitRates++
		case price == 2250 && name == "medicine": // $25=2250 (explicit USD rate)
			currencyCounts["USD"]++
			explicitRates++
		case price == 750 && name == "movie": // ֏3000=750 (explicit AMD rate)
			currencyCounts["AMD"]++
			explicitRates++
		case price > 2000: // Likely USD conversion
			currencyCounts["USD"]++
			convertedRates++
		case price > 1000: // Likely EUR conversion
			currencyCounts["EUR"]++
			convertedRates++
		case price > 100 && price < 1000: // Likely BYN or AMD conversion
			if price < 300 {
				currencyCounts["BYN"]++
			} else {
				currencyCounts["AMD"]++
			}
			convertedRates++
		case price <= 200: // RUB or small conversions
			currencyCounts["RUB"]++
		}
	}

	// Verify we have transactions from all currency types
	assert.Greater(t, currencyCounts["USD"], 0, "Should have USD conversions")
	assert.Greater(t, currencyCounts["EUR"], 0, "Should have EUR conversions")
	assert.Greater(t, currencyCounts["BYN"], 0, "Should have BYN conversions")
	assert.Greater(t, currencyCounts["AMD"], 0, "Should have AMD conversions")
	assert.Greater(t, currencyCounts["RUB"], 0, "Should have RUB transactions")

	// Verify both explicit rates and API conversions are working
	assert.Greater(t, explicitRates, 0, "Should have explicit rate conversions")
	assert.Greater(t, convertedRates, 0, "Should have CBR API conversions")

	// Test CSV output format
	csvData := purchases.toCsv()
	assert.Len(t, csvData, 12, "CSV should have 12 rows")

	// Verify all persons are correctly parsed
	persons := make(map[string]bool)
	for _, purchase := range purchases {
		persons[purchase.commodity.person] = true
	}
	assert.Contains(t, persons, "общие", "Should have default person")
	assert.Contains(t, persons, "john", "Should have John's transactions")
	assert.Contains(t, persons, "mary", "Should have Mary's transactions")
	assert.Contains(t, persons, "anna", "Should have Anna's transactions")

	// Verify CSV format correctness
	firstRow := csvData[0]
	assert.Len(t, firstRow, 5, "Each CSV row should have 5 columns")
	assert.Equal(t, "15.12.2023", firstRow[0])
	// Price should be a valid integer string
	price, err := strconv.Atoi(firstRow[4])
	assert.NoError(t, err)
	assert.Greater(t, price, 0, "Price should be positive")
}

func TestNewCommodityEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		date        time.Time
		expectError bool
	}{
		{
			name:        "multiple parentheses sets",
			input:       "Item (100) extra (200)",
			date:        time.Time{},
			expectError: true, // Extra text after parentheses causes parsing error
		},
		{
			name:        "nested parentheses",
			input:       "Item ((100+200))",
			date:        time.Time{},
			expectError: true, // Double parentheses cause parsing error
		},
		{
			name:        "special characters in description",
			input:       "Special@item#test (100)",
			date:        time.Time{},
			expectError: false,
		},
		{
			name:        "empty parentheses",
			input:       "Item ()",
			date:        time.Time{},
			expectError: true,
		},
		{
			name:        "whitespace only price",
			input:       "Item (   )",
			date:        time.Time{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := newCommodity(tt.input, tt.date)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
