package config

import (
	"log"

	"github.com/spf13/viper"
)

func LoadConfig() {
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// [PERBAIKAN] Kode ini memberitahu Viper untuk juga membaca dari environment variables.
	// Ini penting agar variabel dari docker-compose bisa terbaca.
	viper.AutomaticEnv()

	// [PERBAIKAN] Ubah cara kita membaca file.
	// Kita coba baca file, tapi jika tidak ada, kita tidak akan crash.
	if err := viper.ReadInConfig(); err != nil {
		// Cek jika errornya adalah karena file tidak ditemukan
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// File konfigurasi tidak ditemukan; ini tidak apa-apa,
			// kita akan mengandalkan environment variables.
			log.Println("Config file (.env) not found, relying on environment variables.")
		} else {
			// Jika file ditemukan tapi ada error lain saat membacanya
			log.Fatal("Error reading config file:", err)
		}
	}
}