
// 修改内容:
// 支持通过配置设置日志输出方式,支持同时输出到控制台和文件


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
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mesment/sparrow/pkg/conf"
	"github.com/mesment/sparrow/pkg/defers"
	"github.com/mesment/sparrow/pkg/util/xcolor"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel = zap.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zap.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = zap.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-Level logs.
	ErrorLevel = zap.ErrorLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zap.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zap.FatalLevel
)

// Func ...
type (
	Func   func(string, ...zap.Field)
	Field  = zap.Field
	Level  = zapcore.Level
	Logger struct {
		desugar *zap.Logger
		lv      *zap.AtomicLevel
		config  Config
		sugar   *zap.SugaredLogger
	}
)

var (
	// String ...
	String = zap.String
	// Any ...
	Any = zap.Any
	// Int64 ...
	Int64 = zap.Int64
	// Int ...
	Int = zap.Int
	// Int32 ...
	Int32 = zap.Int32
	// Uint ...
	Uint = zap.Uint
	// Duration ...
	Duration = zap.Duration
	// Durationp ...
	Durationp = zap.Durationp
	// Object ...
	Object = zap.Object
	// Namespace ...
	Namespace = zap.Namespace
	// Reflect ...
	Reflect = zap.Reflect
	// Skip ...
	Skip = zap.Skip()
	// ByteString ...
	ByteString = zap.ByteString
)

func newLogger(config *Config) *Logger {
	zapOptions := make([]zap.Option, 0)
	zapOptions = append(zapOptions, zap.AddStacktrace(zap.DPanicLevel))
	if config.AddCaller {
		zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
	}
	if len(config.Fields) > 0 {
		zapOptions = append(zapOptions, zap.Fields(config.Fields...))
	}

	cores := []zapcore.Core{}
	lv := zap.NewAtomicLevelAt(getZapLevel(config.Level))
	if err := lv.UnmarshalText([]byte(config.Level)); err != nil {
		panic(err)
	}

	encoderConfig := *config.EncoderConfig //getEncoder(config.ConsoleJSONFormat)

	if config.EnableConsole {
		writer := zapcore.Lock(os.Stdout)
		consoleLevel := zap.NewAtomicLevelAt(getZapLevel(config.ConsoleLevel))
		if config.Async {
			var close CloseFunc
			writer, close = Buffer(writer, defaultBufferSize, defaultFlushInterval)

			defers.Register(close)
		}
		core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), writer, consoleLevel)

		cores = append(cores, core)
	}

	if config.EnableFile {
		now := time.Now()
		if config.Name == "" {
			logPath := filepath.Base(config.Dir)
			logName := logPath + "/" + fmt.Sprintf("%04d%02d%02d%02d%02d%02d.log", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
			config.Name = logName
		}
		writer := zapcore.AddSync(newRotate(config))
		if config.Async {
			var close CloseFunc
			writer, close = Buffer(writer, defaultBufferSize, defaultFlushInterval)

			defers.Register(close)
		}
		//还原日志显示级别，日志文件里不显示日志级别颜色
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writer, lv)
		cores = append(cores, core)
	}
	combinedCore := zapcore.NewTee(cores...)

	zapLogger := zap.New(
		combinedCore,
		zapOptions...,
	)
	return &Logger{
		desugar: zapLogger,
		lv:      &lv,
		config:  *config,
		sugar:   zapLogger.Sugar(),
	}
}

func getEncoder(isJSON bool) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.LevelKey = "lv"
	encoderConfig.TimeKey = "ts" // This will change the key from to ts
	//Rest of the code remains same
	if isJSON {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getZapLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}



func (logger *Logger) Suger() *zap.SugaredLogger {
	return logger.sugar
}

func (logger *Logger) Raw() *zap.Logger {
	return logger.desugar
}

// AutoLevel ...
func (logger *Logger) AutoLevel(confKey string) {
	conf.OnChange(func(config *conf.Configuration) {
		lvText := strings.ToLower(config.GetString(confKey))
		if lvText != "" {
			logger.Info("update level", String("level", lvText), String("name", logger.config.Name))
			logger.lv.UnmarshalText([]byte(lvText))
		}
	})
}

// SetLevel ...
func (logger *Logger) SetLevel(lv Level) {
	logger.lv.SetLevel(lv)
}

// Flush ...
func (logger *Logger) Flush() error {
	return logger.desugar.Sync()
}

