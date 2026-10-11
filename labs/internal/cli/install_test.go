package cli

import (
	"strings"
	"testing"
)

// func captureStdout(t *testing.T, fn func()) string {
// 	t.Helper()

// 	oldStdout := os.Stdout
// 	r, w, err := os.Pipe()
// 	if err != nil {
// 		t.Fatalf("failed to create pipe: %v", err)
// 	}

// 	os.Stdout = w
// 	fn()

// 	if err := w.Close(); err != nil {
// 		t.Fatalf("failed to close writer: %v", err)
// 	}
// 	os.Stdout = oldStdout

// 	var buf bytes.Buffer
// 	if _, err := io.Copy(&buf, r); err != nil {
// 		t.Fatalf("failed to read stdout: %v", err)
// 	}
// 	if err := r.Close(); err != nil {
// 		t.Fatalf("failed to close reader: %v", err)
// 	}

// 	return buf.String()
// }

func TestCLIInstall(t *testing.T) {
	cmd := CLIInstall(nil)

	if cmd == nil {
		t.Fatal("expected CLIInstall to return a command")
	}

	if cmd.getName() != "install" {
		t.Errorf("expected command name %q, got %q", "install", cmd.getName())
	}
}

func TestInstallGetHelp(t *testing.T) {
	cmd := CLIInstall(nil)

	help := cmd.getHelp()

	if !strings.Contains(help, "DOMAIN:PACKAGE:VERSION") {
		t.Errorf("help should contain package format, got %q", help)
	}

	if !strings.Contains(help, "FILENAME") {
		t.Errorf("help should contain FILENAME, got %q", help)
	}
}

func TestInstallHandleInvalidArgumentCount(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no arguments",
			args: []string{"install"},
		},
		{
			name: "missing package ID",
			args: []string{},
		},
		{
			name: "too many arguments",
			args: []string{"install", "domain:package:latest", "extra"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CLIInstall(nil)

			output := captureStdout(t, func() {
				cmd.handle(tt.args)
			})

			if !strings.Contains(output, cmd.getHelp()) {
				t.Errorf("expected help output, got %q", output)
			}
		})
	}
}

func TestInstallHandleInvalidPackageID(t *testing.T) {
	tests := []struct {
		name      string
		packageID string
	}{
		{
			name:      "missing version",
			packageID: "domain:package",
		},
		{
			name:      "missing package",
			packageID: "domain::latest",
		},
		{
			name:      "too many components",
			packageID: "domain:package:1:extra",
		},
		{
			name:      "no separators",
			packageID: "package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CLIInstall(nil)

			_ = captureStdout(t, func() {
				cmd.handle([]string{"install", tt.packageID})
			})
		})
	}
}
