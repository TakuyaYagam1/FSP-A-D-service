package password

import "golang.org/x/crypto/bcrypt"

const bcryptCost = 12

type Hasher struct{}

func New() *Hasher { return &Hasher{} }

func (*Hasher) Hash(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (*Hasher) Verify(hash, value string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(value))
}
