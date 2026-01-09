package services

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/darmawguna/tirtaapp.git/repositories"
	"github.com/darmawguna/tirtaapp.git/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	models "github.com/darmawguna/tirtaapp.git/model"
)

type PasswordResetService interface {
	ForgotPassword(email string) error
	ResetPassword(token string, newPassword string) error
}

type passwordResetService struct {
	userRepo  repositories.UserRepository
	resetRepo repositories.PasswordResetRepository
	mailer    EmailSender
}

func NewPasswordResetService(
	userRepo repositories.UserRepository,
	resetRepo repositories.PasswordResetRepository,
	mailer EmailSender,
) PasswordResetService {
	return &passwordResetService{
		userRepo:  userRepo,
		resetRepo: resetRepo,
		mailer:    mailer,
	}
}

func (s *passwordResetService) ForgotPassword(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		// Anti-enumeration: email tidak ada = tetap sukses
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	// Optional: hapus token aktif sebelumnya untuk user ini (biar rapi dan mudah demo)
	_ = s.resetRepo.DeleteActiveByUser(user.ID)

	token, err := utils.GenerateResetToken(32) // 32 bytes => 64 hex char
	if err != nil {
		return err
	}
	tokenHash := utils.HashTokenSHA256Hex(token)

	expMin := 15
	if v := os.Getenv("PASSWORD_RESET_EXP_MINUTES"); v != "" {
		if n, convErr := strconv.Atoi(v); convErr == nil && n > 0 {
			expMin = n
		}
	}

	now := time.Now()
	reset := models.PasswordReset{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(time.Duration(expMin) * time.Minute),
	}

	if _, err := s.resetRepo.Create(reset); err != nil {
		return err
	}

	deepLinkBase := os.Getenv("APP_DEEPLINK_BASE")
	if deepLinkBase == "" {
		deepLinkBase = "https://tirtapp.fmews.com/reset-password?token="
	}
	resetLink := deepLinkBase + token

	subject := "Reset Password - Tirta App"
	body := fmt.Sprintf(`
	<p>Yth. Pengguna Tirta App,</p>

	<p>Kami menerima permintaan untuk melakukan <b>reset password</b> pada akun Anda.</p>

	<p>
		Silakan klik tautan berikut untuk membuat password baru. 
		Tautan ini bersifat <b>rahasia</b> dan hanya berlaku selama <b>%d menit</b>.
	</p>

	<p>
		<a href="%s" style="color:#1a73e8; font-weight:bold;">
			Reset Password
		</a>
	</p>

	<p>
		Apabila tautan di atas tidak terbuka secara otomatis di perangkat Anda, 
		Anda dapat menyalin dan menggunakan <b>token reset</b> berikut secara manual:
	</p>

	<p style="padding:10px; background-color:#f5f5f5; font-family:monospace; word-break:break-all;">
		%s
	</p>

	<p>
		Jika Anda <b>tidak pernah</b> meminta reset password, 
		abaikan email ini. Akun Anda tetap aman dan tidak akan mengalami perubahan.
	</p>

	<p>Hormat kami,<br>
	<b>Tirta App</b></p>
`, expMin, resetLink, resetLink)

	return s.mailer.Send(user.Email, subject, body)
}

func (s *passwordResetService) ResetPassword(token string, newPassword string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("invalid token")
	}

	now := time.Now()
	tokenHash := utils.HashTokenSHA256Hex(token)

	pr, err := s.resetRepo.FindValidByTokenHash(tokenHash, now)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("token invalid or expired")
		}
		return err
	}

	user, err := s.userRepo.FindByID(pr.UserID)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// ini update penuh; untuk skripsi acceptable, tapi jangan overwrite field lain dari struct kosong.
	// Pastikan user yang di-load benar dari DB (kita sudah FindByID).
	user.Password = string(hashedPassword)

	if _, err := s.userRepo.Update(user); err != nil {
		return err
	}

	if err := s.resetRepo.MarkUsed(pr.ID, now); err != nil {
		return err
	}

	return nil
}
