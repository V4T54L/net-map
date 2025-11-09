package dns

import (
	"context"
	"errors"
	"fmt"
	"internal-dns/internal/domain"
	"internal-dns/internal/infrastructure/cache"
	"internal-dns/internal/usecase"
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
)

// Server is a DNS server implementation.
type Server struct {
	uc    usecase.DNSRecordUseCase
	cache cache.DNSRecordCache
	addr  string
}

// NewServer creates a new DNS server.
func NewServer(addr string, uc usecase.DNSRecordUseCase, cache cache.DNSRecordCache) *Server {
	return &Server{
		uc:    uc,
		cache: cache,
		addr:  addr,
	}
}

// ListenAndServe starts the DNS server.
func (s *Server) ListenAndServe() error {
	dns.HandleFunc(".", s.handleRequest)
	server := &dns.Server{Addr: s.addr, Net: "udp"}
	log.Printf("DNS server listening on %s", s.addr)
	return server.ListenAndServe()
}

func (s *Server) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	ctx := context.Background()

	for _, q := range r.Question {
		// Normalize domain name: lowercase and ensure it's fully qualified.
		domainName := strings.ToLower(dns.Fqdn(q.Name))
		log.Printf("Received query for %s type %s", domainName, dns.TypeToString[q.Qtype])

		record, err := s.resolve(ctx, domainName)
		if err != nil {
			if !errors.Is(err, cache.ErrCacheMiss) {
				log.Printf("Error resolving domain %s: %v", domainName, err)
			}
			msg.SetRcode(r, dns.RcodeNameError) // NXDOMAIN
			_ = w.WriteMsg(msg)
			return
		}

		rr, err := s.buildRR(q, record)
		if err != nil {
			log.Printf("Error building resource record for %s: %v", domainName, err)
			msg.SetRcode(r, dns.RcodeServerFailure)
			_ = w.WriteMsg(msg)
			return
		}
		msg.Answer = append(msg.Answer, rr)
	}

	_ = w.WriteMsg(msg)
}

func (s *Server) resolve(ctx context.Context, domainName string) (*domain.DNSRecord, error) {
	// 1. Check cache
	cachedRecord, err := s.cache.Get(ctx, domainName)
	if err == nil {
		log.Printf("Cache hit for domain: %s", domainName)
		return cachedRecord, nil
	}
	if !errors.Is(err, cache.ErrCacheMiss) {
		log.Printf("Cache error for domain %s: %v", domainName, err)
		// Fall through to DB if cache fails
	}

	log.Printf("Cache miss for domain: %s", domainName)

	// 2. Check database via use case
	dbRecord, err := s.uc.ResolveDomain(ctx, domainName)
	if err != nil {
		return nil, err // Propagate repository.ErrDNSRecordNotFound
	}

	// 3. Set cache
	if err := s.cache.Set(ctx, dbRecord); err != nil {
		log.Printf("Failed to cache record for %s: %v", domainName, err)
	}

	return dbRecord, nil
}

func (s *Server) buildRR(q dns.Question, record *domain.DNSRecord) (dns.RR, error) {
	hdr := dns.RR_Header{Name: q.Name, Rrtype: q.Qtype, Class: dns.ClassINET, Ttl: 300}

	switch domain.RecordType(record.Type) {
	case domain.A:
		if q.Qtype != dns.TypeA {
			return nil, fmt.Errorf("record type mismatch: expected A, got %s", record.Type)
		}
		ip := net.ParseIP(record.Value)
		if ip == nil || ip.To4() == nil {
			return nil, fmt.Errorf("invalid IPv4 address in record value: %s", record.Value)
		}
		return &dns.A{Hdr: hdr, A: ip.To4()}, nil

	case domain.CNAME:
		if q.Qtype != dns.TypeCNAME {
			return nil, fmt.Errorf("record type mismatch: expected CNAME, got %s", record.Type)
		}
		return &dns.CNAME{Hdr: hdr, Target: dns.Fqdn(record.Value)}, nil
	}

	return nil, fmt.Errorf("unsupported record type: %s", record.Type)
}

