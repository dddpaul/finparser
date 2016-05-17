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
}

func TestDesc(t *testing.T) {
	categories, name, err := parseDesc("")
	assert.NotNil(t, err)

	categories, name, err = parseDesc("Продукты/Глобус")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(categories))
	assert.Equal(t, "Продукты", categories[0])
	assert.Equal(t, "Глобус", categories[1])
	assert.Equal(t, "Продукты", name)

	categories, name, err = parseDesc("Кошка - витамины")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(categories))
	assert.Equal(t, "Кошка", categories[0])
	assert.Equal(t, "витамины", name)

	categories, name, err = parseDesc("Маша|обувь - кроссовки")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(categories))
	assert.Equal(t, "Маша", categories[0])
	assert.Equal(t, "обувь", categories[1])
	assert.Equal(t, "кроссовки", name)

	categories, name, err = parseDesc("пиво")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(categories))
	assert.Equal(t, "пиво", name)

	categories, name, err = parseDesc("пиво -раки")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(categories))
	assert.Equal(t, "пиво -раки", name)
}

func TestNewCommodity(t *testing.T) {
	_, err := newCommodity("")
	assert.NotNil(t, err)

	purchase, err := newCommodity("Cat's food (123)")
	assert.Nil(t, err)
	assert.Equal(t, "Cat's food", purchase.name)
	assert.Equal(t, 1, len(purchase.categories))
	assert.Equal(t, "Cat's food", purchase.categories[0])
	assert.Equal(t, 123, purchase.price)

	purchase, err = newCommodity("Food - cat's food and chocolate(123+456)")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(purchase.categories))
	assert.Equal(t, "Food", purchase.categories[0])
	assert.Equal(t, "cat's food and chocolate", purchase.name)
	assert.Equal(t, 579, purchase.price)

	purchase, err = newCommodity("Food/alcohol - chocolate with nuts and some beer (123+456+200)")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(purchase.categories))
	assert.Equal(t, "Food", purchase.categories[0])
	assert.Equal(t, "alcohol", purchase.categories[1])
	assert.Equal(t, "chocolate with nuts and some beer", purchase.name)
	assert.Equal(t, 779, purchase.price)
}
