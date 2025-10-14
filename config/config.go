package config

import (
	"log"

	"github.com/spf13/viper"
)

func LoadConfig() {
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// Ini akan membuat Viper membaca variabel dari environment (seperti yang disuplai Docker)
	viper.AutomaticEnv()

	// [PERBAIKAN FINAL] Coba baca file .env, tapi JANGAN crash jika tidak ada.
	// Viper akan lanjut menggunakan environment variables dari AutomaticEnv() secara otomatis.
	err := viper.ReadInConfig()
	if err != nil {
		// Cetak sebagai peringatan saja, bukan error fatal.
		log.Println("Warning: Could not find or read .env file, relying on environment variables. Error:", err)
	}
}