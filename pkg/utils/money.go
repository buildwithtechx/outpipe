package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func FormatMinorAmount(amount int64, currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	decimals := 2

	if currency == "BHD" || currency == "JOD" || currency == "KWD" {
		decimals = 3
	}

	if currency == "CLP" || currency == "ISK" || currency == "JPY" || currency == "KRW" {
		decimals = 0
	}

	value := float64(amount) / math.Pow10(decimals)
	formatted := strconv.FormatFloat(value, 'f', decimals, 64)
	symbols := map[string]string{"EUR": "€", "GBP": "£", "NGN": "₦", "USD": "$", "JPY": "¥"}

	if symbol, ok := symbols[currency]; ok {
		return symbol + formatted
	}

	return fmt.Sprintf("%s %s", formatted, currency)
}
