package log

import (
	"io"
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	Init(ConfigWithWriters([]io.Writer{
		os.Stdout,
		NewWriterWithAge("./tmp_age.log", SplitByAgeWithRotationTime("2s")),
		NewWriterWithSize("./tmp_size.log", SplitBySizeWithMaxSize(1)),
	}))

	Info("info msg")
	Debug("debug msg")

	Init(ConfigWithEncoder("json"))

	Info("info msg")
	Debug("debug msg")

}
