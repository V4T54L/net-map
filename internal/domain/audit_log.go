package domain

import (
	"encoding/json"
	"time"
)

type ActionType string

const (
	ActionCreateDNSRecord  ActionType = "CREATE_DNS_RECORD"
	ActionUpdateDNSRecord  ActionType = "UPDATE_DNS_RECORD"
	ActionDeleteDNSRecord  ActionType = "DELETE_DNS_RECORD"
	ActionUserRegister     ActionType = "USER_REGISTER"
	ActionUserLoginSuccess ActionType = "USER_LOGIN_SUCCESS"
	ActionUserLoginFailure ActionType = "USER_LOGIN_FAILURE"
	ActionUpdateUserStatus ActionType = "UPDATE_USER_STATUS"
)

type AuditLog struct {
	ID        int64
	UserID    int64 // Can be 0 for system actions or failed logins
	Action    ActionType
	TargetID  int64 // e.g., DNSRecord ID or User ID being acted upon
	OldValue  json.RawMessage
	NewValue  json.RawMessage
	Timestamp time.Time
}

func NewAuditLog(userID int64, action ActionType, targetID int64, oldValue, newValue interface{}) (*AuditLog, error) {
	var oldJSON, newJSON json.RawMessage
	var err error

	if oldValue != nil {
		oldJSON, err = json.Marshal(oldValue)
		if err != nil {
			return nil, err
		}
	}

	if newValue != nil {
		newJSON, err = json.Marshal(newValue)
		if err != nil {
			return nil, err
		}
	}

	return &AuditLog{
		UserID:   userID,
		Action:   action,
		TargetID: targetID,
		OldValue: oldJSON,
		NewValue: newJSON,
	}, nil
}