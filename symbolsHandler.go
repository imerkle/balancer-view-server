package main

import (
	"balancer-view/config"
	"balancer-view/syncer"
	"net/http"
	"strings"
)

// Only search by name or ticker for now
func seachSymbols(tickerOrName, typeStr, exchange string) (result []config.Symbol, count int) {
	tickerOrName = strings.ToLower(tickerOrName)
	// Search by ticker name or symbol
	for _, symbol := range syncer.Symbols {
		if strings.Contains(strings.ToLower(symbol.Name), tickerOrName) ||
			strings.Contains(strings.ToLower(symbol.Ticker), tickerOrName) {
			result = append(result, symbol)
		}
	}
	count = len(result)
	return
}

func findSymbol(symbolStr string) (result config.Symbol, found bool) {
	ar := strings.Split(symbolStr, ":")
	if len(ar) > 1 && strings.TrimSpace(ar[1]) != "" {
		// Ignore exchange
		symbolStr = ar[1]
	}
	for _, symbol := range syncer.Symbols {
		if strings.ToLower(symbol.Name) == strings.ToLower(symbolStr) {
			result = symbol
			found = true
			return
		}
	}
	return
}

// symbolsHandler returns a specific symbol by Ticker or Name or
// Exchange and ticker as in the following format of "Exchange:Ticker"
func symbolsHandler(w http.ResponseWriter, r *http.Request) {
	symbolStr := r.URL.Query().Get("symbol")
	if symbolStr == "" {
		respondError(w, "", err400)
		return
	}

	result, found := findSymbol(symbolStr)
	if !found {
		respondError(w, "", err404)
		return
	}
	respondJSON(w, result, ok200)
}

// SearchResultSymbol describes a symbol for search result
type SearchResultSymbol struct {
	Symbol      string `json:"symbol"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Exchange    string `json:"exchange"`
	Ticker      string `json:"ticker"`
	Type        string `json:"type"`
}

// symbolsHandler returns a specific symbol by Ticker or Name or
// Exchange and ticker as in the following format of "Exchange:Ticker"
// GET Params:
// @query
// @type
// @exchange
// @limit
func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")

	result, count := seachSymbols(query, "", "")
	if count == 0 {
		respondError(w, "", err404)
		return
	}
	srsResult := []SearchResultSymbol{}
	for i := 0; i < len(result); i++ {
		s := result[i]
		srsResult = append(srsResult, SearchResultSymbol{
			Symbol:      s.Name,
			FullName:    s.Description,
			Description: s.Description,
			Exchange:    s.Exchange,
			Ticker:      s.Ticker,
			Type:        s.Type,
		})
	}
	respondJSON(w, srsResult, ok200)
}
