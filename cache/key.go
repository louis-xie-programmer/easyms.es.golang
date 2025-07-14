package cache

import (
	"fmt"
)

// 分类搜索聚合缓存命名
func getCategorySearchAggKey(prentCategory string, category string) string {
	if len(prentCategory) > 0 {
		return fmt.Sprintf("%s-%s:%s", "aggs", prentCategory, category)
	}
	return fmt.Sprintf("%s-%s-0", "aggs", category)
}

// GetCategorySearchCategoryAggKey 分类搜索下一级分类聚合缓存命名
func GetCategorySearchCategoryAggKey(prentCategory string, category string) string {
	return fmt.Sprintf("%s:categoryagg", getCategorySearchAggKey(prentCategory, category))
}

// GetCategorySearchBrandAggKey 分类搜索品牌聚合缓存命名
func GetCategorySearchBrandAggKey(prentCategory string, category string) string {
	return fmt.Sprintf("%s:brandagg", getCategorySearchAggKey(prentCategory, category))
}

// GetCategorySearchDistributorAggKey 分类搜索分销商聚合缓存命名
func GetCategorySearchDistributorAggKey(prentCategory string, category string) string {
	return fmt.Sprintf("%s:distributoragg", getCategorySearchAggKey(prentCategory, category))
}

// GetCategorySearchAttributeNameAggKey 分类搜索属性名聚合缓存命名
func GetCategorySearchAttributeNameAggKey(prentCategory string, category string) string {
	return fmt.Sprintf("%s:attrinameagg", getCategorySearchAggKey(prentCategory, category))
}

// GetCategorySearchAttributeValueAggKey 分类搜索属性值聚合缓存命名
func GetCategorySearchAttributeValueAggKey(prentCategory string, category string) string {
	return fmt.Sprintf("%s:attrivalusagg", getCategorySearchAggKey(prentCategory, category))
}

// GetIndexKey 产品首字母索引缓存命名, index-0-0 为首页索引
func GetIndexKey(indexKey string) string {
	return fmt.Sprintf("pindexs-%s", indexKey)
}

// GetIndexKeyWithCategoryId 首字母索引(分类)
func GetIndexKeyWithCategoryId(categoryId int) string {
	return fmt.Sprintf("indexs-%d", categoryId)
}

// GetBrandCategoryAggKey 品牌下索引聚合
func GetBrandCategoryAggKey(brandId int, prentCategoryId int) string {
	return fmt.Sprintf("brandaggs-%d:%d", brandId, prentCategoryId)
}
func GetBrandCategoryAggParentCategoryKey(brandId int) string {
	return fmt.Sprintf("brandaggs-%d:pcategoryagg", brandId)
}
