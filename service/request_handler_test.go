package service

import (
	"crypto/tls"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/plugin"

	"context"
	"net/http"
	"testing"
)

func TestNonTLSTransportSelection(t *testing.T) {
	var testKVS = initKVStore(t)
	be, err := buildBackend(testKVS, "be1")
	if assert.Nil(t, err) {
		tlsConfig := &tls.Config{RootCAs: be.CACert}

		requestHandler := &requestHandler{
			Transport:    &http.Transport{DisableKeepAlives: false, DisableCompression: false},
			TLSTransport: &http.Transport{DisableKeepAlives: false, DisableCompression: false, TLSClientConfig: tlsConfig},
			Backend:      be,
		}

		ctx := context.Background()
		transport := requestHandler.getTransportForBackend(ctx)
		assert.Equal(t, requestHandler.Transport, transport)
	}

}

func TestNonTLSOnlyTransportSelection(t *testing.T) {
	var testKVS = initKVStore(t)
	be, err := buildBackend(testKVS, "be-tls")
	if assert.Nil(t, err) {
		tlsConfig := &tls.Config{RootCAs: be.CACert}

		requestHandler := &requestHandler{
			Transport:    &http.Transport{DisableKeepAlives: false, DisableCompression: false},
			TLSTransport: &http.Transport{DisableKeepAlives: false, DisableCompression: false, TLSClientConfig: tlsConfig},
			Backend:      be,
		}

		ctx := context.Background()
		transport := requestHandler.getTransportForBackend(ctx)
		assert.Equal(t, requestHandler.TLSTransport, transport)
	}

}

func TestNonTLSTransportHttpsContext(t *testing.T) {
	var testKVS = initKVStore(t)
	be, err := buildBackend(testKVS, "be1")
	if assert.Nil(t, err) {
		tlsConfig := &tls.Config{RootCAs: be.CACert}

		requestHandler := &requestHandler{
			Transport:    &http.Transport{DisableKeepAlives: false, DisableCompression: false},
			TLSTransport: &http.Transport{DisableKeepAlives: false, DisableCompression: false, TLSClientConfig: tlsConfig},
			Backend:      be,
		}

		ctx := context.Background()
		ctx = plugin.AddUseHttpsToContext(ctx, true)
		transport := requestHandler.getTransportForBackend(ctx)
		assert.Equal(t, requestHandler.TLSTransport, transport)
	}

}
