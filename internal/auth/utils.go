package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/nyaruka/phonenumbers/v2"
	"golang.org/x/crypto/argon2"
)

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

func GenerateRegistrationToken() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate registration token: %w", err)
	}

	return hex.EncodeToString(b), nil
}

var (
	ownerRegex = regexp.MustCompile(
		`^[A-Za-z]+(?:[-'][A-Za-z]+)? [A-Za-z]+(?:[-'][A-Za-z]+)?$`,
	)
	pinRegex = regexp.MustCompile(`^[0-9]{4}$`)
)

func ValidateRegistrationInput(
	owner, pin, currency string,
) error {
	owner = strings.Join(strings.Fields(owner), " ")

	if !ownerRegex.MatchString(owner) {
		return errors.New("owner must contain two Latin names")
	}

	if !pinRegex.MatchString(pin) {
		return errors.New("PIN must contain exactly 4 digits")
	}

	if currency != "USD" && currency != "RUB" {
		return errors.New("unsupported currency")
	}

	return nil
}

func HashPIN(pin string) ([]byte, []byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, err
	}

	hash := argon2.IDKey(
		[]byte(pin),
		salt,
		3,
		64*1024,
		4,
		32,
	)

	return hash, salt, nil
}
