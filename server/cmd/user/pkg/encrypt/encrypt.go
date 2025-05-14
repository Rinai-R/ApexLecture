package encrypt

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func ComparePasswords(hashedPwd string, plainPw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPw))
	if err != nil {
		return false
	}
	return true
}

func EncryptPassword(password string) string {
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return string(encryptedPassword)
}
