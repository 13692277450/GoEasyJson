package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// JSONFormatInfo 存储JSON字段的格式信息
type JSONFormatInfo struct {
	IsInt         bool
	IsFloat       bool
	IntLength     int
	FloatDecimals int
}

// AnalyzeJSONStructure 分析JSON结构并记录每个字段的格式信息
func AnalyzeJSONStructure(jsonData []byte) (map[string]JSONFormatInfo, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	formatInfo := make(map[string]JSONFormatInfo)
	for key, value := range data {
		formatInfo[key] = analyzeValueFormat(value)
	}

	return formatInfo, nil
}

// analyzeValueFormat 分析单个值的格式信息
func analyzeValueFormat(value interface{}) JSONFormatInfo {
	var info JSONFormatInfo

	switch v := value.(type) {
	case float64:
		// 检查是否为整数
		if math.Floor(v) == v {
			info.IsInt = true
			info.IntLength = len(strconv.FormatInt(int64(v), 10))
		} else {
			info.IsFloat = true
			// 计算小数位数
			str := strconv.FormatFloat(v, 'f', -1, 64)
			parts := strings.Split(str, ".")
			if len(parts) > 1 {
				info.FloatDecimals = len(parts[1])
			}
		}
	case int, int8, int16, int32, int64:
		info.IsInt = true
		info.IntLength = len(fmt.Sprintf("%d", v))
	}

	return info
}
