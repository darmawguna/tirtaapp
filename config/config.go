package config

import (
	"log"

	"github.com/spf13/viper"
)

// LoadConfig memuat konfigurasi dari file .env di root proyek.
func LoadConfig() {
	// viper.AddConfigPath("../../") // Path ke direktori root dari main.go
	viper.AddConfigPath(".") 
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv() // Otomatis membaca environment variable sistem

	if err := viper.ReadInConfig(); err != nil {
		// Gunakan 'Is' untuk error yang lebih spesifik jika perlu
		log.Fatal("Error reading config file", err)
	}
}