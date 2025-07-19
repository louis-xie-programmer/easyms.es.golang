package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"sync"
)

// easyAppConfig 系统配置
var easyAppConfig *ViperConfig

// easyTaskConfigList 配置列表
var easyTaskConfigList map[string]*ViperConfig

type ViperConfig struct {
	Name       string
	ConfigFile *viper.Viper
	Mutex      sync.RWMutex
}

// InitConfig 初始化函数，用于设置默认配置
func InitConfig(appName string) {
	//创建系统配置
	CreateAppConfig(appName, "config.yaml")

	easyTaskConfigList = make(map[string]*ViperConfig)
}

// GetAppConfig 获取项目默认配置
func GetAppConfig() *ViperConfig {
	return easyAppConfig
}

// GetTaskConfig 获取任务配置
func GetTaskConfig(name string) *ViperConfig {
	v, exists := easyTaskConfigList[name]
	if !exists {
		return nil
	}
	return v
}

// CreateAppConfig 创建项目默认配置
func CreateAppConfig(appName string, filename string) {
	easyAppConfig = &ViperConfig{
		Name:       appName,
		ConfigFile: viper.New(),
	}
	easyAppConfig.Mutex.Lock()
	defer easyAppConfig.Mutex.Unlock()

	easyAppConfig.ConfigFile.SetConfigFile(fmt.Sprintf("/conf/%s/%s", appName, filename))
	if err := easyAppConfig.ConfigFile.ReadInConfig(); err != nil {
		log.Fatalf("Error loading config file %s: %s", filename, err)
	}
}

// CreateTaskConfig 创建任务配置
func CreateTaskConfig(name string, filename string) {
	v := GetTaskConfig(name)
	// 如果 appName 不存在，则创建并追加到列表
	if v == nil {
		v = &ViperConfig{
			Name:       name,
			ConfigFile: viper.New(),
		}
		easyTaskConfigList[name] = v
	}

	v.Mutex.Lock()
	defer v.Mutex.Unlock()

	v.ConfigFile.SetConfigFile(filename)
}

// GetTaskConfigValue 获取配置
func GetTaskConfigValue[T any](name string) (*T, bool) {
	v := GetTaskConfig(name)
	var value T
	err := v.ConfigFile.Unmarshal(value)
	if err != nil {
		return nil, false
	}
	return &value, true
}

// GetAppConfigValue 获取系统配置
func GetAppConfigValue[T any](key string) (*T, bool) {
	v := GetAppConfig()
	return getConfigValue[T](v, key)
}

// 获取配置值
func getConfigValue[T any](v *ViperConfig, key string) (*T, bool) {
	if v == nil {
		return nil, false
	}
	v.Mutex.Lock()
	defer v.Mutex.Unlock()

	if v.ConfigFile.IsSet(key) == false {
		return nil, false
	}

	value := v.ConfigFile.Get(key).(T)

	return &value, true
}

// UpdateAppConfig 修改配置
func UpdateAppConfig(key any, value any) {
	easyAppConfig.Mutex.Lock()
	defer easyAppConfig.Mutex.Unlock()

	easyAppConfig.ConfigFile.Set(key.(string), value)
}

// UpdateTaskConfig 修改任务配置
func UpdateTaskConfig(name string, key any, value any) {
	v := GetTaskConfig(name)

	v.Mutex.Lock()
	defer v.Mutex.Unlock()

	v.ConfigFile.Set(key.(string), value)
}

// SaveAppConfig 保存配置
func SaveAppConfig() error {
	easyAppConfig.Mutex.Lock()
	defer easyAppConfig.Mutex.Unlock()

	if err := easyAppConfig.ConfigFile.SafeWriteConfig(); err != nil {
		return err
	}
	return nil
}

// SaveTaskConfig 保存任务配置 写入文件
func SaveTaskConfig(name string) error {
	v := GetTaskConfig(name)

	v.Mutex.Lock()
	defer v.Mutex.Unlock()

	if err := v.ConfigFile.SafeWriteConfig(); err != nil {
		return err
	}
	return nil
}

func RemoveTaskConfig(name string) {
	delete(easyTaskConfigList, name)
}
