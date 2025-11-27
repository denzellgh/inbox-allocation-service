package testutil

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
)

// ==================== MockConversationRepository ====================

type MockConversationRepository struct {
	mu            sync.RWMutex
	conversations map[uuid.UUID]*domain.ConversationRef

	// For controlling behavior in tests
	GetByIDError      error
	CreateError       error
	UpdateError       error
	AllocateNextError error
}

func NewMockConversationRepository() *MockConversationRepository {
	return &MockConversationRepository{
		conversations: make(map[uuid.UUID]*domain.ConversationRef),
	}
}

func (m *MockConversationRepository) Create(ctx context.Context, conv *domain.ConversationRef) error {
	if m.CreateError != nil {
		return m.CreateError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conversations[conv.ID] = conv
	return nil
}

func (m *MockConversationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ConversationRef, error) {
	if m.GetByIDError != nil {
		return nil, m.GetByIDError
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	conv, ok := m.conversations[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return conv, nil
}

func (m *MockConversationRepository) Update(ctx context.Context, conv *domain.ConversationRef) error {
	if m.UpdateError != nil {
		return m.UpdateError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conversations[conv.ID] = conv
	return nil
}

func (m *MockConversationRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.ConversationRef, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.ConversationRef
	for _, conv := range m.conversations {
		if conv.TenantID == tenantID {
			result = append(result, conv)
		}
	}
	return result, nil
}

func (m *MockConversationRepository) GetByInboxID(ctx context.Context, inboxID uuid.UUID, state *domain.ConversationState) ([]*domain.ConversationRef, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.ConversationRef
	for _, conv := range m.conversations {
		if conv.InboxID == inboxID {
			if state == nil || conv.State == *state {
				result = append(result, conv)
			}
		}
	}
	return result, nil
}

func (m *MockConversationRepository) GetByOperatorID(ctx context.Context, operatorID uuid.UUID) ([]*domain.ConversationRef, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.ConversationRef
	for _, conv := range m.conversations {
		if conv.AssignedOperatorID != nil && *conv.AssignedOperatorID == operatorID {
			result = append(result, conv)
		}
	}
	return result, nil
}

func (m *MockConversationRepository) GetQueuedForOperator(ctx context.Context, operatorID uuid.UUID, inboxIDs []uuid.UUID, limit int) ([]*domain.ConversationRef, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.ConversationRef
	for _, conv := range m.conversations {
		if conv.State == domain.ConversationStateQueued {
			for _, inboxID := range inboxIDs {
				if conv.InboxID == inboxID {
					result = append(result, conv)
					if len(result) >= limit {
						return result, nil
					}
				}
			}
		}
	}
	return result, nil
}

func (m *MockConversationRepository) AllocateNext(ctx context.Context, tenantID, operatorID uuid.UUID, inboxIDs []uuid.UUID) (*domain.ConversationRef, error) {
	if m.AllocateNextError != nil {
		return nil, m.AllocateNextError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, conv := range m.conversations {
		if conv.TenantID == tenantID && conv.State == domain.ConversationStateQueued {
			for _, inboxID := range inboxIDs {
				if conv.InboxID == inboxID {
					conv.State = domain.ConversationStateAllocated
					conv.AssignedOperatorID = &operatorID
					return conv, nil
				}
			}
		}
	}
	return nil, nil // No conversation available
}

// AddConversation adds a conversation to the mock (for test setup)
func (m *MockConversationRepository) AddConversation(conv *domain.ConversationRef) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conversations[conv.ID] = conv
}

// ==================== MockOperatorStatusRepository ====================

type MockOperatorStatusRepository struct {
	mu       sync.RWMutex
	statuses map[uuid.UUID]*domain.OperatorStatus
}

func NewMockOperatorStatusRepository() *MockOperatorStatusRepository {
	return &MockOperatorStatusRepository{
		statuses: make(map[uuid.UUID]*domain.OperatorStatus),
	}
}

func (m *MockOperatorStatusRepository) Create(ctx context.Context, status *domain.OperatorStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statuses[status.OperatorID] = status
	return nil
}

func (m *MockOperatorStatusRepository) GetByOperatorID(ctx context.Context, operatorID uuid.UUID) (*domain.OperatorStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	status, ok := m.statuses[operatorID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return status, nil
}

func (m *MockOperatorStatusRepository) Update(ctx context.Context, status *domain.OperatorStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statuses[status.OperatorID] = status
	return nil
}

func (m *MockOperatorStatusRepository) Upsert(ctx context.Context, status *domain.OperatorStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statuses[status.OperatorID] = status
	return nil
}

func (m *MockOperatorStatusRepository) Delete(ctx context.Context, operatorID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.statuses, operatorID)
	return nil
}

// AddStatus adds a status to the mock (for test setup)
func (m *MockOperatorStatusRepository) AddStatus(status *domain.OperatorStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statuses[status.OperatorID] = status
}

// ==================== MockSubscriptionRepository ====================

type MockSubscriptionRepository struct {
	mu            sync.RWMutex
	subscriptions map[uuid.UUID]*domain.OperatorInboxSubscription
}

func NewMockSubscriptionRepository() *MockSubscriptionRepository {
	return &MockSubscriptionRepository{
		subscriptions: make(map[uuid.UUID]*domain.OperatorInboxSubscription),
	}
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, sub *domain.OperatorInboxSubscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscriptions[sub.ID] = sub
	return nil
}

func (m *MockSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.OperatorInboxSubscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sub, ok := m.subscriptions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return sub, nil
}

func (m *MockSubscriptionRepository) GetByOperatorID(ctx context.Context, operatorID uuid.UUID) ([]*domain.OperatorInboxSubscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.OperatorInboxSubscription
	for _, sub := range m.subscriptions {
		if sub.OperatorID == operatorID {
			result = append(result, sub)
		}
	}
	return result, nil
}

func (m *MockSubscriptionRepository) GetByInboxID(ctx context.Context, inboxID uuid.UUID) ([]*domain.OperatorInboxSubscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.OperatorInboxSubscription
	for _, sub := range m.subscriptions {
		if sub.InboxID == inboxID {
			result = append(result, sub)
		}
	}
	return result, nil
}

func (m *MockSubscriptionRepository) GetByOperatorAndInbox(ctx context.Context, operatorID, inboxID uuid.UUID) (*domain.OperatorInboxSubscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, sub := range m.subscriptions {
		if sub.OperatorID == operatorID && sub.InboxID == inboxID {
			return sub, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.subscriptions, id)
	return nil
}

func (m *MockSubscriptionRepository) DeleteByOperatorAndInbox(ctx context.Context, operatorID, inboxID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, sub := range m.subscriptions {
		if sub.OperatorID == operatorID && sub.InboxID == inboxID {
			delete(m.subscriptions, id)
			return nil
		}
	}
	return nil
}

// AddSubscription adds a subscription to the mock (for test setup)
func (m *MockSubscriptionRepository) AddSubscription(sub *domain.OperatorInboxSubscription) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscriptions[sub.ID] = sub
}

// ==================== MockIdempotencyRepository ====================

type MockIdempotencyRepository struct {
	mu   sync.RWMutex
	keys map[string]*domain.IdempotencyKey // key: tenant_id:key
}

func NewMockIdempotencyRepository() *MockIdempotencyRepository {
	return &MockIdempotencyRepository{
		keys: make(map[string]*domain.IdempotencyKey),
	}
}

func (m *MockIdempotencyRepository) Create(ctx context.Context, ik *domain.IdempotencyKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := ik.TenantID.String() + ":" + ik.Key
	m.keys[key] = ik
	return nil
}

func (m *MockIdempotencyRepository) GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (*domain.IdempotencyKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	k := tenantID.String() + ":" + key
	ik, ok := m.keys[k]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return ik, nil
}

func (m *MockIdempotencyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, ik := range m.keys {
		if ik.ID == id {
			delete(m.keys, k)
			return nil
		}
	}
	return nil
}

func (m *MockIdempotencyRepository) DeleteExpired(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	for k, ik := range m.keys {
		if ik.IsExpired() {
			delete(m.keys, k)
			count++
		}
	}
	return count, nil
}

func (m *MockIdempotencyRepository) GetExpiredForCleanup(ctx context.Context, limit int) ([]*domain.IdempotencyKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.IdempotencyKey
	for _, ik := range m.keys {
		if ik.IsExpired() {
			result = append(result, ik)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}
