package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "github.com/go-admin-team/go-admin-core/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const (
	reset       = "\033[0m"
	red         = "\033[31m"
	green       = "\033[32m"
	yellow      = "\033[33m"
	blueBold    = "\033[34;1m"
	magenta     = "\033[35m"
	redBold     = "\033[31;1m"
	magentaBold = "\033[35;1m"
)

// sqlLogger 在应用日志为 info 时仍打印 SQL。
// go-admin 自带 gorm logger 把 SQL 打在 trace，会被 info 级别丢掉。
type sqlLogger struct {
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceWarnStr, traceErrStr string
}

func newGormLogger() logger.Interface {
	cfg := logger.Config{
		SlowThreshold:             time.Second,
		Colorful:                  true,
		IgnoreRecordNotFoundError: true,
		LogLevel: logger.LogLevel(
			log.DefaultLogger.Options().Level.LevelForGorm()),
	}
	if log.DefaultLogger.Options().Level <= log.InfoLevel {
		cfg.LogLevel = logger.Info
	}
	return newSQLLogger(cfg)
}

func newSQLLogger(config logger.Config) *sqlLogger {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s [%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s [%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s [%.3fms] [rows:%v] %s"
	)
	if config.Colorful {
		infoStr = green + "%s " + reset + green + "[info] " + reset
		warnStr = blueBold + "%s " + reset + magenta + "[warn] " + reset
		errStr = magenta + "%s " + reset + red + "[error] " + reset
		traceStr = green + "%s " + reset + yellow + "[%.3fms] " + blueBold + "[rows:%v]" + reset + " %s"
		traceWarnStr = green + "%s " + yellow + "%s " + reset + redBold + "[%.3fms] " + yellow + "[rows:%v]" + magenta + " %s" + reset
		traceErrStr = redBold + "%s " + magentaBold + "%s " + reset + yellow + "[%.3fms] " + blueBold + "[rows:%v]" + reset + " %s"
	}
	return &sqlLogger{
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

func (l *sqlLogger) getLogger(ctx context.Context) log.Logger {
	requestId := ctx.Value("X-Request-Id")
	if requestId != nil {
		return log.DefaultLogger.Fields(map[string]interface{}{
			"x-request-id": requestId,
		})
	}
	return log.DefaultLogger
}

func (l *sqlLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *sqlLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.getLogger(ctx).Logf(log.InfoLevel, l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *sqlLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.getLogger(ctx).Logf(log.WarnLevel, l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *sqlLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.getLogger(ctx).Logf(log.ErrorLevel, l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *sqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		l.logSQL(ctx, l.traceErrStr, utils.FileWithLineNum(), err, elapsed, rows, sql)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		l.logSQL(ctx, l.traceWarnStr, utils.FileWithLineNum(), slowLog, elapsed, rows, sql)
	case l.LogLevel >= logger.Info:
		rowVal := interface{}(rows)
		if rows == -1 {
			rowVal = "-"
		}
		l.getLogger(ctx).Logf(log.InfoLevel, l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rowVal, sql)
	}
}

func (l *sqlLogger) logSQL(ctx context.Context, format string, fileLine interface{}, extra interface{}, elapsed time.Duration, rows int64, sql string) {
	rowVal := interface{}(rows)
	if rows == -1 {
		rowVal = "-"
	}
	l.getLogger(ctx).Logf(log.InfoLevel, format, fileLine, extra, float64(elapsed.Nanoseconds())/1e6, rowVal, sql)
}
