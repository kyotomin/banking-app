package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/nyaruka/phonenumbers/v2"
)

// Todo 1
func validatePhoneNumber(phone string) (string, error) {
	number, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return "", fmt.Errorf("unable to parse phone number")
	}

	if !phonenumbers.IsValidNumber(number) {
		return "", fmt.Errorf("entered number is invalid")
	}

	return phonenumbers.Format(number, phonenumbers.E164), nil
}

func GenOTP() string {
	max := big.NewInt(10000)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return ""
	}

	return fmt.Sprint("%04d", n)
}

func HashOTP(secret, challengeID, code string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(challengeID + ":" + code))
	return mac.Sum(nil)
}
