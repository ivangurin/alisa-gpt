package history

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"alisa-gpt/internal/models"
)

// newTestStore — хранилище с управляемыми часами (детерминированный TTL).
func newTestStore(limit int, ttl time.Duration) (*Store, *time.Time) {
	now := time.Unix(0, 0)
	s := NewStore(limit, ttl)
	s.setNow(func() time.Time { return now })

	return s, &now
}

func userMsg(content string) models.Message {
	return models.Message{Role: models.RoleUser, Content: content}
}

func TestAppendTrimsToLimit(t *testing.T) {
	s, _ := newTestStore(2, time.Hour)

	s.Append("u1", userMsg("один"))
	s.Append("u1", userMsg("два"))
	s.Append("u1", userMsg("три"))

	got := s.Snapshot("u1")
	require.Len(t, got, 2)
	require.Equal(t, "два", got[0].Content)
	require.Equal(t, "три", got[1].Content)
}

func TestTTLCleansExpiredDialogs(t *testing.T) {
	s, now := newTestStore(10, time.Minute)

	s.Append("u1", userMsg("вопрос"))
	s.Append("u2", userMsg("вопрос"))

	*now = now.Add(2 * time.Minute)
	s.Append("u3", userMsg("вопрос"))

	require.Empty(t, s.Snapshot("u1"))
	require.Empty(t, s.Snapshot("u2"))
	require.Len(t, s.Snapshot("u3"), 1)
}

func TestResetDropsDialog(t *testing.T) {
	s, _ := newTestStore(10, time.Hour)

	s.Append("u1", userMsg("вопрос"))
	s.Reset("u1")

	require.Empty(t, s.Snapshot("u1"))
	require.Equal(t, 0, s.Len())
}

func TestSnapshotUnknownUserEmpty(t *testing.T) {
	s, _ := newTestStore(10, time.Hour)

	require.Empty(t, s.Snapshot("ghost"))
}

func TestSnapshotReturnsCopy(t *testing.T) {
	s, _ := newTestStore(10, time.Hour)

	s.Append("u1", userMsg("вопрос"))

	got := s.Snapshot("u1")
	got[0].Content = "испорчено"

	require.Equal(t, "вопрос", s.Snapshot("u1")[0].Content)
}
