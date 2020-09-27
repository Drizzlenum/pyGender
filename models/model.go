package models
import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)
type RespData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *Data  `json:"data"`
}
type Data struct {
	Result *Result `json:"result"`
	Status int     `json:"status"`
}
type Result struct {
	Age    Age    `json:"age"`
	Gender Gender `json:"gender"`
}
type Age struct {
	AgeType string `json:"age_type"`
	Child   string `json:"child"`
	Middle  string `json:"middle"`
	Old     string `json:"old"`
}
type Gender struct {
	Female      string `json:"female"`
	Gender_type string `json:"gender_type"`
	Male        string `json:"male"`
}
var GlobalConfig *viper.Viper
var Logger *zap.Logger

//日志类初始化方法
func InitLogger() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
	}

	// 设置日志级别
	atom := zap.NewAtomicLevelAt(zap.DebugLevel)

	lp := GlobalConfig.GetString("Log.Path")
	config := zap.Config{
		Level:            atom,                                                // 日志级别
		Development:      true,                                                // 开发模式，堆栈跟踪
		Encoding:         "json",                                              // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,                                       // 编码器配置
	//	InitialFields:    map[string]interface{}{"serviceName": "spikeProxy"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout", lp},         // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},
	}

	// 构建日志
	var err error
	Logger, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("log 初始化失败: %v", err))
	}
	Logger.Info("log 初始化成功")

}

