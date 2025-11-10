package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDNSRecord(t *testing.T) {
	testCases := []struct {
		name          string
		userID        int64
		domainName    string
		value         string
		recordType    RecordType
		expectError   error // Changed to error type
		expectedName  string
		expectedValue string
	}{
		{
			name:          "Valid CNAME Record",
			userID:        1,
			domainName:    "service.example.com",
			value:         "target.example.com",
			recordType:    CNAME,
			expectError:   nil,
			expectedName:  "service.example.com",
			expectedValue: "target.example.com",
		},
		{
			name:          "Valid A Record",
			userID:        1,
			domainName:    "host.internal.net",
			value:         "192.168.1.100",
			recordType:    A,
			expectError:   nil,
			expectedName:  "host.internal.net",
			expectedValue: "192.168.1.100",
		},
		{
			name:          "Valid A Record with trimming and lowercasing",
			userID:        1,
			domainName:    "  HoSt.Internal.Net  ",
			value:         "  192.168.1.100  ",
			recordType:    A,
			expectError:   nil,
			expectedName:  "host.internal.net",
			expectedValue: "192.168.1.100",
		},
		{
			name:        "Invalid Domain Name - leading hyphen",
			userID:      1,
			domainName:  "-invalid.com",
			value:       "1.2.3.4",
			recordType:  A,
			expectError: ErrInvalidDomainName,
		},
		{
			name:        "Invalid Domain Name - trailing hyphen",
			userID:      1,
			domainName:  "invalid-.com",
			value:       "1.2.3.4",
			recordType:  A,
			expectError: ErrInvalidDomainName,
		},
		{
			name:        "Invalid Domain Name - no TLD",
			userID:      1,
			domainName:  "invalid",
			value:       "1.2.3.4",
			recordType:  A,
			expectError: ErrInvalidDomainName,
		},
		{
			name:        "Invalid Record Type",
			userID:      1,
			domainName:  "test.com",
			value:       "1.2.3.4",
			recordType:  "MX",
			expectError: ErrInvalidRecordType,
		},
		{
			name:        "Invalid A Record Value - not an IP",
			userID:      1,
			domainName:  "test.com",
			value:       "not-an-ip",
			recordType:  A,
			expectError: ErrInvalidRecordValue,
		},
		{
			name:        "Invalid A Record Value - out of range",
			userID:      1,
			domainName:  "test.com",
			value:       "256.0.0.1",
			recordType:  A,
			expectError: ErrInvalidRecordValue,
		},
		{
			name:        "Invalid CNAME Record Value - not a domain",
			userID:      1,
			domainName:  "test.com",
			value:       "192.168.1.1",
			recordType:  CNAME,
			expectError: ErrInvalidRecordValue,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			record, err := NewDNSRecord(tc.userID, tc.domainName, tc.value, tc.recordType)

			if tc.expectError != nil {
				assert.ErrorIs(t, err, tc.expectError)
				assert.Nil(t, record)
			} else {
				require.NoError(t, err)
				require.NotNil(t, record)
				assert.Equal(t, tc.userID, record.UserID)
				assert.Equal(t, tc.expectedName, record.DomainName)
				assert.Equal(t, tc.expectedValue, record.Value)
				assert.Equal(t, tc.recordType, record.Type)
			}
		})
	}
}
