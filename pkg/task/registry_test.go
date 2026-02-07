package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegistryAdd(t *testing.T) {
	reg := NewRegistry()

	task := New("ua-001", "Implement OAuth")
	err := reg.Add(task)
	if err != nil {
		t.Fatalf("failed to add task: %v", err)
	}

	// Verify it's in the registry
	got, err := reg.Get("ua-001")
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if got.ID != "ua-001" {
		t.Errorf("expected ID 'ua-001', got '%s'", got.ID)
	}
}

func TestRegistryAddDuplicate(t *testing.T) {
	reg := NewRegistry()

	task1 := New("ua-001", "First")
	reg.Add(task1)

	task2 := New("ua-001", "Duplicate")
	err := reg.Add(task2)
	if err == nil {
		t.Error("expected error for duplicate ID")
	}
}

func TestRegistryAddInvalidTask(t *testing.T) {
	reg := NewRegistry()

	task := &Task{ID: "", Title: "No ID"}
	err := reg.Add(task)
	if err == nil {
		t.Error("expected error for invalid task")
	}
}

func TestRegistryAddWithInvalidDeps(t *testing.T) {
	reg := NewRegistry()

	task := New("ua-002", "Depends on non-existent")
	task.Deps = []string{"ua-001"} // Doesn't exist

	err := reg.Add(task)
	if err == nil {
		t.Error("expected error for invalid dependency")
	}
}

func TestRegistryAddWithValidDeps(t *testing.T) {
	reg := NewRegistry()

	// Add dependency first
	dep := New("ua-001", "Dependency")
	reg.Add(dep)

	// Now add task that depends on it
	task := New("ua-002", "Depends on ua-001")
	task.Deps = []string{"ua-001"}

	err := reg.Add(task)
	if err != nil {
		t.Fatalf("failed to add task with valid deps: %v", err)
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	reg := NewRegistry()

	_, err := reg.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent task")
	}
}

func TestRegistryUpdate(t *testing.T) {
	reg := NewRegistry()

	task := New("ua-001", "Original")
	reg.Add(task)

	task.Title = "Updated"
	err := reg.Update(task)
	if err != nil {
		t.Fatalf("failed to update: %v", err)
	}

	got, _ := reg.Get("ua-001")
	if got.Title != "Updated" {
		t.Errorf("expected title 'Updated', got '%s'", got.Title)
	}
}

func TestRegistryUpdateNotFound(t *testing.T) {
	reg := NewRegistry()

	task := New("ua-001", "Does not exist")
	err := reg.Update(task)
	if err == nil {
		t.Error("expected error for updating nonexistent task")
	}
}

func TestRegistryDelete(t *testing.T) {
	reg := NewRegistry()

	task := New("ua-001", "To delete")
	reg.Add(task)

	err := reg.Delete("ua-001")
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	_, err = reg.Get("ua-001")
	if err == nil {
		t.Error("expected task to be deleted")
	}
}

