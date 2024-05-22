package log

import (
	"os"

	"github.com/charmbracelet/log"
)

var Logger *log.Logger

func InitLogger() {
	styles := log.DefaultStyles()
	newStyles := newLogStyles()

	styles.Levels[log.DebugLevel] = newStyles.DebugStyle
	styles.Levels[log.InfoLevel] = newStyles.InfoStyle
	styles.Levels[log.WarnLevel] = newStyles.WarnStyle
	styles.Levels[log.ErrorLevel] = newStyles.ErrorStyle
	styles.Levels[log.FatalLevel] = newStyles.FatalStyle

	Logger = log.New(os.Stderr)
	Logger.SetStyles(styles)
	Logger.SetReportTimestamp(true)
	Logger.SetReportCaller(true)
	Logger.SetLevel(log.InfoLevel)
}
