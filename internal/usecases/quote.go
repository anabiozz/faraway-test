package usecases

import (
	"math/rand"
)

// QuoteUsecase defines the interface for quote retrieval.
type QuoteUsecase interface {
	GetRandomQuote() string
}

type quoteUsecaseImpl struct{}

func NewQuoteUsecase() QuoteUsecase {
	return &quoteUsecaseImpl{}
}

// GetRandomQuote returns a random quote from a predefined list.
func (q *quoteUsecaseImpl) GetRandomQuote() string {
	quotes := []string{
		"Life is what happens when you're busy making other plans.",
		"The greatest glory in living lies not in never falling, but in rising every time we fall.",
		"The way to get started is to quit talking and begin doing.",
	}
	return quotes[rand.Intn(len(quotes))]
}
