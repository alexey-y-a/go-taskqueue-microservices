package logger

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	log = nil

	Init()

	require.NotNil(t, log, "логгер должен быть инициализирован")

	_, ok := log.Formatter.(*logrus.JSONFormatter)
	require.True(t, ok, "должен быть JSONFormatter")

	require.Equal(t, logrus.InfoLevel, log.Level, "уровень должен быть Info")

}
func TestL(t *testing.T) {
	log = nil

	logger := L()

	require.NotNil(t, logger, "должен вернуть не-nil логгер")
}
func TestWithComponent(t *testing.T) {
	entry := WithComponent("test-service")

	require.NotNil(t, entry, "WithComponent должен вернуть не-nil entry")

	data, err := entry.String()
	require.NoError(t, err, "не должно быть ошибок при получении данных entry")
	require.Contains(t, data, "test-service", "должен содержать имя компонента test-service")
}

func TestWithRequest(t *testing.T) {
	entry := WithRequest("api-gateway", "GET", "/health", "req-123")
	require.NotNil(t, entry, "WithRequest должен вернуть не-nil entry")

	require.Equal(t, "api-gateway", entry.Data["component"])
	require.Equal(t, "GET", entry.Data["http_method"])
	require.Equal(t, "/health", entry.Data["http_path"])
	require.Equal(t, "req-123", entry.Data["request_id"])
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestWithError(t *testing.T) {
	testErr := "database connection failed"
	err := &testError{msg: testErr}

	entry := WithError("queue-service", err)

	require.Equal(t, "queue-service", entry.Data["component"])
	require.Equal(t, testErr, entry.Data["error"])
}

func TestWithFields(t *testing.T) {
	fields := logrus.Fields{
		"user_id": 123,
		"task_id": "task-456",
		"retry":   3,
	}

	entry := WithFields("worker-service", fields)

	require.Equal(t, "worker-service", entry.Data["component"])
	require.Equal(t, 123, entry.Data["user_id"])
	require.Equal(t, "task-456", entry.Data["task_id"])
	require.Equal(t, 3, entry.Data["retry"])
}

func BenchmarkLogging(b *testing.B) {
	Init()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithComponent("benchmark").WithField("iteration", i).Info("benchmark log")
	}
}
