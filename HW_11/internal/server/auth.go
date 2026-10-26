package server

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type LoginRequest struct {
	Login    string `json:"login" example:"admin"`
	Password string `json:"password" example:"поменяй_меня"`
}

type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

// login godoc
//
//	@Summary		Авторизация
//	@Description	Проверяет логин и пароль из .env и выдаёт JWT на 1 час
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		LoginRequest	true	"логин и пароль"
//	@Success		200		{object}	LoginResponse
//	@Failure		400		{string}	string	"неправильный json"
//	@Failure		401		{string}	string	"неверный логин или пароль"
//	@Router			/login [post]
func login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "неправильный json", http.StatusBadRequest)
		return
	}
	if req.Login != os.Getenv("LOGIN") || req.Password != os.Getenv("PASSWORD") {
		http.Error(w, "неверный логин или пароль", http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"sub": req.Login,
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "не получилось создать токен", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: tokenString})
}

// проверка токена
func authRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "нужна авторизация", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(header, "Bearer ")

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
		if err != nil || !token.Valid {
			http.Error(w, "неправильный токен", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
