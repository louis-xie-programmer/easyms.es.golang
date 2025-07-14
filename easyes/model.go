package easyes

import (
	"encoding/json"
)

type DocResponse struct {
	Source json.RawMessage `json:"_source"`
}

type SearchResponse struct {
	Took int
	Hits struct {
		Total struct {
			Value    int
			Relation string
		}
		Hits []struct {
			Source    json.RawMessage     `json:"_source"`
			Highlight map[string][]string `json:"highlight"`
		}
	}
}

type CollapseSearchResponse struct {
	Took int
	Hits struct {
		Total struct {
			Value    int
			Relation string
		}
		Hits []struct {
			Source    json.RawMessage  `json:"_source"`
			Fields    map[string][]int `json:"fields"`
			InnerHits map[string]struct {
				Hits struct {
					Total struct {
						Value    int
						Relation string
					}
					Hits []struct {
						Source json.RawMessage `json:"_source"`
					}
				}
			} `json:"inner_hits,omitempty"`
		}
	}
}

type Token struct {
	Token       string `json:"token,omitempty"`
	StartOffset int32  `json:"start_offset,omitempty"`
	EndOffset   int32  `json:"end_offset,omitempty"`
	Type        string `json:"type,omitempty"`
	Position    int32  `json:"position,omitempty"`
	OldToken    string `json:"old_token,omitempty"`
}

type Tokens struct {
	Tokens []Token `json:"tokens,omitempty"`
}

type Aggregation struct {
	Buckets []struct {
		Key      any `json:"key,omitempty"`
		DocCount int `json:"doc_count,omitempty"`
		BgCount  int `json:"bg_count,omitempty"`
		Top      struct {
			Hits struct {
				Hits []struct {
					Source json.RawMessage `json:"_source,omitempty"`
				} `json:"hits,omitempty"`
			} `json:"hits,omitempty"`
		} `json:"top,omitempty"`
	} `json:"buckets,omitempty"`
}

type AggResult struct {
	Aggregations map[string]Aggregation `json:"aggregations,omitempty"`
}

type GradeAggregation struct {
	Buckets []struct {
		Key          any         `json:"key,omitempty"`
		DocCount     int         `json:"doc_count,omitempty"`
		Aggregations Aggregation `json:"child_agg,omitempty"`
	} `json:"buckets,omitempty"`
}

type GradeAggResult struct {
	Aggregations map[string]GradeAggregation `json:"aggregations,omitempty"`
}
