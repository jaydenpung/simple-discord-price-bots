package utils

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func FormatAmount(amount float64) string {
	amountPostfix := ""
	if amount > 1_000_000_000 {
		amountPostfix = "B"
		amount = amount / 1_000_000_000
	} else if amount > 1_000_000 {
		amountPostfix = "M"
		amount = amount / 1_000_000
	} else if amount > 100_000 {
		amountPostfix = "K"
		amount = amount / 1_000
	}

	p := message.NewPrinter(language.English)
	formattedAmount := p.Sprintf("%.2f%s\n", amount, amountPostfix)

	return formattedAmount
}
