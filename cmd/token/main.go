package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
)

type claims struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
}

func main() {
	secret := flag.String("secret", "dev-secret", "JWT secret")
	sub := flag.String("sub", "user-1", "subject/user id")
	role := flag.String("role", "user", "role: user/admin")
	flag.Parse()

	payload, _ := json.Marshal(claims{Sub: *sub, Role: *role})
	h := hmac.New(sha256.New, []byte(*secret))
	h.Write(payload)
	sig := h.Sum(nil)
	token := base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(sig)
	fmt.Println(token)
}
