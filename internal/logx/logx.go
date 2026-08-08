package logx

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wassermanproductions/slate-installer/internal/paths"
)

var (
	mu   sync.Mutex
	file *os.File
)

func Init() error {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		return nil
	}
	f, err := os.OpenFile(paths.TempLog(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	file = f
	_, _ = fmt.Fprintf(file, "\n==== Slate Setup log %s ====\n", time.Now().Format(time.RFC3339))
	return nil
}

func Log(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	line := fmt.Sprintf(format, args...)
	ts := time.Now().Format("15:04:05")
	msg := fmt.Sprintf("[%s] %s\n", ts, line)
	if file != nil {
		_, _ = file.WriteString(msg)
	}
	fmt.Print(msg)
}

func Path() string {
	return paths.TempLog()
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		_ = file.Close()
		file = nil
	}
}
