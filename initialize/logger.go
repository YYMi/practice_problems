package initialize

import (
	"io"
	"os"
	"path/filepath"
	"practice_problems/global"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 定义颜色代码
const (
	colorRed     = "\x1b[31m"
	colorGreen   = "\x1b[32m"
	colorYellow  = "\x1b[33m"
	colorBlue    = "\x1b[34m"
	colorMagenta = "\x1b[35m"
	colorCyan    = "\x1b[36m"
	colorWhite   = "\x1b[97m"
	colorReset   = "\x1b[0m"
)

// =================================================================
// ★★★ 核心黑科技：自定义 Core，用于给 Message 上色 ★★★
// =================================================================
type ColorMessageCore struct {
	zapcore.Core
}

// 覆写 Check 方法，确保使用我们的 Core
func (c *ColorMessageCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

// 覆写 Write 方法，在这里修改 Message 的颜色
func (c *ColorMessageCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// 根据等级给 Message 加颜色
	var messageColor string
	switch entry.Level {
	case zapcore.DebugLevel:
		messageColor = colorWhite // Debug 消息: 白色
	case zapcore.InfoLevel:
		messageColor = colorBlue // Info 消息: 白色 (标准 Java 风格)
	case zapcore.WarnLevel:
		messageColor = colorYellow // Warn 消息: 黄色 (你要的效果)
	case zapcore.ErrorLevel, zapcore.FatalLevel, zapcore.PanicLevel:
		messageColor = colorRed // Error 消息: 红色 (你要的效果)
	default:
		messageColor = colorWhite
	}

	// 强制修改消息内容：颜色 + 原消息 + 重置
	entry.Message = messageColor + entry.Message + colorReset

	// 传给原本的 Core 去打印
	return c.Core.Write(entry, fields)
}

// =================================================================

func InitLogger() {
	logDir := "log"
	logFileName := "practice_problems.log"
	logPath := filepath.Join(logDir, logFileName)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		_ = os.MkdirAll(logDir, 0755)
	}

	hook := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    50,
		MaxBackups: 200,
		MaxAge:     180,
		Compress:   true,
	}

	// 1. 基础配置 (文件用，纯净无色)
	fileConfig := zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		ConsoleSeparator: " ",

		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString("[" + t.Format("2006-01-02 15:04:05.000"+"]"))
		},
		EncodeLevel: func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString("[" + l.CapitalString() + "]")
		},
		EncodeName: func(name string, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString("[" + name + "]")
		},
		EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(caller.TrimmedPath() + " :")
		},
	}

	// 2. 控制台配置 (头部带颜色)
	consoleConfig := fileConfig
	consoleConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(colorGreen + "[" + t.Format("2006-01-02 15:04:05.000") + "]" + colorReset)
	}
	consoleConfig.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		var color string
		switch l {
		case zapcore.DebugLevel:
			color = colorWhite
		case zapcore.InfoLevel:
			color = colorBlue // Info 标签用蓝色/绿色
		case zapcore.WarnLevel:
			color = colorYellow
		case zapcore.ErrorLevel, zapcore.FatalLevel:
			color = colorRed
		default:
			color = colorWhite
		}
		// 补齐空格对齐
		s := l.CapitalString()
		if len(s) == 4 {
			s += " "
		}
		enc.AppendString(color + "[" + s + "]" + colorReset)
	}
	consoleConfig.EncodeName = func(name string, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(colorMagenta + "[" + name + "]" + colorReset)
	}
	consoleConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(colorCyan + caller.TrimmedPath() + colorReset + " :")
	}

	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.InfoLevel)

	// ==========================================
	// 组装 Core
	// ==========================================

	// 1. 创建原始控制台 Core
	baseConsoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleConfig),
		zapcore.AddSync(os.Stdout),
		atomicLevel,
	)

	// ★★★ 重点：用我们的 ColorMessageCore 包裹它 ★★★
	// 这样消息内容就会根据等级变色了
	consoleCore := &ColorMessageCore{
		Core: baseConsoleCore,
	}

	// 2. 文件 Core (保持纯净，不包裹)
	fileCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(fileConfig),
		zapcore.AddSync(hook),
		atomicLevel,
	)

	// 3. 组合
	core := zapcore.NewTee(consoleCore, fileCore)
	logger := zap.New(core, zap.AddCaller())
	global.Log = logger.Sugar()

	ginOutput := io.MultiWriter(os.Stdout, hook)
	gin.DefaultWriter = ginOutput
	gin.DefaultErrorWriter = ginOutput

	global.Log.Info("✅ 终极彩色日志系统已启动")
	global.Log.Warn("这是一条警告信息 (应该是黄色的)")
	global.Log.Error("这是一条错误信息 (应该是红色的)")
}
