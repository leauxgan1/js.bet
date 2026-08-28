package internal

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	// "strings"
)

const userIDKey string = "userId"

type UserClaims struct {
	UserID   int64  `json:"user"`
	Password string `json:"pass"`
	jwt.RegisteredClaims
}

// Passes on userID
func authMiddlewarePermissive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr string
		cookie, err := r.Cookie("jwt_token")
		if err == nil && cookie.Value != "" {
			tokenStr = cookie.Value
			log.Printf("jwt cookie found: %s\n", cookie)
		}

		if tokenStr == "" {
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenStr = authHeader[7:]
			}
		}

		if tokenStr == "" {
			log.Printf("Could not find auth header 'Authorization'")
			next.ServeHTTP(w, r)
			return

		}

		token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (any, error) {
			return []byte(SECRET), nil
		})

		if err != nil || !token.Valid {
			next.ServeHTTP(w, r) // Pass along to next handler
			// http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*UserClaims)
		if !ok {
			next.ServeHTTP(w, r) // Pass along to next handler
			// http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Passes userid and fails when not authorized
func authMiddlewareStrict(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Starting auth")
		var tokenStr string
		cookie, err := r.Cookie("jwt_token")
		if err != nil {
			log.Printf("Error retrieving cookie: %v", err)
		} else {
			if cookie.Value != "" {
				tokenStr = cookie.Value
				log.Printf("jwt cookie found: %s\n", cookie.Value)
			}
		}

		if tokenStr == "" {
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenStr = authHeader[7:]
			}
		}

		if tokenStr == "" {
			log.Printf("Could not find auth header 'Authorization'")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (any, error) {
			return []byte(SECRET), nil
		})

		if err != nil || !token.Valid {
			log.Printf("Error or invalid token when parsing token string")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*UserClaims)
		if !ok {
			log.Printf("Invalid claims")
			http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
