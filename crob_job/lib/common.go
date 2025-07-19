package lib

import (
	"easyms-es/config"
	"easyms-es/db"
	"easyms-es/model"
	"easyms-es/utility"
	"encoding/json"
	"fmt"
	mssql "github.com/microsoft/go-mssqldb"
	"github.com/vanh01/lingo"
	"log"
	"strconv"
	"strings"
)

func GetMaxPid() int {
	maxPid, _ := config.GetAppConfigValue[int]("watchjob.maxpid")
	return *maxPid
}

// QueryProduct 查询产品信息，输出插入或更新的商品条目，要删除的商品条目
func QueryProduct(param string, limit int) ([]model.Product, []model.Product, int, error) {
	var addProducts []model.Product
	var delProducts []model.Product

	// 产品表数据， 注意：对数据做分类处理：新增+更新，删除
	//sql := fmt.Sprintf(`SELECT TOP (%d) PID,ProductName,Brand,BrandID,CategoryID,ParentID,IsDeleted FROM Products with(nolock) WHERE %s ORDER BY PID`, limit, param)
	type Product struct {
		PID         int
		ProductName string
		Brand       string
		BrandID     int
		CategoryID  int
		ParentID    int
		IsDeleted   bool
	}

	var ps []Product
	//db.BasicDB.Raw(sql).Scan(&addProducts)
	db.TenantPoolInstance.GetTable("Products").Where(param).Order("PID").Limit(limit).Find(&ps)
	if len(addProducts) < 1 {
		return nil, nil, 0, nil
	}

	// 一般性赋值
	for _, p := range ps {
		// 删除的型号归类
		if p.IsDeleted {
			delProducts = append(delProducts, model.Product{
				PID: p.PID,
			})
			continue
		}
		product := model.Product{
			PID:              p.PID,
			StandProductName: utility.ReplaceStandProductNameStr(p.ProductName),
			BrandID:          p.BrandID,
			CategoryID:       p.CategoryID,
			ParentID:         p.ParentID,
			StandBrand:       utility.ReplaceMultipleSpacesWithSingle(p.Brand),
		}

		productName := strings.ToUpper(utility.ReplaceMultipleSpacesWithSingle(p.ProductName))

		product.ProductNameIndex = utility.GetProductNameIndex(productName)

		brand := utility.ReplaceMultipleSpacesWithSingle(p.Brand)

		// 将商品名，品牌整合入ProductExt，注意后续的拼接{ProductName},{Brand},{Description},{Canonical}
		product.ProductExt = fmt.Sprintf("{%s},{%s}", utility.RemoveFlagStr(productName), utility.RemoveFlagStr(brand))

		addProducts = append(addProducts, product)
	}

	// 分类,产品图片，价格，等扩展内容赋值()

	return addProducts, delProducts, addProducts[len(addProducts)-1].PID, nil
}

