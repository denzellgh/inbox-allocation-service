package middleware

import (
	"context"
	"net/http"

	"github.com/inbox-allocation-service/internal/api/response"
	"github.com/inbox-allocation-service/internal/domain"
	"github.com/inbox-allocation-service/internal/repository"
)

const OperatorRoleKey ContextKey = "operator_role"

// OperatorLoader loads operator role into context
func OperatorLoader(repos *repository.RepositoryContainer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			operatorID, ok := GetOperatorUUID(ctx)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			operator, err := repos.Operators.GetByID(ctx, operatorID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx = context.WithValue(ctx, OperatorRoleKey, operator.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetOperatorRole(ctx context.Context) (domain.OperatorRole, bool) {
	role, ok := ctx.Value(OperatorRoleKey).(domain.OperatorRole)
	return role, ok
}

// RequireAdmin ensures only ADMIN can access
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetOperatorRole(r.Context())
		if !ok || role != domain.OperatorRoleAdmin {
			response.Forbidden(w, "Admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireManager ensures MANAGER or ADMIN can access
func RequireManager(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetOperatorRole(r.Context())
		if !ok || (role != domain.OperatorRoleManager && role != domain.OperatorRoleAdmin) {
			response.Forbidden(w, "Manager or Admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
