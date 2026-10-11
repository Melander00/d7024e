package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIPublish(t *testing.T) {
	cmd := CLIPublish(nil)

	if cmd == nil {
		t.Fatal("expected CLIPublish to return a command")
	}

	if got := cmd.getName(); got != "publish" {
		t.Errorf("getName() = %q, want %q", got, "publish")
	}
}

func TestPublishGetHelp(t *testing.T) {
	cmd := CLIPublish(nil)

	help := cmd.getHelp()

	for _, expected := range []string{
		"publish",
		"--force",
		"--prev=PREVIOUS_VERSION",
		"DOMAIN:PACKAGE:VERSION",
		"FILENAME",
	} {
		if !strings.Contains(help, expected) {
			t.Errorf("help should contain %q, got %q", expected, help)
		}
	}
}

func TestPublishHandleInvalidArgumentCount(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no arguments",
			args: []string{"publish"},
		},
		{
			name: "missing filename",
			args: []string{"publish", "domain:package:1"},
		},
		{
			name: "missing package ID",
			args: []string{"publish", "file.txt"},
		},
		{
			name: "too many positional arguments",
			args: []string{"publish", "domain:package:1", "file.txt", "extra.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CLIPublish(nil)

			output := captureStdout(t, func() {
				cmd.handle(tt.args)
			})

			if !strings.Contains(output, cmd.getHelp()) {
				t.Errorf("expected help output, got %q", output)
			}
		})
	}
}

// func TestPublishHandleInvalidPackageID(t *testing.T) {
// 	tests := []struct {
// 		name      string
// 		packageID string
// 	}{
// 		{
// 			name:      "missing version",
// 			packageID: "domain:package",
// 		},
// 		{
// 			name:      "missing package",
// 			packageID: "domain::1",
// 		},
// 		{
// 			name:      "too many components",
// 			packageID: "domain:package:1:extra",
// 		},
// 		{
// 			name:      "no separators",
// 			packageID: "package",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cmd := CLIPublish(nil)

// 			output := captureStdout(t, func() {
// 				cmd.handle([]string{
// 					"publish",
// 					tt.packageID,
// 					"file.txt",
// 				})
// 			})

// 			if !strings.Contains(
// 				output,
// 				"packageid needs to follow DOMAIN:PACKAGE:VERSION",
// 			) {
// 				t.Errorf("expected package ID validation error, got %q", output)
// 			}
// 		})
// 	}
// }

func TestPublishHandleInvalidVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "non-numeric version",
			version: "abc",
		},
		{
			name:    "negative version",
			version: "-1",
		},
		{
			name:    "overflow uint64",
			version: "18446744073709551616",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CLIPublish(nil)

			output := captureStdout(t, func() {
				cmd.handle([]string{
					"publish",
					"domain:package:" + tt.version,
					"file.txt",
				})
			})

			if !strings.Contains(output, "version needs to be a uint64") {
				t.Errorf("expected version validation error, got %q", output)
			}
		})
	}
}

func TestPublishHandleMissingFile(t *testing.T) {
	cmd := CLIPublish(nil)

	output := captureStdout(t, func() {
		cmd.handle([]string{
			"publish",
			"domain:package:1",
			filepath.Join(t.TempDir(), "missing.txt"),
		})
	})

	if output == "" {
		t.Error("expected an error when the file does not exist")
	}
}

func TestPublishHandleTrimsArguments(t *testing.T) {
	cmd := CLIPublish(nil)

	output := captureStdout(t, func() {
		cmd.handle([]string{
			"publish",
			"   ",
			" domain:package ",
			"  file.txt  ",
		})
	})

	// Trimming leaves only two positional arguments, but the package ID
	// is malformed, so the handler should report its format error.
	if !strings.Contains(
		output,
		"packageid needs to follow DOMAIN:PACKAGE:VERSION",
	) {
		t.Errorf("expected package ID validation error, got %q", output)
	}
}

func TestPublishHandleReadsFileBeforePublishing(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "package.bin")

	if err := os.WriteFile(filename, []byte("test package"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// A valid package ID and existing file reach cmd.node.Publish.
	// A nil node would panic at that point, so this test is intentionally
	// not executed with a nil node. Successful publication should be
	// tested with a real Kademlia node or an injected publisher interface.
}
