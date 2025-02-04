// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package dns

import (
	"net"
	"testing"
	"time"

	"github.com/hashicorp/consul/agent/discovery"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_HandleRequest_ServiceQuestions(t *testing.T) {
	testCases := []HandleTestCase{
		// Service Lookup
		{
			name: "When no data is return from a query, send SOA",
			request: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Opcode: dns.OpcodeQuery,
				},
				Question: []dns.Question{
					{
						Name:   "foo.service.consul.",
						Qtype:  dns.TypeA,
						Qclass: dns.ClassINET,
					},
				},
			},
			configureDataFetcher: func(fetcher discovery.CatalogDataFetcher) {
				fetcher.(*discovery.MockCatalogDataFetcher).
					On("FetchEndpoints", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, discovery.ErrNoData).
					Run(func(args mock.Arguments) {
						req := args.Get(1).(*discovery.QueryPayload)
						reqType := args.Get(2).(discovery.LookupType)

						require.Equal(t, discovery.LookupTypeService, reqType)
						require.Equal(t, "foo", req.Name)
					})
			},
			validateAndNormalizeExpected: true,
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Opcode:        dns.OpcodeQuery,
					Response:      true,
					Authoritative: true,
					Rcode:         dns.RcodeSuccess,
				},
				Compress: true,
				Question: []dns.Question{
					{
						Name:   "foo.service.consul.",
						Qtype:  dns.TypeA,
						Qclass: dns.ClassINET,
					},
				},
				Ns: []dns.RR{
					&dns.SOA{
						Hdr: dns.RR_Header{
							Name:   "consul.",
							Rrtype: dns.TypeSOA,
							Class:  dns.ClassINET,
							Ttl:    4,
						},
						Ns:      "ns.consul.",
						Serial:  uint32(time.Now().Unix()),
						Mbox:    "hostmaster.consul.",
						Refresh: 1,
						Expire:  3,
						Retry:   2,
						Minttl:  4,
					},
				},
			},
		},
		{
			// TestDNS_ExternalServiceToConsulCNAMELookup
			name: "req type: service / question type: SRV / CNAME required: no",
			request: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Opcode: dns.OpcodeQuery,
				},
				Question: []dns.Question{
					{
						Name:  "alias.service.consul.",
						Qtype: dns.TypeSRV,
					},
				},
			},
			configureDataFetcher: func(fetcher discovery.CatalogDataFetcher) {
				fetcher.(*discovery.MockCatalogDataFetcher).
					On("FetchEndpoints", mock.Anything,
						&discovery.QueryPayload{
							Name:    "alias",
							Tenancy: discovery.QueryTenancy{},
						}, discovery.LookupTypeService).
					Return([]*discovery.Result{
						{
							Type:    discovery.ResultTypeVirtual,
							Service: &discovery.Location{Name: "alias", Address: "web.service.consul"},
							Node:    &discovery.Location{Name: "web", Address: "web.service.consul"},
						},
					},
						nil).On("FetchEndpoints", mock.Anything,
					&discovery.QueryPayload{
						Name:    "web",
						Tenancy: discovery.QueryTenancy{},
					}, discovery.LookupTypeService).
					Return([]*discovery.Result{
						{
							Type:    discovery.ResultTypeNode,
							Service: &discovery.Location{Name: "web", Address: "webnode"},
							Node:    &discovery.Location{Name: "webnode", Address: "127.0.0.2"},
						},
					}, nil).On("ValidateRequest", mock.Anything,
					mock.Anything).Return(nil).On("NormalizeRequest", mock.Anything)
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response:      true,
					Authoritative: true,
				},
				Compress: true,
				Question: []dns.Question{
					{
						Name:  "alias.service.consul.",
						Qtype: dns.TypeSRV,
					},
				},
				Answer: []dns.RR{
					&dns.SRV{
						Hdr: dns.RR_Header{
							Name:   "alias.service.consul.",
							Rrtype: dns.TypeSRV,
							Class:  dns.ClassINET,
							Ttl:    123,
						},
						Target:   "web.service.consul.",
						Priority: 1,
					},
				},
				Extra: []dns.RR{
					&dns.A{
						Hdr: dns.RR_Header{
							Name:   "web.service.consul.",
							Rrtype: dns.TypeA,
							Class:  dns.ClassINET,
							Ttl:    123,
						},
						A: net.ParseIP("127.0.0.2"),
					},
				},
			},
		},
	}

	testCases = append(testCases, getAdditionalTestCases(t)...)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runHandleTestCases(t, tc)
		})
	}
}
