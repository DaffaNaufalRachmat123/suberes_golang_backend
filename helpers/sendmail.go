package helpers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"sync" // Import sync untuk menangani Singleton

	"encoding/json"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/hibiken/asynq"
)

type EmailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

const TypeEmailDelivery = "email:send"

// --- Global Variable untuk Token Caching ---
var (
	globalTokenSource oauth2.TokenSource
	tokenOnce         sync.Once
	asynqClient       *asynq.Client
	asynqServer       *asynq.Server
)

// Config menampung konfigurasi OAuth2
type Config struct {
	UserEmail    string
	ClientID     string
	ClientSecret string
	RefreshToken string
}

func enqueueEmail(from, to, subject, body string) {

	payload, err := json.Marshal(EmailPayload{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
	})
	if err != nil {
		log.Println("Failed marshal email:", err)
		return
	}

	task := asynq.NewTask(TypeEmailDelivery, payload)

	_, err = asynqClient.Enqueue(
		task,
		asynq.MaxRetry(5),
		asynq.Timeout(30*time.Second),
		asynq.Queue("critical"),
	)

	if err != nil {
		log.Println("Failed enqueue email:", err)
	}
}

func handleEmailTask(ctx context.Context, t *asynq.Task) error {

	var p EmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	log.Printf("📨 Sending email to %s", p.To)

	err := sendMail(p.From, p.To, p.Subject, p.Body)
	if err != nil {
		log.Printf("❌ Failed email to %s: %v", p.To, err)
		return err // otomatis retry
	}

	log.Printf("✅ Email sent to %s", p.To)
	return nil
}

func InitEmailQueue(redisAddr string) {

	asynqClient = asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
	})

	asynqServer = asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeEmailDelivery, handleEmailTask)

	go func() {
		if err := asynqServer.Run(mux); err != nil {
			log.Fatal(err)
		}
	}()
}

// LoadConfig mengambil env variables
func LoadConfig() Config {
	return Config{
		UserEmail:    os.Getenv("USER_EMAIL"),
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RefreshToken: os.Getenv("REFRESH_TOKEN"),
	}
}

// getGlobalTokenSource: Singleton untuk inisialisasi token source
// Ini kunci optimasinya: Hanya dijalankan 1x seumur hidup aplikasi
func getGlobalTokenSource() oauth2.TokenSource {
	tokenOnce.Do(func() {
		cfg := LoadConfig()

		conf := &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     google.Endpoint,
			Scopes:       []string{"https://mail.google.com/"},
		}

		initialToken := &oauth2.Token{
			RefreshToken: cfg.RefreshToken,
		}

		// TokenSource ini pintar, dia menyimpan token di memori
		// dan hanya me-refresh ke Google jika token sudah expired.
		globalTokenSource = conf.TokenSource(context.Background(), initialToken)
	})

	return globalTokenSource
}

// sendMail mengirim email menggunakan token yang sudah di-cache
func sendMail(from, to, subject, htmlBody string) error {
	// 1. Ambil TokenSource global (Instan, tidak bikin baru)
	tokenSource := getGlobalTokenSource()

	// 2. Ambil token (Cepat, kecuali expired baru dia request HTTP)
	token, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("gagal mendapatkan token: %v", err)
	}

	// Setup Header Email (MIME)
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	// Gmail SMTP Server
	addr := "smtp.gmail.com:587"

	// Implementasi Auth XOAUTH2 Custom
	// Kita perlu LoadConfig lagi cuma buat ambil UserEmail, ini ringan (baca env var)
	auth := xoauth2Auth{
		username:    LoadConfig().UserEmail,
		accessToken: token.AccessToken,
	}

	err = smtp.SendMail(addr, auth, from, []string{to}, msg.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// xoauth2Auth implementasi smtp.Auth untuk Gmail
type xoauth2Auth struct {
	username, accessToken string
}

func (a xoauth2Auth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "XOAUTH2", []byte(fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.username, a.accessToken)), nil
}

func (a xoauth2Auth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return []byte{}, nil
	}
	return nil, nil
}

// --- Helper Template Parser ---

