package utility

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/vanh01/lingo"
)

// ReplaceStandStr 标准化字符串，特殊字符更换成空格，并将多空格改成单空格
func ReplaceStandStr(str string, repstr ...string) string {
	rep := " "
	if len(repstr) > 0 {
		rep = repstr[0]
	}
	str = strings.ReplaceAll(str, "é", "e")
	str = strings.ReplaceAll(str, "ü", "u")
	reg, err := regexp.Compile(fmt.Sprintf("[^a-zA-Z0-9%s]+", rep))
	if err == nil {
		str = reg.ReplaceAllString(str, rep)
	}
	return strings.ToLower(ReplaceMultipleSpacesWithSingle(str, rep))
}

// StandProductKeyWord 标准化关键字
func StandProductKeyWord(str string) (string, error) {
	str = ReplaceStandStr(str)

	strs := strings.Split(str, " ")

	var rst []string
	var currStr string
	for _, str := range strs {
		if len(str) > 2 {
			rst = append(rst, str)
			if len(currStr) > 0 {
				currStr += str
			}
		} else {
			if len(rst) > 0 {
				currStr = rst[len(rst)-1] + str
			} else {
				currStr += str
			}
		}
		if len(currStr) > 2 {
			rst = append(rst, currStr)
			currStr = ""
		}
	}

	if len(rst) < 1 {
		return "", fmt.Errorf("KeyWord length cannot be less than 3")
	}

	return strings.Join(rst, " "), nil
}

// ReplaceStandProductNameStr 标准化型号，大写，去掉空格
func ReplaceStandProductNameStr(str string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err == nil {
		str = reg.ReplaceAllString(str, "")
	}
	return strings.ToUpper(ReplaceMultipleSpacesWithSingle(str))
}

// ReplaceMultipleSpacesWithSingle 多空格更换为单空格,去除左右空格
func ReplaceMultipleSpacesWithSingle(str string, repstr ...string) string {
	rep := " "
	if len(repstr) > 0 {
		rep = repstr[0]
	}
	str = strings.TrimSpace(str)
	if rep == "" {
		return str
	}
	str = strings.Trim(str, rep)
	reg := regexp.MustCompile(fmt.Sprintf("[%s]+", rep))
	return reg.ReplaceAllString(str, rep)
}

// RemoveRepeatStr 去除重复
func RemoveRepeatStr(str string) string {
	strs := strings.Split(str, " ")
	strs = lingo.AsEnumerable(strs).Where(func(s string) bool {
		return s != ""
	}).Distinct().ToSlice()
	return strings.Join(strs, " ")
}

