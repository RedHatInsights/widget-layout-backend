package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/RedHatInsights/widget-layout-backend/pkg/config"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/sirupsen/logrus"
)

func InjectUserIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ih := r.Header.Get("x-rh-identity")
		i, err := identity.DecodeAndCheckIdentity(ih)
		if err != nil {
			logrus.Errorf("Failed to decode identity: %v", err)
			http.Error(w, "Invalid identity header", http.StatusBadRequest)
			return
		}
		ctx = context.WithValue(ctx, config.IdentityContextKey, i)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIdentity(ctx context.Context) identity.XRHID {
	id, ok := ctx.Value(config.IdentityContextKey).(identity.XRHID)
	if !ok {
		logrus.Error("Identity not found in context")
		panic(errors.New("identity not found in context"))
	}
	return id
}
