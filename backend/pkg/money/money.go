package money

import (
	"fmt"
	"math/big"
)

// Currency represents a currency code
type Currency string

const (
	ILS Currency = "ILS"
	USD Currency = "USD"
	EUR Currency = "EUR"
)

// Money represents a monetary amount using integer minor units
type Money struct {
	Amount   int64   // Amount in minor units (e.g., cents/agarot)
	Currency Currency
}

// NewMoney creates a new Money instance
func NewMoney(amount int64, currency Currency) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

// NewFromFloat creates Money from float (for compatibility, not recommended)
func NewFromFloat(amount float64, currency Currency) Money {
	// Convert to minor units (multiply by 100 for 2 decimal places)
	minorUnits := int64(amount * 100)
	return Money{
		Amount:   minorUnits,
		Currency: currency,
	}
}

// ToFloat converts Money to float (for display purposes)
func (m Money) ToFloat() float64 {
	return float64(m.Amount) / 100
}

// ToDecimal converts Money to decimal string
func (m Money) ToDecimal() string {
	return fmt.Sprintf("%.2f", m.ToFloat())
}

// Add adds two Money instances
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.Currency, other.Currency)
	}
	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}, nil
}

// Sub subtracts other from m
func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.Currency, other.Currency)
	}
	return Money{
		Amount:   m.Amount - other.Amount,
		Currency: m.Currency,
	}, nil
}

// Mul multiplies Money by a factor
func (m Money) Mul(factor float64) Money {
	result := int64(float64(m.Amount) * factor)
	return Money{
		Amount:   result,
		Currency: m.Currency,
	}
}

// Div divides Money by a divisor
func (m Money) Div(divisor float64) Money {
	result := int64(float64(m.Amount) / divisor)
	return Money{
		Amount:   result,
		Currency: m.Currency,
	}
}

// IsZero checks if the amount is zero
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsNegative checks if the amount is negative
func (m Money) IsNegative() bool {
	return m.Amount < 0
}

// IsPositive checks if the amount is positive
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// Compare compares two Money instances
// Returns: -1 if m < other, 0 if m == other, 1 if m > other
func (m Money) Compare(other Money) (int, error) {
	if m.Currency != other.Currency {
		return 0, fmt.Errorf("cannot compare different currencies: %s and %s", m.Currency, other.Currency)
	}
	if m.Amount < other.Amount {
		return -1, nil
	} else if m.Amount > other.Amount {
		return 1, nil
	}
	return 0, nil
}

// Equals checks if two Money instances are equal
func (m Money) Equals(other Money) (bool, error) {
	cmp, err := m.Compare(other)
	if err != nil {
		return false, err
	}
	return cmp == 0, nil
}

// GreaterThan checks if m > other
func (m Money) GreaterThan(other Money) (bool, error) {
	cmp, err := m.Compare(other)
	if err != nil {
		return false, err
	}
	return cmp == 1, nil
}

// LessThan checks if m < other
func (m Money) LessThan(other Money) (bool, error) {
	cmp, err := m.Compare(other)
	if err != nil {
		return false, err
	}
	return cmp == -1, nil
}

// ToBigInt converts Amount to big.Int for precise calculations
func (m Money) ToBigInt() *big.Int {
	return big.NewInt(m.Amount)
}

// FromBigInt creates Money from big.Int
func FromBigInt(amount *big.Int, currency Currency) Money {
	return Money{
		Amount:   amount.Int64(),
		Currency: currency,
	}
}

// Format formats the money for display with currency symbol
func (m Money) Format() string {
	symbol := ""
	switch m.Currency {
	case ILS:
		symbol = "₪"
	case USD:
		symbol = "$"
	case EUR:
		symbol = "€"
	}
	return fmt.Sprintf("%s%.2f", symbol, m.ToFloat())
}

// FormatWithCode formats the money with currency code
func (m Money) FormatWithCode() string {
	return fmt.Sprintf("%.2f %s", m.ToFloat(), m.Currency)
}