// Unique 删除重复
func Unique(intSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// RemoveFlagStr 删除标志字符
func RemoveFlagStr(str string) string {
	str = strings.ReplaceAll(str, "}", "")
	str = strings.ReplaceAll(str, "{", "")
	return str
}

// splitSegments splits the input string into segments separated by outer semicolons
func splitSegments(input string) []string {
	re := regexp.MustCompilePOSIX(`\{[^}]*\}|[^;]+`)
	matches := re.FindAllString(input, -1)
	var segments []string
	for _, match := range matches {
		if match != ";" {
			segments = append(segments, match)
		}
	}
	return segments
}

// splitProducts splits a segment into products separated by outer commas
func splitProducts(segment string) []string {
	re := regexp.MustCompile(`\{[^}]*\}|[^,]+`)
	matches := re.FindAllString(segment, -1)
	var products []string
	for _, match := range matches {
		if match != "," && match != "" {
			part := strings.Trim(match, "{}")
			products = append(products, part)
		}
	}
	return products
}

func SplitAttrValue(input string) (string, string, error) {
	// 定义正则表达式模式，匹配以 { 开始并以 } 结束的部分
	re := regexp.MustCompile(`^\{([^}]*)\}:\{(.*)\}$`)

	// 查找匹配项
	matches := re.FindStringSubmatch(input)
	if matches == nil || len(matches) != 3 {
		return "", "", fmt.Errorf("input string does not match the expected pattern")
	}

	// 返回匹配的两个部分
	return matches[1], matches[2], nil
}

// 转化拼接
func SplitStrToMap(input string) map[int][]string {
	if strings.HasPrefix(input, ";{") {
		input = "{}" + input
	}
	if strings.HasSuffix(input, "};") {
		input += "{}"
	}
	segments := splitSegments(input)

	result := make(map[int][]string)

	for i, segment := range segments {
		segment = strings.TrimSpace(segment)
		if len(segment) < 1 {
			result[i] = make([]string, 0)
			continue
		}

		// 分割 segment 以获取每个部分
		products := splitProducts(segment)

		// 将 products 中的元素添加到对应的 result 列表
		result[i] = products
	}

	return result
}

// GetProductNameIndex 特殊字符开头则返回37
func GetProductNameIndex(partno string) int {
	const strs = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	initial := string(partno[0])
	index := strings.Index(strs, initial)
	if index == -1 {
		index = 37
	} else {
		index++
	}
	return index
}

// GetProductNameStartStr 反向获取型号首字母
func GetProductNameStartStr(partNoIndex int) string {
	const strs = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if partNoIndex > len(strs) {
		return "00"
	}
	return string(strs[partNoIndex-1])
}

func EsRegexStandStr(str string) string {
	strs := `\,.,?,+,*,|,{,},[,],(,),",#,@,&,<,>,~`
	for _, str1 := range strings.Split(strs, ",") {
		str = strings.ReplaceAll(str, str1, "\\\\"+str1)
	}
	return str
}

// extractHighlightedStr 从查询字符串中提取<em>标签中的内容
func ExtractHighlightedStr(query string) string {
	re := regexp.MustCompile(`<em>([a-zA-Z0-9]+)</em>`)
	match := re.FindStringSubmatch(query)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

// highlightProductialMatch 在 target 中找到与 str1 部分匹配的最长子串并高亮显示
func HighlightProductialMatch(target, str1 string) string {
	if str1 == "" {
		return target
	}

	// 为了加快效率, 可对直接
	if strings.Contains(target, str1) {
		target = strings.ReplaceAll(target, str1, fmt.Sprintf("<em>%s</em>", str1))
		return target
	}

	// 用于存储最长匹配的结果
	var longestMatch string
	var startIndex int

	var str = "0123456789ABCDEFGHIJKLMNOPQRSTVUWXYZ"
	// 逐个字符检查
	for i, ch := range target {
		curr := fmt.Sprintf("%c", ch)
		if !strings.Contains(str, curr) {
			continue
		}
		longestMatch += curr
		if strings.HasPrefix(str1, longestMatch) {
			if len(longestMatch) == len(str1) {
				startIndex = i
				break
			}
		} else {
			longestMatch = ""
			break
		}
	}

	// 如果找到了匹配项，则进行高亮处理
	if len(longestMatch) > 0 {
		highlighted := fmt.Sprintf("<em>%s</em>", target[:startIndex+1])
		if len(target) > startIndex+1 {
			return highlighted + target[startIndex+1:]
		} else {
			return highlighted
		}
	}

	// 如果没有找到任何匹配项，返回原始字符串
	return target
}

func ReSegmentationProductName(partno string) string {
	partnos := strings.Split(partno, " ")

	var rst string
	for _, part := range partnos {
		rst = rst + SegmentationProductName(part)
	}

	return rst
}

func SegmentationProductName(partno string) string {
	if len(partno) <= 3 {
		return partno
	}
	var rst = partno[:3]
	for i := 2; i < len(partno); i++ {
		rst += fmt.Sprintf(" %s", partno[:i+1])
	}
	return rst
}

// ExtractFirstEmContent 从字符串中提取第一个 <em> 标签中的内容
func ExtractFirstEmContent(str string) (string, bool) {
	re := regexp.MustCompile(`<em>(.*?)</em>`)
	matches := re.FindStringSubmatch(str)
	if len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}

// ContainsString 检查字符串数组中是否包含指定的字符串
func ContainsString(arr []string, target string) bool {
	for _, str := range arr {
		if str == target {
			return true
		}
	}
	return false
}