// DefaultZapConfig ...
func DefaultZapConfig() *zapcore.EncoderConfig {
	return &zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "lv",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder, //timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// DebugEncodeLevel ...
func DebugEncodeLevel(lv zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var colorize = xcolor.Red
	switch lv {
	case zapcore.DebugLevel:
		colorize = xcolor.Blue
	case zapcore.InfoLevel:
		colorize = xcolor.Green
	case zapcore.WarnLevel:
		colorize = xcolor.Yellow
	case zapcore.ErrorLevel, zap.PanicLevel, zap.DPanicLevel, zap.FatalLevel:
		colorize = xcolor.Red
	default:
	}
	enc.AppendString(colorize(lv.CapitalString()))
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(t.Unix())
}

// IsDebugMode ...
func (logger *Logger) IsDebugMode() bool {
	return logger.config.Debug
}

func normalizeMessage(msg string) string {
	return fmt.Sprintf("%-32s", msg)
}

// Debug ...
func (logger *Logger) Debug(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.desugar.Debug(msg, fields...)
}

// Debugw ...
func (logger *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Debugw(msg, keysAndValues...)
}

func sprintf(template string, args ...interface{}) string {
	msg := template
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(template, args...)
	}
	return msg
}

// StdLog ...
func (logger *Logger) StdLog() *log.Logger {
	return zap.NewStdLog(logger.desugar)
}

// Debugf ...
func (logger *Logger) Debugf(template string, args ...interface{}) {
	logger.sugar.Debugw(sprintf(template, args...))
}

// Info ...
func (logger *Logger) Info(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.desugar.Info(msg, fields...)
}

// Infow ...
func (logger *Logger) Infow(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Infow(msg, keysAndValues...)
}

// Infof ...
func (logger *Logger) Infof(template string, args ...interface{}) {
	logger.sugar.Infof(sprintf(template, args...))
}

// Warn ...
func (logger *Logger) Warn(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.desugar.Warn(msg, fields...)
}

// Warnw ...
func (logger *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Warnw(msg, keysAndValues...)
}

// Warnf ...
func (logger *Logger) Warnf(template string, args ...interface{}) {
	logger.sugar.Warnf(sprintf(template, args...))
}

// Error ...
func (logger *Logger) Error(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.desugar.Error(msg, fields...)
}

// Errorw ...
func (logger *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Errorw(msg, keysAndValues...)
}

// Errorf ...
func (logger *Logger) Errorf(template string, args ...interface{}) {
	logger.sugar.Errorf(sprintf(template, args...))
}

// Panic ...
func (logger *Logger) Panic(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		panicDetail(msg, fields...)
		msg = normalizeMessage(msg)
	}
	logger.desugar.Panic(msg, fields...)
}

// Panicw ...
func (logger *Logger) Panicw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Panicw(msg, keysAndValues...)
}

// Panicf ...
func (logger *Logger) Panicf(template string, args ...interface{}) {
	logger.sugar.Panicf(sprintf(template, args...))
}

// DPanic ...
func (logger *Logger) DPanic(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		panicDetail(msg, fields...)
		msg = normalizeMessage(msg)
	}
	logger.desugar.DPanic(msg, fields...)
}

// DPanicw ...
func (logger *Logger) DPanicw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.DPanicw(msg, keysAndValues...)
}

// DPanicf ...
func (logger *Logger) DPanicf(template string, args ...interface{}) {
	logger.sugar.DPanicf(sprintf(template, args...))
}

// Fatal ...
func (logger *Logger) Fatal(msg string, fields ...Field) {
	if logger.IsDebugMode() {
		panicDetail(msg, fields...)
		msg = normalizeMessage(msg)
		return
	}
	logger.desugar.Fatal(msg, fields...)
}

// Fatalw ...
func (logger *Logger) Fatalw(msg string, keysAndValues ...interface{}) {
	if logger.IsDebugMode() {
		msg = normalizeMessage(msg)
	}
	logger.sugar.Fatalw(msg, keysAndValues...)
}

// Fatalf ...
func (logger *Logger) Fatalf(template string, args ...interface{}) {
	logger.sugar.Fatalf(sprintf(template, args...))
}

func panicDetail(msg string, fields ...Field) {
	enc := zapcore.NewMapObjectEncoder()
	for _, field := range fields {
		field.AddTo(enc)
	}

	// 控制台输出
	fmt.Printf("%s: \n    %s: %s\n", xcolor.Red("panic"), xcolor.Red("msg"), msg)
	if _, file, line, ok := runtime.Caller(3); ok {
		fmt.Printf("    %s: %s:%d\n", xcolor.Red("loc"), file, line)
	}
	for key, val := range enc.Fields {
		fmt.Printf("    %s: %s\n", xcolor.Red(key), fmt.Sprintf("%+v", val))
	}

}

// With ...
func (logger *Logger) With(fields ...Field) *Logger {
	desugarLogger := logger.desugar.With(fields...)
	return &Logger{
		desugar: desugarLogger,
		lv:      logger.lv,
		sugar:   desugarLogger.Sugar(),
		config:  logger.config,
	}
}
