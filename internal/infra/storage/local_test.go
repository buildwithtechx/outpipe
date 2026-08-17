package storage

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestLocalStorePutOpenDelete(t *testing.T) {
	store, err := NewLocal(Config{RootDir: t.TempDir()})

	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	object, err := store.Put(context.Background(), "tunnels/one.txt", strings.NewReader("hello"), "text/plain")

	if err != nil {
		t.Fatalf("put object: %v", err)
	}

	if object.Size != 5 || object.ContentType != "text/plain" {
		t.Fatalf("unexpected object metadata: %+v", object)
	}

	file, err := store.Open(object.Key)

	if err != nil {
		t.Fatalf("open object: %v", err)
	}

	content, err := io.ReadAll(file)
	file.Close()

	if err != nil {
		t.Fatalf("read object: %v", err)
	}

	if string(content) != "hello" {
		t.Fatalf("unexpected object content: %s", content)
	}

	if err := store.Delete(object.Key); err != nil {
		t.Fatalf("delete object: %v", err)
	}
}

func TestLocalStoreRejectsTraversal(t *testing.T) {
	store, err := NewLocal(Config{RootDir: t.TempDir()})

	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	if _, err := store.Open("../outside"); err == nil {
		t.Fatal("expected traversal key to be rejected")
	}
}
