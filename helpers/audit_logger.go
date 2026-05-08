package helpers

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type AuditEventType string

const (
	AuditLogin           AuditEventType = "LOGIN"
	AuditLoginFailed     AuditEventType = "LOGIN_FAILED"
	AuditLogout          AuditEventType = "LOGOUT"
	AuditTransaction     AuditEventType = "TRANSACTION"
	AuditDataAccess      AuditEventType = "DATA_ACCESS"
	AuditPasswordChange  AuditEventType = "PASSWORD_CHANGE"
	AuditAccountLocked   AuditEventType = "ACCOUNT_LOCKED"
	AuditSuspiciousInput AuditEventType = "SUSPICIOUS_INPUT"
	AuditRateLimited     AuditEventType = "RATE_LIMITED"
)

type AuditLog struct {
	Timestamp string         `json:"timestamp"`
	Event     AuditEventType `json:"event"`
	UserID    string         `json:"user_id,omitempty"`
	UserType  string         `json:"user_type,omitempty"`
	IP        string         `json:"ip"`
	UserAgent string         `json:"user_agent,omitempty"`
	Resource  string         `json:"resource,omitempty"`
	Details   string         `json:"details,omitempty"`
	Success   bool           `json:"success"`
}

var (
	auditFile   *os.File
	auditMu     sync.Mutex
	auditInited bool
)

func initAuditLog() {
	if auditInited {
		return
	}
	logPath := os.Getenv("AUDIT_LOG_PATH")
	if logPath == "" {
		logPath = "logs/audit.log"
	}
	_ = os.MkdirAll("logs", 0750)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return
	}
	auditFile = f
	auditInited = true
}

// WriteAuditLog writes a structured audit event to the audit log file.
// Never logs passwords, tokens, or other secrets.
func WriteAuditLog(event AuditLog) {
	auditMu.Lock()
	defer auditMu.Unlock()

	initAuditLog()
	if auditFile == nil {
		return
	}

	event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	auditFile.Write(append(data, '\n'))
}
