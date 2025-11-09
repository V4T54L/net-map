package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDNSRecord(t *testing.T) {
	testCases := []struct {
		name        string
		userID      int64
		domainName  string
		value       string
		recordType  RecordType
		expectError bool
		errorType   error
	}{
		{
			name:        "Valid CNAME Record",
			userID:      1,
			domainName:  "service.internal.local",
			value:       "other-service.internal.local",
			recordType:  CNAME,
			expectError: false,
		},
		{
			name:        "Valid A Record",
			userID:      1,
			domainName:  "db.internal.local",
			value:       "192.168.1.10",
			recordType:  A,
			expectError: false,
		},
		{
			name:        "Invalid Domain Name",
			userID:      1,
			domainName:  "invalid-domain",
			value:       "192.168.1.10",
			recordType:  A,
			expectError: true,
			errorType:   ErrInvalidDomainName,
		},
		{
			name:        "Invalid Record Type",
			userID:      1,
			domainName:  "service.internal.local",
			value:       "192.168.1.10",
			recordType:  "INVALID",
			expectError: true,
			errorType:   ErrInvalidRecordType,
		},
		{
			name:        "Invalid CNAME Value (IP Address)",
			userID:      1,
			domainName:  "service.internal.local",
			value:       "192.168.1.10",
			recordType:  CNAME,
			expectError: true,
			errorType:   ErrInvalidRecordValue,
		},
		{
			name:        "Invalid A Value (Domain Name)",
			userID:      1,
			domainName:  "service.internal.local",
			value:       "other-service.internal.local",
			recordType:  A,
			expectError: true,
			errorType:   ErrInvalidRecordValue,
		},
		{
			name:        "Domain Name with leading/trailing spaces",
			userID:      1,
			domainName:  "  spaced.domain.com  ",
			value:       "1.2.3.4",
			recordType:  A,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			record, err := NewDNSRecord(tc.userID, tc.domainName, tc.value, tc.recordType)

			if tc.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.errorType)
				assert.Nil(t, record)
			} else {
				require.NoError(t, err)
				require.NotNil(t, record)
				assert.Equal(t, tc.userID, record.UserID)
				assert.Equal(t, strings.TrimSpace(tc.domainName), record.DomainName)
				assert.Equal(t, tc.value, record.Value)
				assert.Equal(t, tc.recordType, record.Type)
			}
		})
	}
}

