package account

import "github.com/nyaruka/phonenumbers/v2"

// Todo 1
func validatePhoneNumber(phone string) (string, bool) {
	number, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return "", false
	}

	if !phonenumbers.IsValidNumber(number) {
		return "", false
	}

	return phonenumbers.Format(number, phonenumbers.E164), true
}