func TestRegistryDeleteWithDependents(t *testing.T) {
	reg := NewRegistry()

	// Add tasks with dependency
	dep := New("ua-001", "Dependency")
	reg.Add(dep)

	task := New("ua-002", "Depends on ua-001")
	task.Deps = []string{"ua-001"}
	reg.Add(task)

	// Try to delete the dependency
	err := reg.Delete("ua-001")
	if err == nil {
		t.Error("expected error when deleting task with dependents")
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()

	reg.Add(New("ua-001", "First"))
	reg.Add(New("ua-002", "Second"))
	reg.Add(New("ua-003", "Third"))

	tasks := reg.List()
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestRegistryListByStatus(t *testing.T) {
	reg := NewRegistry()

	t1 := New("ua-001", "Pending")
	reg.Add(t1)

	t2 := New("ua-002", "In Progress")
	reg.Add(t2)
	t2.SetStatus(StatusInProgress)
	reg.Update(t2)

	t3 := New("ua-003", "Also Pending")
	reg.Add(t3)

	pending := reg.ListByStatus(StatusPending)
	if len(pending) != 2 {
		t.Errorf("expected 2 pending tasks, got %d", len(pending))
	}

	inProgress := reg.ListByStatus(StatusInProgress)
	if len(inProgress) != 1 {
		t.Errorf("expected 1 in_progress task, got %d", len(inProgress))
	}
}

func TestRegistryListByRepo(t *testing.T) {
	reg := NewRegistry()

	t1 := New("ua-001", "Android task")
	t1.Repo = "android"
	reg.Add(t1)

	t2 := New("ua-002", "iOS task")
	t2.Repo = "ios"
	reg.Add(t2)

	t3 := New("ua-003", "Another Android task")
	t3.Repo = "android"
	reg.Add(t3)

	android := reg.ListByRepo("android")
	if len(android) != 2 {
		t.Errorf("expected 2 android tasks, got %d", len(android))
	}
}

func TestRegistryGetReady(t *testing.T) {
	reg := NewRegistry()

	// Task with no deps - should be ready
	t1 := New("ua-001", "No deps")
	reg.Add(t1)

	// Task with incomplete dep - not ready
	t2 := New("ua-002", "Has dep")
	t2.Deps = []string{"ua-001"}
	reg.Add(t2)

	ready := reg.GetReady()
	if len(ready) != 1 {
		t.Errorf("expected 1 ready task, got %d", len(ready))
	}
	if ready[0].ID != "ua-001" {
		t.Errorf("expected ua-001 to be ready, got %s", ready[0].ID)
	}

	// Complete the dependency
	t1.SetStatus(StatusInProgress)
	reg.Update(t1)
	t1.SetStatus(StatusComplete)
	reg.Update(t1)

	// Now t2 should be ready
	ready = reg.GetReady()
	if len(ready) != 1 {
		t.Errorf("expected 1 ready task after dep complete, got %d", len(ready))
	}
	if ready[0].ID != "ua-002" {
		t.Errorf("expected ua-002 to be ready, got %s", ready[0].ID)
	}
}

func TestRegistryGetDeps(t *testing.T) {
	reg := NewRegistry()

	t1 := New("ua-001", "Dep 1")
	t2 := New("ua-002", "Dep 2")
	reg.Add(t1)
	reg.Add(t2)

	t3 := New("ua-003", "Has deps")
	t3.Deps = []string{"ua-001", "ua-002"}
	reg.Add(t3)

	deps, err := reg.GetDeps("ua-003")
	if err != nil {
		t.Fatalf("failed to get deps: %v", err)
	}
	if len(deps) != 2 {
		t.Errorf("expected 2 deps, got %d", len(deps))
	}
}

func TestRegistryGetDependents(t *testing.T) {
	reg := NewRegistry()

	t1 := New("ua-001", "Base")
	reg.Add(t1)

	t2 := New("ua-002", "Depends on base")
	t2.Deps = []string{"ua-001"}
	reg.Add(t2)

	t3 := New("ua-003", "Also depends on base")
	t3.Deps = []string{"ua-001"}
	reg.Add(t3)

	dependents, err := reg.GetDependents("ua-001")
	if err != nil {
		t.Fatalf("failed to get dependents: %v", err)
	}
	if len(dependents) != 2 {
		t.Errorf("expected 2 dependents, got %d", len(dependents))
	}
}

func TestRegistryCircularDependency(t *testing.T) {
	reg := NewRegistry()

	// Create circular: A -> B -> C -> A
	tA := New("ua-A", "A")
	reg.Add(tA)

	tB := New("ua-B", "B")
	tB.Deps = []string{"ua-A"}
	reg.Add(tB)

	tC := New("ua-C", "C")
	tC.Deps = []string{"ua-B"}
	reg.Add(tC)

	// Try to make A depend on C (creates cycle)
	tA.Deps = []string{"ua-C"}
	err := reg.Update(tA)
	if err == nil {
		t.Error("expected error for circular dependency")
	}
}

func TestRegistrySaveLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "tasks.json")

	// Create and save registry
	reg := NewRegistry()
	reg.Add(New("ua-001", "First"))

	t2 := New("ua-002", "Second")
	t2.Deps = []string{"ua-001"}
	reg.Add(t2)

	err := reg.Save(filePath)
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("save file not created")
	}

	// Load into new registry
	reg2 := NewRegistry()
	err = reg2.Load(filePath)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	// Verify contents
	tasks := reg2.List()
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}

	task2, _ := reg2.Get("ua-002")
	if len(task2.Deps) != 1 || task2.Deps[0] != "ua-001" {
		t.Error("deps not preserved after load")
	}
}

