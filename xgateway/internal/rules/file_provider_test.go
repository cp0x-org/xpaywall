package rules

import (
	"context"
	"testing"

	"github.com/cp0x-org/xpaywall/xgateway/internal/config"
)

func TestFileProviderList(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		X402: []config.X402Method{
			{
				Name:           "x402-base",
				FacilitatorURL: "https://facilitator.example.com",
				Network:        "base",
				Scheme:         "exact",
				Merchant:       "0x1234567890abcdef",
			},
		},
		Outbound: config.OutboundConfig{
			Target: "http://upstream.example.com",
			Rules: []config.Rule{
				{
					Name:           "health",
					Path:           "/health",
					Price:          "$0.001",
					Description:    "Health endpoint",
					PaymentMethods: []string{"x402-base"},
				},
			},
		},
	}

	provider := NewFileProvider(cfg)

	rules, err := provider.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if got, want := len(rules), 1; got != want {
		t.Fatalf("len(rules) = %d, want %d", got, want)
	}

	rule := rules[0]
	if rule.Name != "health" {
		t.Fatalf("rule.Name = %q, want %q", rule.Name, "health")
	}
	if rule.InboundPath != "/health" {
		t.Fatalf("rule.InboundPath = %q, want %q", rule.InboundPath, "/health")
	}
	if rule.Price != "$0.001" {
		t.Fatalf("rule.Price = %q, want %q", rule.Price, "$0.001")
	}
	if len(rule.PaymentChannels) != 1 {
		t.Fatalf("len(rule.PaymentChannels) = %d, want 1", len(rule.PaymentChannels))
	}
	ch := rule.PaymentChannels[0]
	if ch.Protocol != "x402" {
		t.Fatalf("ch.Protocol = %q, want x402", ch.Protocol)
	}
	if ch.ChannelConfig["network"] != "base" {
		t.Fatalf("ch.ChannelConfig[network] = %q, want base", ch.ChannelConfig["network"])
	}
}

func TestFileProviderListReturnsCopy(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		X402: []config.X402Method{
			{
				Name:           "x402-base",
				FacilitatorURL: "https://facilitator.example.com",
				Network:        "base",
				Scheme:         "exact",
				Merchant:       "0x1234567890abcdef",
			},
		},
		Outbound: config.OutboundConfig{
			Target: "http://upstream.example.com",
			Rules: []config.Rule{
				{
					Name:           "health",
					Path:           "/health",
					PaymentMethods: []string{"x402-base"},
				},
			},
		},
	}

	provider := NewFileProvider(cfg)

	rules, err := provider.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	rules[0].Name = "changed"

	rules2, err := provider.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if rules2[0].Name != "health" {
		t.Fatalf("rules2[0].Name = %q, want %q", rules2[0].Name, "health")
	}
}
