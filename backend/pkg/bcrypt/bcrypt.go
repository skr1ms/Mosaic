package bcrypt

import (
	"github.com/skr1ms/mosaic/pkg/middleware"
	"golang.org/x/crypto/bcrypt"
)

var logger = middleware.NewLogger()

// HashPassword creates bcrypt hash from plain text password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.GetZerologLogger().Error().Err(err).Msg("Password hashing failed")
		return "", err
	}

	logger.GetZerologLogger().Info().Msg("Password hashed successfully")
	return string(bytes), nil
}

// CheckPassword verifies password against bcrypt hash
func CheckPassword(password, hash string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	success := err == nil

	if success {
		logger.GetZerologLogger().Info().Msg("Password verification successful")
	} else {
		logger.GetZerologLogger().Error().Err(err).Msg("Password verification failed")
	}

	return success
}
