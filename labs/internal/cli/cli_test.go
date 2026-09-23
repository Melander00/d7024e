package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

type T = testing.T

func TestExitCommand(t *T) {
	cmd := CLIExit(nil)
	cmd.getName()
	cmd.getHelp()
	// We dont do .handle() since it will cause exit which cannot be tested
}

type fakeKademlia struct {
	data []byte
	err  error

	lookupHash string
	storedData []byte
}

func (f *fakeKademlia) LookupData(hash string) ([]byte, error) {
	f.lookupHash = hash
	return f.data, f.err
}

func (f *fakeKademlia) Store(data []byte) {
	f.storedData = data
}

func TestGetCommandPrint(t *T) {
	data := []byte("hello")

	fakeNode := &fakeKademlia{
		data: data,
	}

	cmd := CLIGet(fakeNode)
	cmd.getName()
	cmd.getHelp()

	cmd.handle([]string{"get", "some-key"})

	if fakeNode.lookupHash != "some-key" {
		t.Fatalf("expected lookup of some-key, got %s", fakeNode.lookupHash)
	}
}

func TestGetCommandFilename(t *T) {
	data := []byte("hello world")

	fakeNode := &fakeKademlia{
		data: data,
	}

	cmd := CLIGet(fakeNode)
	cmd.getName()
	cmd.getHelp()

	dir := t.TempDir()
	filename := filepath.Join(dir, "output.txt")

	cmd.handle([]string{"get", "some-key", filename})

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", string(data))
	}
}

func TestPingCommand(t *T) {
	cmd := CLIPing(nil)
	cmd.getName()
	cmd.getHelp()
	// TODO: Test .handle()
}

func TestPutCmdHandle(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.txt")

	expected := []byte("hello world")

	if err := os.WriteFile(filename, expected, 0644); err != nil {
		t.Fatal(err)
	}

	node := &fakeKademlia{}
	cmd := CLIPut(node)
	cmd.getName()
	cmd.getHelp()
	formatBytes([]byte("1"))

	cmd.handle([]string{"put", filename})

	if !bytes.Equal(node.storedData, expected) {
		t.Fatalf(
			"expected stored data %q, got %q",
			expected,
			node.storedData,
		)
	}
}

func TestPutCmdHandleWithoutFilename(t *testing.T) {
	node := &fakeKademlia{}
	cmd := CLIPut(node)

	cmd.handle([]string{"put"})

	if node.storedData != nil {
		t.Fatal("expected Store not to be called")
	}
}

func TestPutCmdHandleMissingFile(t *testing.T) {
	node := &fakeKademlia{}
	cmd := CLIPut(node)

	cmd.handle([]string{"put", "/does/not/exist"})

	if node.storedData != nil {
		t.Fatal("expected Store not to be called")
	}
}
