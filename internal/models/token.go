package models

import "os"

type GenerateTokenRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var SecretKey = func() []byte {
	key := os.Getenv("JWT_SECRET_KEY")
	if key == "" {
		return []byte("secretKey")
	}
	return []byte(key)
}()
