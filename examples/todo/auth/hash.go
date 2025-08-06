package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(pw string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(hashed), err
}

func ComparePassword(hashed, cmp string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(cmp))
	return err == nil
}
