package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var secretKey = []byte("secret -key")

type authKey string

const (
	AuthUserID authKey = "middleware.auth.userID"

	Issuer = "obsidian-rt"
)

type CustomClaims struct {
	jwt.RegisteredClaims
}

func writeUnauthed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)

	if _, err := w.Write([]byte(http.StatusText(http.StatusUnauthorized))); err != nil {
		http.Error(w, "error sending response", http.StatusInternalServerError)
		return
	}
}

func IsAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		if !strings.HasPrefix(authorization, "Bearer ") {
			writeUnauthed(w)
			return
		}

		encodedToken := strings.TrimPrefix(authorization, "Bearer ")
		userID, err := VerifyToken(encodedToken)
		if err != nil {
			writeUnauthed(w)
			return
		}

		ctx := context.WithValue(r.Context(), AuthUserID, userID)
		req := r.WithContext(ctx)

		next.ServeHTTP(w, req)
	})
}

func CreateToken(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		CustomClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				Issuer:    Issuer,
				Subject:   strconv.Itoa(userID),
				ID:        uuid.New().String(),
			},
		})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", nil
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithLeeway(5*time.Second),
		jwt.WithIssuer(Issuer),
	)
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims := token.Claims.(*CustomClaims)
	sub, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, fmt.Errorf("invalid sub")
	}

	return sub, nil
}

func UserIDFromCtx(ctx context.Context) int {
	return ctx.Value(AuthUserID).(int)
}