func TestRegistryConcurrentReads(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "tasks.json")

	// Create and save initial registry
	reg := NewRegistry()
	for i := 0; i < 10; i++ {
		task := New(fmt.Sprintf("ua-%03d", i), fmt.Sprintf("Task %d", i))
		reg.Add(task)
	}
	if err := reg.Save(filePath); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Perform concurrent reads
	const numReaders = 10
	done := make(chan bool, numReaders)
	errors := make(chan error, numReaders)

	for i := 0; i < numReaders; i++ {
		go func() {
			regReader := NewRegistry()
			if err := regReader.Load(filePath); err != nil {
				errors <- err
				done <- false
				return
			}
			if len(regReader.List()) != 10 {
				errors <- fmt.Errorf("expected 10 tasks, got %d", len(regReader.List()))
				done <- false
				return
			}
			done <- true
		}()
	}

	// Wait for all readers
	for i := 0; i < numReaders; i++ {
		<-done
	}

	close(errors)
	for err := range errors {
		t.Errorf("concurrent read error: %v", err)
	}
}

func TestRegistryConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "tasks.json")

	// Create initial empty registry
	reg := NewRegistry()
	if err := reg.Save(filePath); err != nil {
		t.Fatalf("failed to create initial file: %v", err)
	}

	// Perform sequential writes from different registry instances
	// This simulates separate processes writing to the same file
	const numWrites = 5
	errors := make(chan error, numWrites)

	for i := 0; i < numWrites; i++ {
		taskID := fmt.Sprintf("ua-%03d", i)
		
		// Load fresh registry
		regWriter := NewRegistry()
		if err := regWriter.Load(filePath); err != nil {
			errors <- fmt.Errorf("write %d load failed: %w", i, err)
			continue
		}

		// Add task
		task := New(taskID, fmt.Sprintf("Task %d", i))
		if err := regWriter.Add(task); err != nil {
			errors <- fmt.Errorf("write %d add failed: %w", i, err)
			continue
		}

		// Save
		if err := regWriter.Save(filePath); err != nil {
			errors <- fmt.Errorf("write %d save failed: %w", i, err)
			continue
		}
	}

	close(errors)
	for err := range errors {
		t.Errorf("concurrent write error: %v", err)
	}

	// Verify final state
	finalReg := NewRegistry()
	if err := finalReg.Load(filePath); err != nil {
		t.Fatalf("failed to load final state: %v", err)
	}

	if len(finalReg.List()) != numWrites {
		t.Errorf("expected %d tasks in final state, got %d", numWrites, len(finalReg.List()))
	}
}

func TestRegistryVersionConflict(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "tasks.json")

	// Create initial registry
	reg1 := NewRegistry()
	task1 := New("ua-001", "First")
	reg1.Add(task1)
	if err := reg1.Save(filePath); err != nil {
		t.Fatalf("failed to save reg1: %v", err)
	}

	// Load into two separate registries
	reg2 := NewRegistry()
	if err := reg2.Load(filePath); err != nil {
		t.Fatalf("failed to load reg2: %v", err)
	}

	reg3 := NewRegistry()
	if err := reg3.Load(filePath); err != nil {
		t.Fatalf("failed to load reg3: %v", err)
	}

	// reg2 adds a task and saves successfully
	task2 := New("ua-002", "Second")
	reg2.Add(task2)
	if err := reg2.Save(filePath); err != nil {
		t.Fatalf("reg2 save should succeed: %v", err)
	}

	// reg3 tries to save - should fail with version conflict
	task3 := New("ua-003", "Third")
	reg3.Add(task3)
	err := reg3.Save(filePath)
	if err == nil {
		t.Error("expected version conflict error for reg3 save")
	}
	if err != nil && !strings.Contains(err.Error(), "version conflict") {
		t.Errorf("expected version conflict error, got: %v", err)
	}
}
