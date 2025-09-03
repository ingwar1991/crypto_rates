package email_helper

import (
    "fmt"
    "strings"
    "net/smtp"

	"crypto_rates_auth/internal/config"
)

func SendOTP(cfg *config.Config, email, code string) error {
    if cfg.SMTP.Host == "" || cfg.SMTP.Port == 0 || cfg.SMTP.User == "" || cfg.SMTP.Pass == "" || cfg.SMTP.FromEmail == "" {
        return fmt.Errorf("smtp not configured")
    }

    auth := smtp.PlainAuth("", cfg.SMTP.User, cfg.SMTP.Pass, cfg.SMTP.Host)
    addr := fmt.Sprintf("%s:%d", cfg.SMTP.Host, cfg.SMTP.Port)
    msg := strings.Join([]string{
        fmt.Sprintf("From: %s", cfg.SMTP.FromEmail),
        fmt.Sprintf("To: %s", email),
        "Subject: Your login code",
        "MIME-Version: 1.0",
        "Content-Type: text/plain; charset=utf-8",
        "",
        fmt.Sprintf("Your one-time code is: %s\nIt expires in 5 minutes.", code),
    }, "\r\n")

    return smtp.SendMail(addr, auth, cfg.SMTP.FromEmail, []string{email}, []byte(msg))
}
