// Package mailer envoie des courriels via le serveur SMTP interne (SGMail).
package mailer

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net/smtp"
	"strings"
	"time"
)

// Config SMTP. Host/Port pointent vers le container mailserver (réseau Docker interne).
type Config struct {
	Host string // ex: mailserver
	Port string // ex: 587
	User string // compte d'authentification, ex: contact@mcserver.vbt-prog.com
	Pass string
	From string // expéditeur
	To   string // destinataire des messages de support
}

// Enabled indique si l'envoi est configuré (mot de passe présent).
func (c Config) Enabled() bool {
	return c.Host != "" && c.User != "" && c.Pass != "" && c.To != ""
}

// SendSupport envoie un message du formulaire de support.
// replyTo = email du visiteur (pour répondre directement).
func (c Config) SendSupport(name, replyTo, subject, message string) error {
	body := fmt.Sprintf("Nom: %s\r\nEmail: %s\r\n\r\n%s\r\n", name, replyTo, message)
	enc := mime.QEncoding.Encode("utf-8", "[Support Playrena] "+subject)

	var sb strings.Builder
	fmt.Fprintf(&sb, "From: Playrena Support <%s>\r\n", c.From)
	fmt.Fprintf(&sb, "To: %s\r\n", c.To)
	if replyTo != "" {
		fmt.Fprintf(&sb, "Reply-To: %s\r\n", replyTo)
	}
	fmt.Fprintf(&sb, "Subject: %s\r\n", enc)
	fmt.Fprintf(&sb, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)

	return c.send([]byte(sb.String()))
}

func (c Config) send(msg []byte) error {
	addr := c.Host + ":" + c.Port
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer client.Close()

	if err := client.Hello("mcserver.vbt-prog.com"); err != nil {
		return fmt.Errorf("smtp hello: %w", err)
	}

	// STARTTLS — le mailserver présente le cert Origin (*.vbt-prog.com) ; on est
	// sur un réseau Docker interne de confiance, donc on n'exige pas la chaîne
	// publique (InsecureSkipVerify) tout en chiffrant bien la connexion.
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: c.Host, InsecureSkipVerify: true}); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	auth := smtp.PlainAuth("", c.User, c.Pass, c.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := client.Mail(c.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(c.To); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	return client.Quit()
}
