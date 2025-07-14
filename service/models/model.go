package models

import (
	"easyms-es/model"
)

type SearchResult struct {
	Total int32 `json:"total,omitempty"`
	Above bool  `json:"above,omitempty"`
	From  int32 `json:"from,omitempty"`
	Size  int32 `json:"size,omitempty"`
}

type SearchProductResult struct {
	SearchResult
	Results             []model.Product `json:"results,omitempty"`
	IsReSearch          bool
	SimilarProductNames []string
}

type SearchPriceResult struct {
	SearchResult
	Results []model.StockPrice `json:"results,omitempty"`
}

type CollapsePrice struct {
	PID     int32              `json:"pid"`
	Total   int32              `json:"total,omitempty"`
	Results []model.StockPrice `json:"results,omitempty"`
}

type CollapseSearchPriceResult struct {
	SearchResult
	Results []CollapsePrice `json:"results,omitempty"`
}
