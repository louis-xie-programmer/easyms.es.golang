package model

// 分类下主要的参数
type Category_Attribute struct {
	CategoryID    int
	AttributeName string
	Sort          int
}

// 属性名称分类及排序
type AttributeModel struct {
	AttributeName string `json:"AttributeName"`
	Type          string `json:"Type,omitempty"` //将属性分类信息存储到sqlite里，不再存入es，每次查询直接从sqlite里查一次，查询的作业中注意定时更新
	Sort          int    `json:"Sort,omitempty"` //将属性分类信息存储到sqlite里，不再存入es，每次查询直接从sqlite里查一次，查询的作业中注意定时更新
}

// 分类信息
type Category struct {
	CategoryID   int
	CategoryName string
}

// PrentCategory 父级分类信息
type PrentCategory struct {
	CategoryID   int
	CategoryName string
	Categorys    []Category
}
