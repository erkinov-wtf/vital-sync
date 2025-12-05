package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/erkinov-wtf/vital-sync/internal/constants"
)

// fileWriter handles concurrent writes to the log file
type fileWriter struct {
	mu       sync.Mutex
	logDir   string
	currFile *os.File
	currDay  string
}

// Write implements io.Writer and handles rotating files daily
func (fw *fileWriter) Write(p []byte) (n int, err error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	today := time.Now().Local().Format(constants.DateFormat)
	if fw.currFile == nil || fw.currDay != today {
		if fw.currFile != nil {
			_ = fw.currFile.Close()
		}

		if err = os.MkdirAll(fw.logDir, 0755); err != nil {
			return 0, err
		}

		filename := fmt.Sprintf("vital_sync_%v.log", today)
		filePath := filepath.Join(fw.logDir, filename)

		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}

		fw.currFile = file
		fw.currDay = today
	}

	return fw.currFile.Write(p)
}
