package store

import (
	"testing"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_Create(t *testing.T) {
	store := NewMemoryStore(10)

	task := taskmodel.NewTask(uuid.New().String(), "email", "test payload")

	err := store.Create(task)
	require.NoError(t, err, "Создание задачи не должно возвращать ошибку")

	saved, err := store.Get(task.ID)
	require.NoError(t, err)
	require.Equal(t, task.ID, saved.ID)
	require.Equal(t, task.Type, saved.Type)
}

func TestMemoryStore_Create_Duplicate(t *testing.T) {
	store := NewMemoryStore(10)
	task := taskmodel.NewTask("task-1", "email", "test")

	err := store.Create(task)
	require.NoError(t, err)

	err = store.Create(task)
	require.Error(t, err, ErrTaskExists, "Должна быть ошибка ErrTaskExists")
}

func TestMemoryStore_Get_NotFound(t *testing.T) {
	store := NewMemoryStore(10)

	_, err := store.Get("non-existent")

	require.ErrorIs(t, err, ErrTaskNotFound, "Должна быть ошибка ErrTaskNotFound")
}

func TestMemoryStore_Update(t *testing.T) {
	store := NewMemoryStore(10)

	task := taskmodel.NewTask("task-1", "email", "original")
	err := store.Create(task)
	require.NoError(t, err)

	task.UpdateStatus(taskmodel.StatusCompleted, "")

	err = store.Update(task)
	require.NoError(t, err)

	updated, err := store.Get(task.ID)
	require.NoError(t, err)
	require.Equal(t, taskmodel.StatusCompleted, updated.Status)
}
func TestMemoryStore_List(t *testing.T) {
	store := NewMemoryStore(10)

	for i := 0; i < 5; i++ {
		task := taskmodel.NewTask(uuid.New().String(), "email", "test payload")
		err := store.Create(task)
		require.NoError(t, err)
	}

	tasks, err := store.List(10, 0)
	require.NoError(t, err)
	require.Len(t, tasks, 5, "Должно быть 5 задач")

	tasks, err = store.List(2, 0)
	require.NoError(t, err)
	require.Len(t, tasks, 2, "Лимит 2 должен вернуть 2 задачи")
}

func TestMemoryStore_GetPending(t *testing.T) {
	store := NewMemoryStore(10)

	task1 := taskmodel.NewTask("task-1", "email", "pending")
	err := store.Create(task1)
	require.NoError(t, err)

	task2 := taskmodel.NewTask("task-2", "email", "completed")
	task2.UpdateStatus(taskmodel.StatusCompleted, "")
	err = store.Create(task2)
	require.NoError(t, err)

	task3 := taskmodel.NewTask("task-3", "email", "pending")
	err = store.Create(task3)
	require.NoError(t, err)

	pending, err := store.GetPending(10)
	require.NoError(t, err)
	require.Len(t, pending, 2, "Должно быть 2 задачи со статусом pending")

	for _, task := range pending {
		require.Equal(t, taskmodel.StatusPending, task.Status)
	}
}
