package config

import (
	"github.com/spf13/viper"
	"log"
	"sync"
)

var EasyViperConfigList []*ViperConfig

type ViperConfig struct {
	JobName    string
	ConfigFile *viper.Viper
	Mutex      sync.RWMutex
}

// InitConfig 初始化函数，用于设置默认配置
func InitConfig() {
	EasyViperConfigList = []*ViperConfig{}
}

// EasyViperConfigListJobNameFirst 根据JobName查找 ViperConfig
func EasyViperConfigListJobNameFirst(jobName string) *ViperConfig {
	if jobName == "" {
		jobName = "config"
	}
	for _, viperConfig := range EasyViperConfigList {
		if viperConfig.JobName == jobName {
			return viperConfig
		}
	}
	return nil
}

// CreateConfig 追加创建配置文件实例
func CreateConfig(jobName string, filename string) {
	v := EasyViperConfigListJobNameFirst(jobName)

	// 如果 jobName 不存在，则创建并追加到列表
	if v == nil {
		v = &ViperConfig{
			JobName:    jobName,
			ConfigFile: viper.New(),
		}
		EasyViperConfigList = append(EasyViperConfigList, v)
	}

	v.Mutex.Lock()
	defer v.Mutex.Unlock()

	v.ConfigFile.SetConfigFile(filename)
	if err := v.ConfigFile.ReadInConfig(); err != nil {
		log.Fatalf("Error loading config file %s: %s", filename, err)
	}
}

// GetSyncConfig 获取配置
func GetSyncConfig(jobName string, key string) string {
	v := EasyViperConfigListJobNameFirst(jobName)

	return v.ConfigFile.GetString(key)
}

// GetSyncConfig_Type 获取配置 Type
func GetSyncConfig_Type[T any](jobName string, key string) T {
	v := EasyViperConfigListJobNameFirst(jobName)

	var zero T // 定义类型T的零值
	var result any

	// 使用反射判断类型，根据类型调用不同的 viper 方法
	switch any(zero).(type) {
	case string:
		result = v.ConfigFile.GetString(key)
	case int:
		result = v.ConfigFile.GetInt(key)
	case float64:
		result = v.ConfigFile.GetFloat64(key)
	case int64:
		result = v.ConfigFile.GetInt64(key)
	case bool:
		result = v.ConfigFile.GetBool(key)
	case []int:
		result = v.ConfigFile.GetIntSlice(key)
	case []string:
		result = v.ConfigFile.GetStringSlice(key)
	case map[string]interface{}:
		result = v.ConfigFile.GetStringMap(key)
	case map[string]string:
		result = v.ConfigFile.GetStringMapString(key)
	default:
		return zero
	}

	// 类型断言并返回结果
	if val, ok := result.(T); ok {
		return val
	}
	return zero
}

// UpdateSyncConfig 修改配置
func UpdateSyncConfig(jobName string, key any, value any) {
	v := EasyViperConfigListJobNameFirst(jobName)

	v.ConfigFile.Set(key.(string), value)
}

// Save 保存配置 写入文件
func Save(jobName string) error {
	v := EasyViperConfigListJobNameFirst(jobName)

	v.Mutex.Lock()
	defer v.Mutex.Unlock()

	if err := v.ConfigFile.WriteConfig(); err != nil {
		return err
	}
	return nil
}

//func NewViperConfig(app string) *ViperConfig {
//	return &ViperConfig{
//		Path: "./conf/" + app,
//		Type: "yaml",
//		Name: "config",
//	}
//}
//
//func (v *ViperConfig) Load() error {
//	v.mutex.Lock()
//	defer v.mutex.Unlock()
//
//	viper.AddConfigPath(v.Path)
//	viper.SetConfigName(v.Name)
//	viper.SetConfigType(v.Type)
//	if err := viper.ReadInConfig(); err != nil {
//		return err
//	}
//	return nil
//}
