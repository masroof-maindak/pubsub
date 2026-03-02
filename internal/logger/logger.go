package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

type LoggerConfig struct {
	InfoLogPath  string
	ErrorLogPath string
}

const (
	FileOpenMode = os.O_RDWR | os.O_CREATE | os.O_APPEND
	FilePerm     = 0666
)

func Initialize(cfg LoggerConfig) error {
	infoFile, err := os.OpenFile(cfg.InfoLogPath, FileOpenMode, FilePerm)
	if err != nil {
		return fmt.Errorf("failed to create info log file at %s: %w", cfg.InfoLogPath, err)
	}

	errFile, err := os.OpenFile(cfg.ErrorLogPath, FileOpenMode, FilePerm)
	if err != nil {
		return fmt.Errorf("failed to create error log file at %s: %w\n", cfg.ErrorLogPath, err)
	}

	writers := []io.Writer{
		&zerolog.FilteredLevelWriter{
			Writer: zerolog.LevelWriterAdapter{Writer: infoFile},
			Level:  zerolog.InfoLevel,
		},

		&zerolog.FilteredLevelWriter{
			Writer: zerolog.LevelWriterAdapter{Writer: errFile},
			Level:  zerolog.ErrorLevel,
		},
	}

	writer := zerolog.MultiLevelWriter(writers...)
	Log = zerolog.New(writer).Level(zerolog.InfoLevel).With().Timestamp().Logger()

	return nil
}
