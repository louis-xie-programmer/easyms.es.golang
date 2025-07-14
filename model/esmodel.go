package model

import "time"

var EsProductIndexName string
var EsPriceIndexName string

var EsPriceJoinStart = 1
var EsPriceCrawlStart = 2

// Product 商品库
type Product struct {
	//主键 IdProperty
	PID int `json:"PID" es:"type:integer" id_property:"true"`

	// 产品名称分词
	StandProductName string `json:"StandProductName,omitempty" es:"type:text,analyzer:easy_product_max,search_analyzer:easy_product"`
	// 品牌分词
	StandBrand string `json:"StandBrand,omitempty" es:"type:text,analyzer:easy_brand,search_analyzer:easy_brand"`
	// 一级分类 + 空格 + 二级分类， 去除特殊符； 空格分词
	StandCategory string `json:"StandCategory,omitempty" es:"type:text,analyzer:easy_category,search_analyzer:easy_category"` //easy_category
	PassiveParam  string `json:"PassiveParam,omitempty" es:"type:text,analyzer:easy_param,search_analyzer:easy_param"`        //easy_param

	//索引字段(filter) 型号首字母1-36,99
	ProductNameIndex int  `json:"ProductNameIndex,omitempty" es:"type:integer"`
	Follow           bool `json:"Follow,omitempty" es:"type:boolean"`     // SEO 推广用
	PriceGroup       int  `json:"PriceGroup,omitempty" es:"type:integer"` // 调整为只对入驻商家和非入驻商家的区别（0,1,2）0:无价格；1：有价格；2：有入驻分销商价格

	//基本信息
	ProductExt   string `json:"ProductExt,omitempty" es:"type:keyword,index:false"`   //{ProductName},{Brand},{Description},{Canonical}
	ResUrls      string `json:"ResUrls,omitempty" es:"type:keyword,index:false"`      //{PicUrl1},{PicUrl2};{PdfUrl}
	Manufacturer string `json:"Manufacturer,omitempty" es:"type:keyword,index:false"` // {品牌表的品牌名称},{品牌ID对应的标准名},{品牌图标}  品牌扩展信息
	Category     string `json:"Category,omitempty" es:"type:keyword,index:false"`     // {父级分类},{父级分类SEOName},{父级分类SEOTitle};{分类},{分类SEOName},{分类SEOTitle}

	//聚合索引字段(aggs,filter)
	BrandID    int `json:"BrandID,omitempty" es:"type:integer"`
	ParentID   int `json:"ParentID,omitempty" es:"type:integer"`
	CategoryID int `json:"CategoryID,omitempty" es:"type:integer"`

	DistributorIDs   []int    `json:"DistributorIDs,omitempty" es:"type:integer"`
	DistributorNames []string `json:"DistributorNames,omitempty" es:"type:keyword"` //原始名称
	AttributeNames   []string `json:"AttributeNames,omitempty" es:"type:keyword"`   //{属性分类}:{属性名}
	AttributeValues  []string `json:"AttributeValues,omitempty" es:"type:keyword"`  //{属性名}:{属性值}

	//为聚合搜索进行优化独立字段
	OriginalCategoryIDs []int  `json:"OriginalCategoryID,omitempty" es:"type:keyword"`
	OriginalCategoryExt string `json:"OriginalCategoryExt,omitempty" es:"type:keyword"` // {1name}:{2name}:{3name}
}

// 商家库存价格
type StockPrice struct {
	SPID string `json:"SPID" es:"type:keyword" id_property:"true"`
	//主键 IdProperty  DistributorType + '-' + DistributorID + '-' + PID

	PID                int  `json:"PID" es:"type:integer"`
	BrandID            int  `json:"BrandID,omitempty" es:"type:integer"`
	DistributorID      int  `json:"DistributorID" es:"type:integer"`
	DistributorType    int  `json:"DistributorType" es:"type:keyword"`              //1：入驻的分销商；2：非入驻分销商
	IsAuthorizeddealer bool `json:"ISAuthorizeddealer,omitempty" es:"type:keyword"` //分销商是否认证

	ProductExt     string `json:"ProductExt,omitempty" es:"type:keyword,index:false"`     //{ProductName},{Brand}
	DistributorExt string `json:"DistributorExt,omitempty" es:"type:keyword,index:false"` //{Distributor},{DistributorProductUrl} 销售商家扩展信息

	Currency string `json:"Currency,omitempty" es:"type:integer"`
	StockNum int    `json:"StockNum,omitempty" es:"type:integer"`

	PriceExt   string `json:"PriceExt,omitempty" es:"type:keyword,index:false"` //{SID},{ProductName},{Multiples},{Warehouse}
	StepPrice1 string `json:"StepPrice1,omitempty" es:"type:keyword"`
	StepPrice2 string `json:"StepPrice2,omitempty" es:"type:keyword"`
	StepPrice3 string `json:"StepPrice100,omitempty" es:"type:keyword"`
	StepPrice4 string `json:"StepPrice1k,omitempty" es:"type:keyword"`
	StepPrice5 string `json:"StepPrice1w,omitempty" es:"type:keyword"`

	UpdateTime time.Time `json:"UpdateTime,omitempty" es:"type:date"`
	Sort       int       `json:"Sort" es:"type:integer"`
}
