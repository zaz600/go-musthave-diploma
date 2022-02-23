package hasher

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	// TODO соль нужна и замена на scrypt
	// https://www.sohamkamani.com/golang/password-authentication-and-storage/
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	return string(bytes), err
}

func CheckPasswordHash(password string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
