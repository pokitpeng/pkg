package log

import (
	"testing"
)

func TestLog(t *testing.T) {
	// Init(ConfigWithWriters([]io.Writer{
	// 	os.Stdout,
	// 	// utils.NewWriterWithAge("./tmp.log", utils.SplitByAgeWithRotationTime("2s")),
	// 	// utils.NewWriterWithSize("./tmp.log", utils.SplitBySizeWithMaxSize(1)),
	// }))

	Info("info msg")
	Debug("debug msg")

	Init(ConfigWithEncoder("console"))

	Info("info msg")
	Debug("debug msg")
}
