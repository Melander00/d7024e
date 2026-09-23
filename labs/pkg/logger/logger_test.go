package logger

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileLogger(t *testing.T) {
	logger := NewFileLogger("test.log")

	if logger.file != "test.log" {
		t.Errorf("expected file %q, got %q", "test.log", logger.file)
	}
}

func TestFileLogger_Log(t *testing.T) {
	file := filepath.Join(t.TempDir(), "test.log")
	logger := NewFileLogger(file)

	logger.Log("hello")

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if got, want := string(content), "hello"; got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestFileLogger_Error(t *testing.T) {
	file := filepath.Join(t.TempDir(), "test.log")
	logger := NewFileLogger(file)

	errToLog := errors.New("something went wrong")
	logger.Error(errToLog)

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if got, want := string(content), "something went wrong"; got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestFileLogger_Warn(t *testing.T) {
	file := filepath.Join(t.TempDir(), "test.log")
	logger := NewFileLogger(file)

	logger.Warn("be careful")

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if got, want := string(content), "be careful"; got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestFileLogger_AppendsMessages(t *testing.T) {
	file := filepath.Join(t.TempDir(), "test.log")
	logger := NewFileLogger(file)

	logger.Log("first")
	logger.Log("second")

	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if got, want := string(content), "firstsecond"; got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
