package file_history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRewindDeletesFileCreatedAfterTargetSnapshot(t *testing.T) {
	base := t.TempDir()
	projectDir := filepath.Join(base, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	h := New(base, "session-1")

	// 第一轮：没有任何文件改动，纯对话，打一个快照。
	h.MakeSnapshot(0, "第一轮")

	// 第二轮：新建一个文件。TrackEdit 在写入前调用，此时文件还不存在。
	newFile := filepath.Join(projectDir, "new_file.go")
	h.TrackEdit(newFile)
	if err := os.WriteFile(newFile, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	h.MakeSnapshot(2, "第二轮：新建文件")

	if _, err := os.Stat(newFile); err != nil {
		t.Fatalf("newFile should exist before rewind: %v", err)
	}

	// 回滚到第一轮的快照，也就是这个文件创建之前的状态。
	changed, err := h.Rewind(0)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(newFile); !os.IsNotExist(err) {
		t.Fatalf("回滚到文件创建之前，文件应该被删除，但仍然存在")
	}

	found := false
	for _, c := range changed {
		if c == newFile {
			found = true
		}
	}
	if !found {
		t.Fatalf("changed 列表里应该包含被删除的文件，got %v", changed)
	}
}

func TestRewindRestoresEditOnExistingFile(t *testing.T) {
	base := t.TempDir()
	projectDir := filepath.Join(base, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	h := New(base, "session-1")

	existing := filepath.Join(projectDir, "existing.go")
	if err := os.WriteFile(existing, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}

	h.TrackEdit(existing)
	h.MakeSnapshot(0, "第一轮：修改前的快照")

	if err := os.WriteFile(existing, []byte("modified"), 0o644); err != nil {
		t.Fatal(err)
	}
	h.MakeSnapshot(2, "第二轮：改了内容")

	changed, err := h.Rewind(0)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Fatalf("expected 'original', got %q", string(data))
	}
	if len(changed) != 1 || changed[0] != existing {
		t.Fatalf("expected changed=[%s], got %v", existing, changed)
	}
}

func TestRewindToLatestSnapshotKeepsCreatedFile(t *testing.T) {
	base := t.TempDir()
	projectDir := filepath.Join(base, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	h := New(base, "session-1")

	h.MakeSnapshot(0, "第一轮")

	newFile := filepath.Join(projectDir, "new_file.go")
	h.TrackEdit(newFile)
	if err := os.WriteFile(newFile, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	h.MakeSnapshot(2, "第二轮：新建文件")

	// 回滚到文件创建之后的这个快照本身，文件应该保留（内容还原成当时写入的内容）。
	if _, err := h.Rewind(1); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(newFile)
	if err != nil {
		t.Fatalf("newFile should still exist: %v", err)
	}
	if string(data) != "package main" {
		t.Fatalf("expected 'package main', got %q", string(data))
	}
}
