package dns

import (
	"context"
	"errors"
	"internal-dns/internal/domain"
	"internal-dns/internal/infrastructure/cache"
	"internal-dns/internal/repository"
	"net"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockDNSRecordUseCase struct {
	mock.Mock
}

func (m *MockDNSRecordUseCase) ResolveDomain(ctx context.Context, domainName string) (*domain.DNSRecord, error) {
	args := m.Called(ctx, domainName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSRecord), args.Error(1)
}
func (m *MockDNSRecordUseCase) CreateRecord(context.Context, int64, string, string, domain.RecordType) (*domain.DNSRecord, error) {
	return nil, errors.New("not implemented")
}
func (m *MockDNSRecordUseCase) GetRecordByID(context.Context, int64, int64) (*domain.DNSRecord, error) {
	return nil, errors.New("not implemented")
}
func (m *MockDNSRecordUseCase) ListRecordsByUser(context.Context, int64, int, int) ([]*domain.DNSRecord, int, error) {
	return nil, 0, errors.New("not implemented")
}
func (m *MockDNSRecordUseCase) UpdateRecord(context.Context, int64, int64, string, string, domain.RecordType) (*domain.DNSRecord, error) {
	return nil, errors.New("not implemented")
}
func (m *MockDNSRecordUseCase) DeleteRecord(context.Context, int64, int64) error {
	return errors.New("not implemented")
}

type MockDNSRecordCache struct {
	mock.Mock
}

func (m *MockDNSRecordCache) Get(ctx context.Context, domainName string) (*domain.DNSRecord, error) {
	args := m.Called(ctx, domainName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSRecord), args.Error(1)
}
func (m *MockDNSRecordCache) Set(ctx context.Context, record *domain.DNSRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}
func (m *MockDNSRecordCache) Delete(ctx context.Context, domainName string) error {
	args := m.Called(ctx, domainName)
	return args.Error(0)
}

func TestServer_handleRequest(t *testing.T) {
	aRecord := &domain.DNSRecord{DomainName: "test-a.local.", Type: domain.A, Value: "1.2.3.4"}
	cnameRecord := &domain.DNSRecord{DomainName: "test-cname.local.", Type: domain.CNAME, Value: "target.local"}

	t.Run("Cache Hit A Record", func(t *testing.T) {
		mockUC := new(MockDNSRecordUseCase)
		mockCache := new(MockDNSRecordCache)
		server := NewServer(":53535", mockUC, mockCache)

		mockCache.On("Get", mock.Anything, "test-a.local.").Return(aRecord, nil).Once()

		req := new(dns.Msg)
		req.SetQuestion("test-a.local.", dns.TypeA)
		w := &mockResponseWriter{}

		server.handleRequest(w, req)

		require.NotNil(t, w.msg)
		assert.Equal(t, dns.RcodeSuccess, w.msg.Rcode)
		require.Len(t, w.msg.Answer, 1)
		rr := w.msg.Answer[0].(*dns.A)
		assert.Equal(t, "1.2.3.4", rr.A.String())

		mockCache.AssertExpectations(t)
		mockUC.AssertNotCalled(t, "ResolveDomain", mock.Anything, mock.Anything)
	})

	t.Run("Cache Miss DB Hit CNAME Record", func(t *testing.T) {
		mockUC := new(MockDNSRecordUseCase)
		mockCache := new(MockDNSRecordCache)
		server := NewServer(":53535", mockUC, mockCache)

		mockCache.On("Get", mock.Anything, "test-cname.local.").Return(nil, cache.ErrCacheMiss).Once()
		mockUC.On("ResolveDomain", mock.Anything, "test-cname.local.").Return(cnameRecord, nil).Once()
		mockCache.On("Set", mock.Anything, cnameRecord).Return(nil).Once()

		req := new(dns.Msg)
		req.SetQuestion("test-cname.local.", dns.TypeCNAME)
		w := &mockResponseWriter{}

		server.handleRequest(w, req)

		require.NotNil(t, w.msg)
		assert.Equal(t, dns.RcodeSuccess, w.msg.Rcode)
		require.Len(t, w.msg.Answer, 1)
		rr := w.msg.Answer[0].(*dns.CNAME)
		assert.Equal(t, "target.local.", rr.Target)

		mockCache.AssertExpectations(t)
		mockUC.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockUC := new(MockDNSRecordUseCase)
		mockCache := new(MockDNSRecordCache)
		server := NewServer(":53535", mockUC, mockCache)

		mockCache.On("Get", mock.Anything, "not-found.local.").Return(nil, cache.ErrCacheMiss).Once()
		mockUC.On("ResolveDomain", mock.Anything, "not-found.local.").Return(nil, repository.ErrDNSRecordNotFound).Once()

		req := new(dns.Msg)
		req.SetQuestion("not-found.local.", dns.TypeA)
		w := &mockResponseWriter{}

		server.handleRequest(w, req)

		require.NotNil(t, w.msg)
		assert.Equal(t, dns.RcodeNameError, w.msg.Rcode) // NXDOMAIN
		assert.Empty(t, w.msg.Answer)

		mockCache.AssertExpectations(t)
		mockUC.AssertExpectations(t)
	})
}

func BenchmarkServer_handleRequest(b *testing.B) {
	mockUC := new(MockDNSRecordUseCase)
	mockCache := new(MockDNSRecordCache)
	server := NewServer(":53535", mockUC, mockCache)

	aRecord := &domain.DNSRecord{DomainName: "bench.local.", Type: domain.A, Value: "5.6.7.8"}

	mockCache.On("Get", mock.Anything, "bench.local.").Return(nil, cache.ErrCacheMiss)
	mockUC.On("ResolveDomain", mock.Anything, "bench.local.").Return(aRecord, nil)
	mockCache.On("Set", mock.Anything, aRecord).Return(nil)

	req := new(dns.Msg)
	req.SetQuestion("bench.local.", dns.TypeA)
	w := &mockResponseWriter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.handleRequest(w, req)
	}
}

type mockResponseWriter struct {
	msg *dns.Msg
}

func (m *mockResponseWriter) LocalAddr() net.Addr         { return &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 53} }
func (m *mockResponseWriter) RemoteAddr() net.Addr        { return &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345} }
func (m *mockResponseWriter) WriteMsg(msg *dns.Msg) error { m.msg = msg; return nil }
func (m *mockResponseWriter) Write([]byte) (int, error)   { return 0, nil }
func (m *mockResponseWriter) Close() error                { return nil }
func (m *mockResponseWriter) TsigStatus() error           { return nil }
func (m *mockResponseWriter) TsigTimersOnly(bool)         {}
func (m *mockResponseWriter) Hijack()                     {}

