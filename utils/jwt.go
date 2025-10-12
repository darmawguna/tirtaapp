package utils

import (
	"os"
	"strconv"
	"time"

	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

func GenerateJWT(user models.User) (string, error) {
	// Ambil durasi expired dari .env
	expHours, err := strconv.Atoi(os.Getenv("JWT_EXPIRATION_HOURS"))
	if err != nil {
		expHours = 24 // Default 24 jam jika .env tidak ada
	}

	// Buat claims (data yang disimpan di dalam token)
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * time.Duration(expHours)).Unix(),
		"iat":     time.Now().Unix(),
	}

	// Buat token baru dengan claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Tandatangani token dengan secret key
	secretKey := viper.GetString("JWT_SECRET_KEY")
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
