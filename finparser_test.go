package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"strings"
	"regexp"
)

func TestIsEmpty(t *testing.T) {
	assert.True(t, isEmpty([]string{"", ""}))
	assert.True(t, isEmpty([]string{"", "", ""}))
	assert.False(t, isEmpty([]string{"", "", "1"}))
	assert.False(t, isEmpty([]string{"abc", "ab", ""}))
}

func TestTimeParse(t *testing.T) {
	_, err := time.Parse("02.01.2006", "15.05.2016")
	assert.Nil(t, err)
	_, err = time.Parse("02.01.2006", "")
	assert.NotNil(t, err)
	_, err = time.Parse("02.01.2006", "Итого")
	assert.NotNil(t, err)
}

func TestSplit(t *testing.T) {
	str := ""
	assert.Equal(t, 1, len(strings.Split(str, ",")))
	str = " "
	assert.Equal(t, 1, len(strings.Split(str, ",")))
	str = ","
	assert.Equal(t, 2, len(strings.Split(str, ",")))
	str = ",2"
	assert.Equal(t, 2, len(strings.Split(str, ",")))
	str = "1,"
	assert.Equal(t, 2, len(strings.Split(str, ",")))
	str = "1,2"
	assert.Equal(t, 2, len(strings.Split(str, ",")))
	str = "Cat food(475+345), kid's hat(386), beer and apples(641+950)"
	assert.Equal(t, 3, len(strings.Split(str, ",")))
	str = "beer"
	assert.Equal(t, 1, len(strings.Split(str, " - ")))
}

func TestMatch(t *testing.T) {
	matched, _ := regexp.MatchString("^\\d+$", "")
	assert.False(t, matched)
	matched, _ = regexp.MatchString("^\\d+$", "123")
	assert.True(t, matched)
	matched, _ = regexp.MatchString("^\\d+$", "123+456")
	assert.False(t, matched)
	matched, _ = regexp.MatchString("^(\\d+\\+)+\\d+$", "123+456")
	assert.True(t, matched)
	matched, _ = regexp.MatchString("^(\\d+\\+)+\\d+$", "123+456+")
	assert.False(t, matched)
}

func TestSum(t *testing.T) {
	sum, err := parseSum("")
	assert.NotNil(t, err)

	sum, err = parseSum("123")
	assert.NotNil(t, err)

	sum, err = parseSum("123+")
	assert.NotNil(t, err)

	sum, err = parseSum("123+456")
	assert.Nil(t, err)
	assert.Equal(t, 579, sum)

	sum, err = parseSum("123+456+")
	assert.NotNil(t, err)

	sum, err = parseSum("123+456+1")
	assert.Nil(t, err)
	assert.Equal(t, 580, sum)

	sum, err = parseSum("$5=338")
	assert.Nil(t, err)
	assert.Equal(t, 338, sum)

	sum, err = parseSum("$17=1144")
	assert.Nil(t, err)
	assert.Equal(t, 1144, sum)

	sum, err = parseSum("2x500")
	assert.Nil(t, err)
	assert.Equal(t, 1000, sum)
}

func TestDesc(t *testing.T) {
	person, category, name, err := parseDesc("")
	assert.NotNil(t, err)

	person, category, name, err = parseDesc("Продукты")
	assert.Nil(t, err)
	assert.Equal(t, "", person)
	assert.Equal(t, "Продукты", category)
	assert.Equal(t, "Продукты", name)

	person, category, name, err = parseDesc("Продукты - Глобус")
	assert.Nil(t, err)
	assert.Equal(t, "", person)
	assert.Equal(t, "Продукты", category)
	assert.Equal(t, "Глобус", name)

	person, category, name, err = parseDesc("Кошка - витамины")
	assert.Nil(t, err)
	assert.Equal(t, "", person)
	assert.Equal(t, "Кошка", category)
	assert.Equal(t, "витамины", name)

	person, category, name, err = parseDesc("Маша|обувь - кроссовки")
	assert.Nil(t, err)
	assert.Equal(t, "Маша", person)
	assert.Equal(t, "обувь", category)
	assert.Equal(t, "кроссовки", name)

	person, category, name, err = parseDesc("Маша|обувь")
	assert.Nil(t, err)
	assert.Equal(t, "Маша", person)
	assert.Equal(t, "обувь", category)
	assert.Equal(t, "обувь", name)

	// invalid input
	person, category, name, err = parseDesc("пиво -раки")
	assert.Nil(t, err)
	assert.Equal(t, "", person)
	assert.Equal(t, "пиво -раки", category)
	assert.Equal(t, "пиво -раки", name)
}

func TestNewCommodity(t *testing.T) {
	_, err := newCommodity("")
	assert.NotNil(t, err)

	c, err := newCommodity("Cat's food (123)")
	assert.Nil(t, err)
	assert.Equal(t, 123, c.price)

	c, err = newCommodity("Food - cat's food and chocolate(123+456)")
	assert.Nil(t, err)
	assert.Equal(t, 579, c.price)

	c, err = newCommodity("Mary/food - chocolate with nuts and some juice (123+456+200)")
	assert.Nil(t, err)
	assert.Equal(t, 779, c.price)
}