func parseTemplate(tplName, tplString string, data interface{}) (string, error) {
	t, err := template.New(tplName).Parse(tplString)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// --- Fungsi-Fungsi Utama (Sesuai JS) ---

// Struct data untuk template
type EmailData struct {
	LogoSuberes          string
	SupportEmail         string
	SupportOfficeAddress string
	// Dynamic Fields
	Text               string
	TimeInvitedString  string
	PlaceInvitedString string
	Email              string
	Password           string
	CompleteName       string
	UserType           string
	UserTypeCapital    string
	Reason             string
}

// NewEmailData helper untuk mengisi data env default
func NewEmailData() EmailData {
	return EmailData{
		LogoSuberes:          os.Getenv("LOGO_SUBERES"),
		SupportEmail:         os.Getenv("SUPPORT_EMAIL"),
		SupportOfficeAddress: os.Getenv("SUPPORT_OFFICE_ADDRESS"),
	}
}

func SendInvitedMailMitra(from, to, subject, name, timeInvited, placeInvited string) {
	data := NewEmailData()
	data.Text = name
	data.TimeInvitedString = timeInvited
	data.PlaceInvitedString = placeInvited

	htmlBody, err := parseTemplate("invited", htmlInvited, data)
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	enqueueEmail(from, to, subject, htmlBody)
}

func SendAcceptedMailMitra(from, to, subject, email, password string) {
	data := NewEmailData()
	data.Email = email
	data.Password = password

	htmlBody, err := parseTemplate("acceptedMitra", htmlAcceptedMitra, data)
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	enqueueEmail(from, to, subject, htmlBody)
}

func SendAcceptedAdminAccount(from, to, subject, email, password string) {
	data := NewEmailData()
	data.Email = email
	data.Password = password

	htmlBody, err := parseTemplate("acceptedAdmin", htmlAcceptedAdmin, data)
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	enqueueEmail(from, to, subject, htmlBody)
}

func SendActiveAdminAccount(from, to, subject, completeName, userType, userTypeCapital, email, password, reason string) {
	data := NewEmailData()
	data.CompleteName = completeName
	data.UserType = userType
	data.UserTypeCapital = userTypeCapital
	data.Email = email
	data.Password = password
	data.Reason = reason

	htmlBody, err := parseTemplate("activeAdmin", htmlActiveAdmin, data)
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	enqueueEmail(from, to, subject, htmlBody)
}

func SendNonactiveAdminAccount(from, to, subject, completeName, userType, userTypeCapital, email, reason string) {
	data := NewEmailData()
	data.CompleteName = completeName
	data.UserType = userType
	data.UserTypeCapital = userTypeCapital
	data.Email = email
	data.Reason = reason

	htmlBody, err := parseTemplate("nonactiveAdmin", htmlNonactiveAdmin, data)
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	enqueueEmail(from, to, subject, htmlBody)
}

func SendRemoveAdminAccount(from, to, subject, completeName, userTypeCapital, email, reason string) {
	data := NewEmailData()
	data.CompleteName = completeName
	data.UserTypeCapital = userTypeCapital
	data.Email = email
	data.Reason = reason

	htmlBody, err := parseTemplate("removeAdmin", htmlRemoveAdmin, data)
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	enqueueEmail(from, to, subject, htmlBody)
}

func SendOtpCodeMail(from, to, subject, otpCode string) {
	data := NewEmailData()
	data.Text = otpCode

	htmlBody, err := parseTemplate("otp", htmlOtp, data)
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	enqueueEmail(from, to, subject, htmlBody)
}

func SendMitraStatus(from, to, subject, mitraStatus, email, text string) {
	data := NewEmailData()
	data.Email = email
	data.Text = text

	var selectedTemplate string
	if mitraStatus == "suspend" {
		selectedTemplate = htmlMitraSuspend
	} else {
		selectedTemplate = htmlMitraActive
	}

	htmlBody, err := parseTemplate("mitraStatus", selectedTemplate, data)
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	enqueueEmail(from, to, subject, htmlBody)
}

func main() {
	// Contoh penggunaan
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Uncomment baris di bawah untuk mengetes
	// SendOtpCodeMail("admin@suberes.com", "target@example.com", "Kode OTP", "123456")
}

// ==========================================
// KUMPULAN HTML TEMPLATES
// ==========================================

const htmlInvited = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD XHTML 1.0 Transitional //EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title></title>
</head>
<body class="clean-body u_body" style="margin: 0;padding: 0;background-color: #ffffff;color: #000000">
    <img align="center" border="0" src="{{.LogoSuberes}}" alt="Suberes Icon" width="158.4"/>
    
    <div style="text-align: center;">
        <p><strong>Proses Verifikasi Data Diri Dan Kelengkapan</strong></p>
    </div>

    <div style="text-align: center;">
        <p>Kepada Yth {{.Text}}, dimohon untuk datang ke kantor Suberes untuk melakukan proses verifikasi data diri dan kelengkapan sebagai syarat untuk persetujuan bergabung sebagai Mitra Suberes pada tanggal berikut :</p>
    </div>

    <div style="text-align: center;">
        <p style="color: #2dc26b;"><strong>{{.TimeInvitedString}}</strong></p>
        <p><strong>{{.PlaceInvitedString}}</strong></p>
        <p><strong>Jakarta Timur</strong></p>
    </div>

    <div style="text-align: center;">
        <p>If you have any questions, reply to this email or <strong>contact us at </strong>{{.SupportEmail}}</p>
        <p>{{.SupportOfficeAddress}}</p>
    </div>
</body>
</html>`

const htmlAcceptedMitra = `<!DOCTYPE HTML>
<html>
<head></head>
<body>
    <img src="{{.LogoSuberes}}" width="158.4"/>
    <p><strong>Kamu Telah Disetujui Menjadi Mitra Suberes</strong></p>
    <p>Selamat, kamu telah disetujui menjadi mitra Suberes.</p>
    
    <p>Akun mu</p>
    <p><strong>Email : <a href="mailto:{{.Email}}">{{.Email}}</a></strong></p>
    <p><strong>Password : {{.Password}}</strong></p>
    
    <p>Harap segera mengganti password Mitra dalam aplikasi.</p>

    <p>Contact: {{.SupportEmail}} | {{.SupportOfficeAddress}}</p>
</body>
</html>`

const htmlAcceptedAdmin = `<!DOCTYPE HTML>
<html>
<body>
    <img src="{{.LogoSuberes}}" width="158.4"/>
    <p><strong>Akun Admin Kamu Telah Dibuat</strong></p>
    <p>Selamat, akun admin kamu telah dibuat dan disetujui oleh Suberes.</p>
    
    <p>Akun mu</p>
    <p><strong>Email : <a href="mailto:{{.Email}}">{{.Email}}</a></strong></p>
    <p><strong>Password : {{.Password}}</strong></p>

    <p>Contact: {{.SupportEmail}} | {{.SupportOfficeAddress}}</p>
</body>
</html>`

const htmlActiveAdmin = `<!DOCTYPE HTML>
<html>
<body>
    <img src="{{.LogoSuberes}}" width="158.4"/>
    <p><strong>Akun {{.UserTypeCapital}} Kamu Telah Diaktifkan Kembali</strong></p>
    
    <p>Yth {{.CompleteName}} dengan email {{.Email}}.<br />
    Pihak Suberes dengan ini telah mengaktifkan kembali akun mu sebagai {{.UserType}} dengan pertimbangan</p>
    
    <p style="font-style:italic">"{{.Reason}}"</p>
    
    <p>Akun mu</p>
    <p><strong>Email : <a href="mailto:{{.Email}}">{{.Email}}</a></strong></p>
    <p><strong>Password : {{.Password}}</strong></p>

    <p>Contact: {{.SupportEmail}} | {{.SupportOfficeAddress}}</p>
</body>
</html>`

const htmlNonactiveAdmin = `<!DOCTYPE HTML>
<html>
<body>
    <img src="{{.LogoSuberes}}" width="158.4"/>
    <p><strong>Akun {{.UserTypeCapital}} Kamu Dinonaktifkan</strong></p>
    
    <p>Yth {{.CompleteName}} dengan email {{.Email}}.<br />
    Berdasarkan keputusan dari pihak Suberes, akun {{.UserType}} dinonaktifkan karena melanggar peraturan kerja dengan pertimbangan:</p>
    
    <p style="font-style:italic">"{{.Reason}}"</p>

    <p>Contact: {{.SupportEmail}} | {{.SupportOfficeAddress}}</p>
</body>
</html>`

const htmlRemoveAdmin = `<!DOCTYPE HTML>
<html>
<body>
    <img src="{{.LogoSuberes}}" width="158.4"/>
    <p><strong>Akun {{.UserTypeCapital}} Telah Dihapus</strong></p>
    
    <p>Yth {{.CompleteName}} dengan email {{.Email}}.<br />
    Berdasarkan keputusan dari pihak Suberes, akun {{.UserTypeCapital}} telah dihapus dengan pertimbangan:</p>
    
    <p style="font-style:italic">"{{.Reason}}"</p>
    
    <p>Terima kasih atas semua kerja keras yang telah kamu berikan.</p>

    <p>Contact: {{.SupportEmail}} | {{.SupportOfficeAddress}}</p>
</body>
</html>`

const htmlOtp = `<!DOCTYPE HTML>
<html>
<body>
    <img src="{{.LogoSuberes}}" width="158.4"/>
    <p>Ini adalah kode OTP kamu, harap jangan berikan kepada siapapun termasuk pihak Suberes</p>
    
    <p style="font-size: 18px; color: #2dc26b;"><strong>{{.Text}}</strong></p>
    <p><strong>Jakarta Timur</strong></p>

    <p>Contact: {{.SupportEmail}} | {{.SupportOfficeAddress}}</p>
</body>
</html>`

const htmlMitraSuspend = `<!DOCTYPE HTML>
<html>
<body>
    <img src="{{.LogoSuberes}}" width="158.4"/>
    <p>Akun mu dengan email : {{.Email}} telah dinonaktifkan karena melanggar Syarat dan Ketentuan Suberes</p>
    
    <p style="font-size: 18px; color: #e53935;"><strong>{{.Text}}</strong></p>
    <p><strong>Jakarta Timur</strong></p>

    <p>Contact: {{.SupportEmail}} | {{.SupportOfficeAddress}}</p>
</body>
</html>`

const htmlMitraActive = `<!DOCTYPE HTML>
<html>
<body>
    <img src="{{.LogoSuberes}}" width="158.4"/>
    <p>Akun mu dengan email : {{.Email}} telah diaktifkan kembali, silahkan kembali masuk ke aplikasi untuk bekerja</p>
    
    <p style="font-size: 18px; color: #2dc26b;"><strong>{{.Text}}</strong></p>
    <p><strong>Jakarta Timur</strong></p>

    <p>Contact: {{.SupportEmail}} | {{.SupportOfficeAddress}}</p>
</body>
</html>`
