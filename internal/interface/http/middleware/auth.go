package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	secret string
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{secret: secret}
}

func (m *AuthMiddleware) RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		claims, ok := m.parse(token)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if claims.Role != role {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		r.Header.Set("X-User-ID", claims.Sub)
		next.ServeHTTP(w, r)
	})
}

type Claims struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
}

func (m *AuthMiddleware) parse(token string) (Claims, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return Claims{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, false
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, false
	}
	h := hmac.New(sha256.New, []byte(m.secret))
	h.Write(payload)
	if !hmac.Equal(sig, h.Sum(nil)) {
		return Claims{}, false
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, false
	}
	if claims.Sub == "" || claims.Role == "" {
		return Claims{}, false
	}
	return claims, true
}
