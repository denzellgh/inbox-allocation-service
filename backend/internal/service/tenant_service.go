package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/pkg/logger"
	"github.com/inbox-allocation-service/internal/repository"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type TenantService struct {
	repos  *repository.RepositoryContainer
	logger *logger.Logger
}

func NewTenantService(repos *repository.RepositoryContainer, log *logger.Logger) *TenantService {
	return &TenantService{repos: repos, logger: log}
}

func (s *TenantService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	return s.repos.Tenants.GetByID(ctx, id)
}

func (s *TenantService) UpdateWeights(ctx context.Context, tenantID uuid.UUID, alpha, beta decimal.Decimal, updatedBy *uuid.UUID) (*domain.Tenant, error) {
	tenant, err := s.repos.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	tenant.PriorityWeightAlpha = alpha
	tenant.PriorityWeightBeta = beta
	tenant.UpdatedAt = time.Now().UTC()
	tenant.UpdatedBy = updatedBy

	if err := s.repos.Tenants.Update(ctx, tenant); err != nil {
		return nil, err
	}

	s.logger.Info("Tenant weights updated",
		zap.String("tenant_id", tenantID.String()),
		zap.String("alpha", alpha.String()),
		zap.String("beta", beta.String()),
	)

	return tenant, nil
}
