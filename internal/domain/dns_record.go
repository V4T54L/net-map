package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type RecordType string

const (
	CNAME RecordType = "CNAME"
	A     RecordType = "A"
)

var (
	ErrInvalidDomainName  = errors.New("invalid domain name format")
	ErrInvalidRecordType  = errors.New("invalid record type, must be CNAME or A")
	ErrInvalidRecordValue = errors.New("invalid record value for the given type")
)

// domainNameRegex is a simple regex for domain name validation.
// It allows for subdomains and a TLD.
var domainNameRegex = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`)

// ipv4Regex checks for a valid IPv4 address.
var ipv4Regex = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)

type DNSRecord struct {
	ID         int64
	UserID     int64
	DomainName string
	Type       RecordType
	Value      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewDNSRecord(userID int64, domainName, value string, recordType RecordType) (*DNSRecord, error) {
	domainName = strings.ToLower(strings.TrimSpace(domainName))
	value = strings.TrimSpace(value)

	if !domainNameRegex.MatchString(domainName) {
		return nil, ErrInvalidDomainName
	}

	switch recordType {
	case CNAME:
		if !domainNameRegex.MatchString(value) {
			return nil, ErrInvalidRecordValue
		}
	case A:
		if !ipv4Regex.MatchString(value) {
			return nil, ErrInvalidRecordValue
		}
	default:
		return nil, ErrInvalidRecordType
	}

	return &DNSRecord{
		UserID:     userID,
		DomainName: domainName,
		Type:       recordType,
		Value:      value,
	}, nil
}

