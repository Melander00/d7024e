package logger

import (
	"errors"
	"io"
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

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stdout = writer

	fn()

	writer.Close()
	os.Stdout = oldStdout

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}

	reader.Close()

	return string(output)
}

func TestNewPrintLogger(t *testing.T) {
	logger := NewPrintLogger("my-app")

	if logger.prefix != "my-app" {
		t.Errorf("expected prefix %q, got %q", "my-app", logger.prefix)
	}
}

func TestPrintLogger_Log(t *testing.T) {
	logger := NewPrintLogger("my-app")

	output := captureOutput(t, func() {
		logger.Log("hello")
	})

	expected := "[INFO] my-app: hello\n"

	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestPrintLogger_Error(t *testing.T) {
	logger := NewPrintLogger("my-app")
	errToLog := errors.New("something went wrong")

	output := captureOutput(t, func() {
		logger.Error(errToLog)
	})

	expected := "[ERROR] my-app: something went wrong\n"

	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestPrintLogger_Warn(t *testing.T) {
	logger := NewPrintLogger("my-app")

	output := captureOutput(t, func() {
		logger.Warn("be careful")
	})

	expected := "[WARN] my-app: be careful\n"

	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
