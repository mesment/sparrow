
// 修改内容:
// 添加配置日志输出方式,支持同时输出到控制台和文件


//``````````````````````````````````````````````````````````````````
// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xlog

import (
	"fmt"
	"time"

	"github.com/mesment/sparrow/pkg/conf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config ...
type Config struct {
	// 日志是否输出到控制台
	EnableConsole     bool
	// 控制台日志显示格式是否以json格式显示
	ConsoleJSONFormat bool
	// 控制台日志输出级别
	ConsoleLevel      string
	// 日志是否输出到文件
	EnableFile        bool
	// 文件日志显示格式是否以json格式显示
	FileJSONFormat    bool
	// Dir 日志文件输出目录
	Dir string
	// Name 日志文件名称
	Name string
	// Level 日志文件初始等级
	Level string
	// 日志初始化字段
	Fields []zap.Field
	// 是否添加调用者信息
	AddCaller bool
	// 日志前缀
	Prefix string
	// 日志输出文件最大长度，超过改值则截断（单位 M）
	MaxSize   int
	// 保存日志文件最长天数
	MaxAge    int
	// 保存日志文件最大个数
	MaxBackup int
	// 日志磁盘刷盘间隔
	Interval      time.Duration
	CallerSkip    int
	Async         bool
	Queue         bool
	QueueSleep    time.Duration
	Core          zapcore.Core
	// 开启日志级别颜色显示
	Debug         bool
	EncoderConfig *zapcore.EncoderConfig
	configKey     string
}

// Filename ...
func (config *Config) Filename() string {
	return fmt.Sprintf("%s/%s", config.Dir, config.Name)
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil {
		panic(err)
	}
	config.configKey = key
	return config
}

// StdConfig Jupiter Standard logger config
func StdConfig(name string) *Config {
	return RawConfig("jupiter.logger." + name)
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Name:          "default.log",
		Dir:           ".",
		Level:         "info",
		MaxSize:       500, // 500M
		MaxAge:        1,   // 1 day
		MaxBackup:     10,  // 10 backup
		Interval:      24 * time.Hour,
		CallerSkip:    1,
		AddCaller:     true,
		Async:         true,
		Queue:         false,
		QueueSleep:    100 * time.Millisecond,
		EncoderConfig: DefaultZapConfig(),
	}
}

// Build ...
func (config Config) Build() *Logger {
	if config.EncoderConfig == nil {
		config.EncoderConfig = DefaultZapConfig()
	}
	if config.Debug {
		// 设置日志级别显示颜色
		config.EncoderConfig.EncodeLevel = DebugEncodeLevel
	}
	logger := newLogger(&config)
	if config.configKey != "" {
		logger.AutoLevel(config.configKey + ".level")
	}
	return logger
}


