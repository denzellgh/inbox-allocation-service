package repository

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RepositoryContainer holds all repository instances
type RepositoryContainer struct {
	pool                   *pgxpool.Pool
	queries                *Queries
	Tenants                *TenantRepositoryImpl
	Inboxes                *InboxRepositoryImpl
	Operators              *OperatorRepositoryImpl
	Subscriptions          *SubscriptionRepositoryImpl
	OperatorStatus         *OperatorStatusRepositoryImpl
	ConversationRefs       *ConversationRefRepositoryImpl
	Labels                 *LabelRepositoryImpl
	ConversationLabels     *ConversationLabelRepositoryImpl
	GracePeriodAssignments *GracePeriodRepositoryImpl
}

// NewRepositoryContainer creates all repository instances
func NewRepositoryContainer(pool *pgxpool.Pool) *RepositoryContainer {
	queries := New(pool)

	return &RepositoryContainer{
		pool:                   pool,
		queries:                queries,
		Tenants:                NewTenantRepository(queries),
		Inboxes:                NewInboxRepository(queries),
		Operators:              NewOperatorRepository(queries),
		Subscriptions:          NewSubscriptionRepository(queries),
		OperatorStatus:         NewOperatorStatusRepository(queries),
		ConversationRefs:       NewConversationRefRepository(queries, pool),
		Labels:                 NewLabelRepository(queries),
		ConversationLabels:     NewConversationLabelRepository(queries),
		GracePeriodAssignments: NewGracePeriodRepository(queries, pool),
	}
}

// WithTx returns queries bound to a transaction
func (rc *RepositoryContainer) WithTx(tx pgx.Tx) *Queries {
	return rc.queries.WithTx(tx)
}
