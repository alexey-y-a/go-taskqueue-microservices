package queue

import (
	"testing"
)

func TestCreateAndGetTask(t *testing.T) {
    store := NewStore()
    task := store.CreateTask("email", "hello")

    got, ok :=  store.GetTask(task.ID)
    if !ok {
        t.Fatalf("task not found by ID")
    }
    if got.Type != "email" || got.Payload != "hello" {
        t.Fatalf("unexpected task data: %+v", got)
    }
}