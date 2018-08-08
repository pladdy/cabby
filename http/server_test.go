package http

import (
	"crypto/tls"
	"testing"
)

func TestSetupTLS(t *testing.T) {
	tlsSetup := setupTLS()

	if tlsSetup.MinVersion != tls.VersionTLS12 {
		t.Error("Got:", tlsSetup.MinVersion, "Expected:", tls.VersionTLS12)
	}

	expectedCurves := map[tls.CurveID]bool{
		tls.CurveP521: true,
		tls.CurveP384: true,
		tls.CurveP256: true,
	}
	for _, curve := range tlsSetup.CurvePreferences {
		if !expectedCurves[curve] {
			t.Error("Invalid CurvePreference:", curve)
		}
	}

	if tlsSetup.PreferServerCipherSuites != true {
		t.Error("Got:", tlsSetup.PreferServerCipherSuites, "Expected:", true)
	}

	expectedCipherSuites := map[uint16]bool{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384: true,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:    true,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384:       true,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA:          true,
	}
	for _, cipherSuite := range tlsSetup.CipherSuites {
		if !expectedCipherSuites[cipherSuite] {
			t.Error("Invalid CurvePreference:", cipherSuite)
		}
	}
}
