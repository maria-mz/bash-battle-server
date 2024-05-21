package log

import (
	"os"

	"github.com/charmbracelet/log"
)

var Logger *log.Logger

func InitLogger() {
	styles := log.DefaultStyles()
	levelStyles := newLogStyles()

	styles.Levels[log.DebugLevel] = levelStyles.DebugStyle
	styles.Levels[log.InfoLevel] = levelStyles.InfoStyle
	styles.Levels[log.WarnLevel] = levelStyles.WarnStyle
	styles.Levels[log.ErrorLevel] = levelStyles.ErrorStyle
	styles.Levels[log.FatalLevel] = levelStyles.FatalStyle

	Logger = log.New(os.Stderr)
	Logger.SetStyles(styles)
	Logger.SetReportTimestamp(true)
	Logger.SetReportCaller(true)
	Logger.SetLevel(log.InfoLevel)
}
