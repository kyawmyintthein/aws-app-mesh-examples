// Code generated by protoc-gen-twirplura v1.0.0, DO NOT EDIT.
// source: protos/colorteller/service.proto

package colorteller

import context "context"
import json "encoding/json"
import fmt "fmt"

import "github.com/kyawmyintthein/lura-twirp"
import "github.com/luraproject/lura/config"
import "github.com/luraproject/lura/logging"
import twirp "github.com/twitchtv/twirp"
import proto "google.golang.org/protobuf/proto"

// Version compatibility assertion.
// If the constant is not defined in the package, that likely means
// the package needs to be updated to work with this generated code.
// See https://twitchtv.github.io/twirp/docs/version_matrix.html
const _ = twirp.TwirpPackageMinVersion_8_1_0

// ==============================
// ColortellerService Lura Client
// ==============================

type colortellerServiceLuraClient struct {
	id      string
	service ColortellerService
	l       logging.Logger
}

// ==========================
// ColortellerService Methods
// ==========================

const (
	_ColortellerServiceMethod_GetColor = "GetColor"
	_ColortellerServiceMethod_GetStage = "GetStage"
	_ColortellerServiceMethod_Ping     = "Ping"
)

// ===========================================================================================================
// NewColortellerServiceLuraClient creates a Protobuf client that implements the ColortellerService interface.
// ===========================================================================================================

func NewColortellerServiceLuraClient(config *config.ServiceConfig, id string, client HTTPClient, l logging.Logger, opts ...twirp.ClientOption) (luratwirp.LuraTwirpStub, error) {
	baseURL, err := getBaseURLByColortellerServiceClientID(config)
	if err != nil {
		return nil, err
	}
	protobufClient := NewColortellerServiceProtobufClient(baseURL, client, opts...)
	return &colortellerServiceLuraClient{
		id:      id,
		service: protobufClient,
		l:       l,
	}, nil
}

// =======================================
// ColortellerService getBaseURLByClientID
// =======================================

func getBaseURLByColortellerServiceClientID(config *config.ServiceConfig) (string, error) {
	for _, endpoint := range config.Endpoints {
		_, ok := endpoint.ExtraConfig[luratwirp.TwirpServiceIdentifierConst].(string)
		if ok {
			for _, backend := range endpoint.Backend {
				_, ok := backend.ExtraConfig[luratwirp.TwirpServiceIdentifierConst].(string)
				if ok {
					if len(backend.Host) <= 0 {
						return "", twirp.InternalError("invalid host configuration")
					}
				}
				return backend.Host[0], nil
			}
		}
	}
	return "", twirp.InternalError(fmt.Sprintf("invalid %s", luratwirp.TwirpServiceIdentifierConst))
}

// ==============================================================
// Invoke invoke RPC function regarding given service and method.
// ==============================================================

func (c *colortellerServiceLuraClient) Invoke(ctx context.Context, service string, method string, in proto.Message) (proto.Message, error) {
	switch method {
	case _ColortellerServiceMethod_GetColor:
		req, ok := in.(*Empty)
		if !ok {
			return nil, twirp.InternalError("invalid protobuf message")
		}
		resp, err := c.service.GetColor(ctx, req)
		if err != nil {
			c.l.Error(err, "failed to invoke : ", _ColortellerServiceMethod_GetColor)
			return resp, err
		}
		return resp, err
	case _ColortellerServiceMethod_GetStage:
		req, ok := in.(*Empty)
		if !ok {
			return nil, twirp.InternalError("invalid protobuf message")
		}
		resp, err := c.service.GetStage(ctx, req)
		if err != nil {
			c.l.Error(err, "failed to invoke : ", _ColortellerServiceMethod_GetStage)
			return resp, err
		}
		return resp, err
	case _ColortellerServiceMethod_Ping:
		req, ok := in.(*Empty)
		if !ok {
			return nil, twirp.InternalError("invalid protobuf message")
		}
		resp, err := c.service.Ping(ctx, req)
		if err != nil {
			c.l.Error(err, "failed to invoke : ", _ColortellerServiceMethod_Ping)
			return resp, err
		}
		return resp, err
	}
	return nil, twirp.InternalError(fmt.Sprintf("invalid %s", luratwirp.TwirpServiceIdentifierConst))
}

// ===================================================================
// Identifier return client identifier to lura-twirp backend registery
// ===================================================================

func (c *colortellerServiceLuraClient) Identifier() string {
	return c.id
}

// ====================================
// Encode convert JSON to proto.Message
// ====================================

func (c *colortellerServiceLuraClient) Encode(ctx context.Context, method string, data []byte) (proto.Message, error) {
	switch method {
	case _ColortellerServiceMethod_GetColor:
		out := new(Empty)
		err := json.Unmarshal(data, out)
		if err != nil {
			c.l.Error(err, "failed to unmarhsal : ", _ColortellerServiceMethod_GetColor)
			return out, err
		}
		return out, err
	case _ColortellerServiceMethod_GetStage:
		out := new(Empty)
		err := json.Unmarshal(data, out)
		if err != nil {
			c.l.Error(err, "failed to unmarhsal : ", _ColortellerServiceMethod_GetStage)
			return out, err
		}
		return out, err
	case _ColortellerServiceMethod_Ping:
		out := new(Empty)
		err := json.Unmarshal(data, out)
		if err != nil {
			c.l.Error(err, "failed to unmarhsal : ", _ColortellerServiceMethod_Ping)
			return out, err
		}
		return out, err
	}
	return nil, twirp.InternalError(fmt.Sprintf("invalid method %s", method))
}

// ====================================
// Decode convert proto.Message to JSON
// ====================================

func (c *colortellerServiceLuraClient) Decode(ctx context.Context, method string, msg proto.Message) ([]byte, error) {
	switch method {
	case _ColortellerServiceMethod_GetColor:
		out, err := proto.Marshal(msg)
		if err != nil {
			c.l.Error(err, "failed to marshal : ", _ColortellerServiceMethod_GetColor)
			return out, err
		}
		return out, err
	case _ColortellerServiceMethod_GetStage:
		out, err := proto.Marshal(msg)
		if err != nil {
			c.l.Error(err, "failed to marshal : ", _ColortellerServiceMethod_GetStage)
			return out, err
		}
		return out, err
	case _ColortellerServiceMethod_Ping:
		out, err := proto.Marshal(msg)
		if err != nil {
			c.l.Error(err, "failed to marshal : ", _ColortellerServiceMethod_Ping)
			return out, err
		}
		return out, err
	}
	return nil, twirp.InternalError(fmt.Sprintf("invalid method %s", method))
}