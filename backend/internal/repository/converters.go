package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/inbox-allocation-service/internal/domain"
)

// ==================== Error Mapping ====================

func mapError(err error) error {
	if err == nil {
		return nil
	}

	// Handle PostgreSQL specific errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "55P03": // lock_not_available (NOWAIT failed)
			return domain.ErrConversationLocked
		case "40P01": // deadlock_detected
			return domain.ErrLockTimeout
		case "23505": // unique_violation
			return domain.ErrAlreadyExists
		}
	}

	// Handle pgx errors
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}

	return err
}

// ==================== UUID Converters ====================

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidPtrToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func pgtypeToUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return id.Bytes
}

func pgtypeToUUIDPtr(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	uid := uuid.UUID(id.Bytes)
	return &uid
}

// ==================== Time Converters ====================

func timeToPgtype(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func timePtrToPgtype(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func pgtypeToTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func pgtypeToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// ==================== Decimal Converters ====================

func decimalToPgtype(d decimal.Decimal) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(d.String())
	return num
}

func pgtypeToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}

	// Convert pgtype.Numeric to string and then to decimal.Decimal
	str := n.Int.String()
	d, err := decimal.NewFromString(str)
	if err != nil {
		return decimal.Zero
	}

	// Apply exponent
	return d.Shift(n.Exp)
}

// ==================== String Converters ====================

func stringPtrToPgtype(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func pgtypeToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// ==================== Domain Value Object Converters ====================

func conversationStateToPgtype(s domain.ConversationState) ConversationState {
	return ConversationState(s)
}

func pgtypeToConversationState(s ConversationState) domain.ConversationState {
	return domain.ConversationState(s)
}

func operatorRoleToPgtype(r domain.OperatorRole) OperatorRole {
	return OperatorRole(r)
}

func pgtypeToOperatorRole(r OperatorRole) domain.OperatorRole {
	return domain.OperatorRole(r)
}

func operatorStatusTypeToPgtype(s domain.OperatorStatusType) OperatorStatusType {
	return OperatorStatusType(s)
}

func pgtypeToOperatorStatusType(s OperatorStatusType) domain.OperatorStatusType {
	return domain.OperatorStatusType(s)
}

func gracePeriodReasonToPgtype(r domain.GracePeriodReason) GracePeriodReason {
	return GracePeriodReason(r)
}

func pgtypeToGracePeriodReason(r GracePeriodReason) domain.GracePeriodReason {
	return domain.GracePeriodReason(r)
}