// QueryStockPrice 价格查询,注意:保持型号索引库的一致性, 同时注意这里不用对逻辑删除的数据进行处理,这里指的删除数据时下架数据,用另一个job进行反向校验
// 返回新增或更新价格，待删除价格, 多分销商价格涉及的排序规则
func QueryStockPrice(sql string, distributorType int) ([]model.StockPrice, []model.StockPrice, int, error) {
	var (
		stockPrices    []model.StockPrice
		delStockPrices []model.StockPrice
	)
	// 价格表
	type PriceStock struct {
		SID                   int
		PID                   int
		ProductName           string
		DistributorID         int
		Distributor           string
		DistributorProductUrl string
		StockNum              int
		Currency              string
		StepPrice             string
		UpdateTime            mssql.DateTime1
		IsDeleted             bool
	}
	var prices []PriceStock
	db.TenantPoolInstance.GetTable("Prices").Raw(sql).Scan(&prices)

	type StepPrice struct {
		Qty   int
		Price float64
	}
	// 一般性数据赋值,阶梯价格处理, 扩展查询参数整理
	for _, price := range prices {
		stockPrice := model.StockPrice{
			DistributorType: distributorType,
			DistributorID:   price.DistributorID,
			PID:             price.PID,
		}
		if distributorType == model.EsPriceCrawlStart {
			stockPrice.SPID = fmt.Sprintf("%d-%d-%d", distributorType, price.DistributorID, price.PID)
		} else {
			stockPrice.SPID = fmt.Sprintf("%d", price.SID)
		}

		if price.DistributorID < 1 || price.PID < 1 {
			delStockPrices = append(delStockPrices, stockPrice)
			continue
		}

		// 阶梯价格处理
		if len(price.StepPrice) > 0 {
			var priceStep []StepPrice
			err := json.Unmarshal([]byte(price.StepPrice), &priceStep)
			if err != nil {
				log.Println("StepPrice json Unmarshal error: ", price.StepPrice)
				continue
			}
			// 对价格基于qty做倒序便于便利赋值
			priceStep = lingo.AsEnumerable(priceStep).OrderByDescending(func(sp StepPrice) any {
				return sp.Qty
			}).ToSlice()
			// 为了获取最准确的价格,这里一定要注意已经赋值后不再重复赋值
			for _, step := range priceStep {
				switch qty := step.Qty; {
				case qty == 1:
					if step.Price > 0 && len(stockPrice.StepPrice1) == 0 {
						stockPrice.StepPrice1 = strconv.FormatFloat(step.Price, 'f', 4, 32)
					}
				case qty <= 10:
					if step.Price > 0 && len(stockPrice.StepPrice2) == 0 {
						stockPrice.StepPrice2 = strconv.FormatFloat(step.Price, 'f', 4, 32)
					}
				case qty <= 100:
					if step.Price > 0 && len(stockPrice.StepPrice3) == 0 {
						stockPrice.StepPrice3 = strconv.FormatFloat(step.Price, 'f', 4, 32)
					}
				case qty <= 1000:
					if step.Price > 0 && len(stockPrice.StepPrice4) == 0 {
						stockPrice.StepPrice4 = strconv.FormatFloat(step.Price, 'f', 4, 32)
					}
				case qty <= 10000:
					if step.Price > 0 && len(stockPrice.StepPrice5) == 0 {
						stockPrice.StepPrice5 = strconv.FormatFloat(step.Price, 'f', 4, 32)
					}
				}
			}
			// 对价格的回写,这里可以根据运营需求做出调整
			if len(stockPrice.StepPrice1) > 0 && len(stockPrice.StepPrice2) == 0 {
				stockPrice.StepPrice2 = stockPrice.StepPrice1
			}
			if len(stockPrice.StepPrice2) > 0 && len(stockPrice.StepPrice3) == 0 {
				stockPrice.StepPrice3 = stockPrice.StepPrice2
			}
			if len(stockPrice.StepPrice3) > 0 && len(stockPrice.StepPrice4) == 0 {
				stockPrice.StepPrice4 = stockPrice.StepPrice3
			}
			if len(stockPrice.StepPrice4) > 0 && len(stockPrice.StepPrice5) == 0 {
				stockPrice.StepPrice5 = stockPrice.StepPrice4
			}
		}

		// 价格排序（按照业务规则定义）
		sort := 0
		if stockPrice.StockNum > 0 {
			sort += 10
		}
		if len(stockPrice.StepPrice1) > 0 {
			sort += 10
		}
		if len(stockPrice.StepPrice2) > 0 {
			sort += 10
		}
		if len(stockPrice.StepPrice3) > 0 {
			sort += 10
		}
		if len(stockPrice.StepPrice4) > 0 {
			sort += 10
		}
		if len(stockPrice.StepPrice5) > 0 {
			sort += 10
		}

		stockPrice.Sort = sort

		stockPrices = append(stockPrices, stockPrice)
	}

	// 分销商基础数据的扩展 ，es 中查询相关型号的查询字段, 为保持于型号库的一致性
	// 排除不合规的价格

	lastSId := 0
	if len(prices) > 0 {
		lastSId = prices[len(prices)-1].SID
	}

	return stockPrices, delStockPrices, lastSId, nil
}

// MapperToProducts 商品对象转化
func MapperToProducts(products []interface{}) []model.Product {
	var ps []model.Product
	for _, p := range products {
		ps = append(ps, p.(model.Product))
	}
	return ps
}

// MapperToPrices 价格对象转化
func MapperToPrices(prices []interface{}) []model.StockPrice {
	var ps []model.StockPrice
	for _, price := range prices {
		ps = append(ps, price.(model.StockPrice))
	}
	return ps
}

// MapperFromProducts 商品对象泛化
func MapperFromProducts(products []model.Product) []interface{} {
	var ps []interface{}
	for _, p := range products {
		products = append(products, p)
	}
	return ps
}

// MapperFromPrices 价格对象泛化
func MapperFromPrices(prices []model.StockPrice) []interface{} {
	var ps []interface{}
	for _, price := range prices {
		ps = append(ps, price)
	}
	return ps
}
