// Package history — in-memory хранилище диалогов с LLM по user_id:
// лимит сообщений на диалог, TTL неактивности, ленивая чистка просроченных.
package history

import (
	"sync"
	"time"

	"alisa-gpt/internal/models"
)

// dialog — история одного пользователя: сообщения и время последнего доступа
// (для TTL). Доступен только через методы Store, которые держат мьютекс.
type dialog struct {
	messages []models.Message
	lastSeen time.Time
}

// Store — потокобезопасное хранилище диалогов. Snapshot возвращает копию,
// поэтому мутации среза снаружи не влияют на хранимую историю.
type Store struct {
	mu      sync.Mutex
	dialogs map[string]*dialog

	limit int
	ttl   time.Duration

	nowFn func() time.Time
}

// NewStore создаёт хранилище: limit — максимум сообщений в диалоге (свежие
// хвостовые), ttl — сколько диалог живёт с последнего доступа.
func NewStore(limit int, ttl time.Duration) *Store {
	return &Store{
		dialogs: make(map[string]*dialog),
		limit:   limit,
		ttl:     ttl,
		nowFn:   time.Now,
	}
}

// setNow подменяет источник времени (для детерминированных тестов).
func (s *Store) setNow(fn func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nowFn = fn
}

// Snapshot возвращает копию истории диалога пользователя. Для неизвестного
// пользователя — пустой срез. Обновляет время последнего доступа.
func (s *Store) Snapshot(userID string) []models.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanup()

	d, ok := s.dialogs[userID]
	if !ok {
		return []models.Message{}
	}

	d.lastSeen = s.nowFn()

	return cloneMessages(d.messages)
}

// Append добавляет сообщение в диалог пользователя, обновляет время доступа
// и обрезает историю до лимита (сохраняются свежие хвостовые сообщения).
func (s *Store) Append(userID string, msg models.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanup()

	d, ok := s.dialogs[userID]
	if !ok {
		d = &dialog{messages: make([]models.Message, 0, 2)}
		s.dialogs[userID] = d
	}

	d.lastSeen = s.nowFn()
	d.messages = append(d.messages, msg)

	if excess := len(d.messages) - s.limit; excess > 0 {
		d.messages = d.messages[excess:]
	}
}

// Reset удаляет диалог пользователя.
func (s *Store) Reset(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.dialogs, userID)
}

// Len возвращает размер хранилища (диалогов) — для логов и тестов.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.dialogs)
}

// cleanup удаляет диалоги, к которым не обращались дольше ttl.
// Вызывается только под мьютексом из точек доступа.
func (s *Store) cleanup() {
	now := s.nowFn()

	for id, d := range s.dialogs {
		if now.Sub(d.lastSeen) > s.ttl {
			delete(s.dialogs, id)
		}
	}
}

func cloneMessages(src []models.Message) []models.Message {
	dst := make([]models.Message, len(src))
	copy(dst, src)

	return dst
}
