package model

// ProductSearchParam 产品搜索参数
type ProductSearchParam struct {
	ProductName  string `json:"ProductName,omitempty"`
	Brand        string `json:"Brand,omitempty"`
	Category     string `json:"Category,omitempty"`
	Package      string `json:"Package,omitempty"`
	PassiveParam string `json:"PassiveParam,omitempty"`
	CurrFitler   string `json:"CurrFitler,omitempty"`

	ParentIDs      []int32     `json:"ParentID,omitempty"`
	CategoryIDs    []int32     `json:"CategoryID,omitempty"`
	BrandIDs       []int32     `json:"BrandIDs,omitempty"`
	DistributorIDs []int32     `json:"DistributorIDs,omitempty"`
	AttributeNames []string    `json:"AttributeValueIDs,omitempty"`
	Attributes     []Attribute `json:"Attributes,omitempty"`

	IsHighlight bool `json:"IsHighlight,omitempty"`

	Size int32 `json:"Size,omitempty"`
	From int32 `json:"From,omitempty"`

	Sources    string `json:"Sources,omitempty"`
	IsReSearch bool   `json:"IsReSearch,omitempty"`
}

// ProductIndexSearchParam 产品索引搜索参数
type ProductIndexSearchParam struct {
	ProductIndexParam

	Size    int32 `json:"Size,omitempty"`
	LastPID int32 `json:"GtFrom,omitempty"`

	Sources string `json:"Sources,omitempty"`
}

// ProductAggSearchParam 产品聚合搜索参数
type ProductIndexParam struct {
	ProductNameIndex int32 `json:"ProductNameIndex,omitempty"`
	IsFollow         bool  `json:"IsFollow,omitempty"`
	ParentID         int32 `json:"ParentID,omitempty"`
	CategoryID       int32 `json:"CategoryID,omitempty"`
}

// ProductAggSearchParam 产品聚合搜索参数
type CategoryAggSearchParam struct {
	BrandID int32 `json:"BrandID,omitempty"`
}

type Attribute struct {
	AttributeName   string   `json:"AttributeName"`
	AttributeValues []string `json:"AttributeValues"`
}

type CustomSearchProductParam struct {
	KeyWords       map[string][]string
	Filters        map[string]string
	TermsFilters   map[string]string
	Range          CustomRange
	AttributeNames []string
	Sources        string
	Sorts          map[string]string
	Boots          []int
	IsHighlight    bool `json:"IsHighlight,omitempty"`
	Size           int32
	From           int32
}

type CustomRange struct {
	Field string
	Gte   int32
	Lte   int32
}
