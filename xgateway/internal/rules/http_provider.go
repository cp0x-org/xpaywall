package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const internalAPIKeyHeader = "X-Api-Key"

var _ Provider = (*HttpProvider)(nil)

type HttpProvider struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewHttpProvider(baseURL, apiKey string) *HttpProvider {
	return &HttpProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

type resolveRouteResponse struct {
	ProjectID       uuid.UUID    `json:"project_id"`
	OutboundRouteID uuid.UUID    `json:"outbound_route_id"`
	Name            string       `json:"name"`
	InboundPath     string       `json:"inbound_path"`
	Target          string       `json:"target"`
	AuthHeaderName  string       `json:"auth_header_name"`
	AuthHeaderValue string       `json:"auth_header_value"`
	AllowUnmatched  bool         `json:"allow_unmatched"`
	Price           string       `json:"price"`
	Free            bool         `json:"free"`
	MimeType        string       `json:"mime_type"`
	Description     string       `json:"description"`
	PaymentChannels []channelDTO `json:"payment_channels"`
}

type channelDTO struct {
	Protocol      string            `json:"protocol"`
	Method        string            `json:"method"`
	Scheme        string            `json:"scheme"`
	Price         string            `json:"price"`
	Enabled       bool              `json:"enabled"`
	ChannelConfig map[string]string `json:"channel_config"`
	ChannelID     uuid.UUID         `json:"channel_id"`
	AssetID       *uuid.UUID        `json:"asset_id,omitempty"`
}

// GetByInboundPath resolves a full inbound path (/{slug}/{path}) via the control-api.
func (p *HttpProvider) GetByInboundPath(ctx context.Context, inboundPath string) (*Rule, error) {
	if !strings.HasPrefix(inboundPath, "/") {
		inboundPath = "/" + inboundPath
	}

	url := p.baseURL + "/proxy/resolve" + inboundPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if p.apiKey != "" {
		req.Header.Set(internalAPIKeyHeader, p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("control-api status %d", resp.StatusCode)
	}

	var dto resolveRouteResponse
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return fromResolveResponse(dto), nil
}

func fromResolveResponse(dto resolveRouteResponse) *Rule {
	channels := make([]*PaymentChannel, 0, len(dto.PaymentChannels))
	for _, ch := range dto.PaymentChannels {
		if !ch.Enabled {
			continue
		}
		channels = append(channels, &PaymentChannel{
			ID:            ch.ChannelID,
			AssetID:       ch.AssetID,
			Protocol:      ch.Protocol,
			Method:        ch.Method,
			Scheme:        ch.Scheme,
			Price:         ch.Price,
			Enabled:       ch.Enabled,
			ChannelConfig: ch.ChannelConfig,
		})
	}

	return &Rule{
		ProjectID:       dto.ProjectID,
		OutboundRouteID: dto.OutboundRouteID,
		Name:            dto.Name,
		InboundPath:     dto.InboundPath,
		OutboundPath:    dto.Target,
		Target:          dto.Target,
		AuthHeaderName:  dto.AuthHeaderName,
		AuthHeaderValue: dto.AuthHeaderValue,
		AllowUnmatched:  dto.AllowUnmatched,
		Price:           dto.Price,
		Free:            dto.Free,
		MimeType:        dto.MimeType,
		Description:     dto.Description,
		PaymentChannels: channels,
	}
}
