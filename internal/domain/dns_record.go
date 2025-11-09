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
	ErrInvalidRecordType  = errors.New("invalid record type") // Updated message
	ErrInvalidRecordValue = errors.New("invalid record value for the given type")
)

// domainNameRegex validates domain names, ensuring labels don't start or end with a hyphen.
var domainNameRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$`)

// ipv4Regex validates IPv4 addresses.
var ipv4Regex = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)

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
	case A: // Reordered cases
		if !ipv4Regex.MatchString(value) {
			return nil, ErrInvalidRecordValue
		}
	case CNAME:
		if !domainNameRegex.MatchString(value) {
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

