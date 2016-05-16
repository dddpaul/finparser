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
	sum, err := parseAndSum("")
	assert.NotNil(t, err)
	sum, err = parseAndSum("123")

	assert.NotNil(t, err)
	sum, err = parseAndSum("123+")
	assert.NotNil(t, err)

	sum, err = parseAndSum("123+456")
	assert.Nil(t, err)
	assert.Equal(t, 579, sum)

	sum, err = parseAndSum("123+456+")
	assert.NotNil(t, err)

	sum, err = parseAndSum("123+456+1")
	assert.Nil(t, err)
	assert.Equal(t, 580, sum)

	sum, err = parseAndSum("$5=338")
	assert.Nil(t, err)
	assert.Equal(t, 338, sum)
}

func TestNewCommodity(t *testing.T) {
	_, err := newCommodity("")
	assert.NotNil(t, err)

	purchase, err := newCommodity("Cat's food (123)")
	assert.Nil(t, err)
	assert.Equal(t, "Cat's food", purchase.desc)
	assert.Equal(t, 123, purchase.price)

	purchase, err = newCommodity("Cat's food and chocolate(123+456)")
	assert.Nil(t, err)
	assert.Equal(t, "Cat's food and chocolate", purchase.desc)
	assert.Equal(t, 579, purchase.price)

	purchase, err = newCommodity("Cat's food and chocolate and some mushrooms (123+456+200)")
	assert.Nil(t, err)
	assert.Equal(t, "Cat's food and chocolate and some mushrooms", purchase.desc)
	assert.Equal(t, 779, purchase.price)
}
