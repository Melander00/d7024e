package cli

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

type mockCLICommand struct {
	name string
	help string

	called bool
	args   []string
}

func (m *mockCLICommand) handle(args []string) {
	m.called = true
	m.args = args
}

func (m *mockCLICommand) getName() string {
	return m.name
}

func (m *mockCLICommand) getHelp() string {
	return m.help
}

func TestCLIRegisterCommand(t *testing.T) {
	cli := &CLI{
		cmds: make(map[string]CLICommand),
	}

	cmd := &mockCLICommand{
		name: "test",
		help: "test command",
	}

	cli.registerCommand(cmd)

	got, exists := cli.cmds["test"]

	if !exists {
		t.Fatal("expected command to be registered")
	}

	if got != cmd {
		t.Fatal("registered command is not the same command")
	}
}

func TestCLIRunCLIDispatchesCommand(t *testing.T) {
	oldStdin := os.Stdin
	defer func() {
		os.Stdin = oldStdin
	}()

	stdin, err := os.CreateTemp("", "cli-test-stdin")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(stdin.Name())

	if _, err := stdin.WriteString("test hello world\n"); err != nil {
		t.Fatal(err)
	}

	if _, err := stdin.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	os.Stdin = stdin

	cmd := &mockCLICommand{
		name: "test",
		help: "test command",
	}

	cli := &CLI{
		ip:   "127.0.0.1",
		cmds: make(map[string]CLICommand),
	}

	cli.registerCommand(cmd)

	cli.runCLI()

	if !cmd.called {
		t.Fatal("expected command to be called")
	}

	expected := []string{"test", "hello", "world"}

	if !reflect.DeepEqual(cmd.args, expected) {
		t.Fatalf(
			"expected args %v, got %v",
			expected,
			cmd.args,
		)
	}
}

func TestCLIRunCLIIgnoresEmptyInput(t *testing.T) {
	restoreStdin := setStdin(t, "\n\n")
	defer restoreStdin()

	cmd := &mockCLICommand{
		name: "test",
		help: "test command",
	}

	cli := &CLI{
		ip:   "127.0.0.1",
		cmds: make(map[string]CLICommand),
	}

	cli.registerCommand(cmd)

	cli.runCLI()

	if cmd.called {
		t.Fatal("expected command not to be called")
	}
}

func setStdin(t *testing.T, input string) func() {
	t.Helper()

	oldStdin := os.Stdin

	file, err := os.CreateTemp("", "cli-stdin-*")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		file.Close()
		os.Remove(file.Name())
		os.Stdin = oldStdin
	})

	if _, err := file.WriteString(input); err != nil {
		t.Fatal(err)
	}

	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	os.Stdin = file

	return func() {
		os.Stdin = oldStdin
	}
}

func TestCLIRunCLIHelp(t *testing.T) {
	restoreStdin := setStdin(t, "help\n")
	defer restoreStdin()

	cli := &CLI{
		ip:   "127.0.0.1",
		cmds: make(map[string]CLICommand),
	}

	cli.registerCommand(&mockCLICommand{
		name: "zebra",
		help: "zebra command",
	})

	cli.registerCommand(&mockCLICommand{
		name: "alpha",
		help: "alpha command",
	})

	output := captureStdout(t, func() {
		cli.runCLI()
	})

	if !strings.Contains(output, "alpha") {
		t.Fatalf("expected alpha command in help, got:\n%s", output)
	}

	if !strings.Contains(output, "zebra") {
		t.Fatalf("expected zebra command in help, got:\n%s", output)
	}
}

func TestCLIRunCLIIgnoresUnknownCommand(t *testing.T) {
	restoreStdin := setStdin(t, "doesnotexist hello\n")
	defer restoreStdin()

	cmd := &mockCLICommand{
		name: "test",
		help: "test command",
	}

	cli := &CLI{
		ip:   "127.0.0.1",
		cmds: make(map[string]CLICommand),
	}

	cli.registerCommand(cmd)

	cli.runCLI()

	if cmd.called {
		t.Fatal("expected registered command not to be called")
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}

	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.String()
}
