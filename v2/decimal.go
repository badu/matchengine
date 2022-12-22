package market

import (
	"github.com/shopspring/decimal"
)

type Decimal = decimal.Decimal

func NewDecimalValue(value int64) Decimal {
	return decimal.New(value, 0)
}

func NewZeroDecimal() Decimal {
	return decimal.New(0, 1)
}
