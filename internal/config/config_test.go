package config

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func lookupMap(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) { value, ok := values[key]; return value, ok }
}

func TestDefaults(t *testing.T) {
	c, err := load(lookupMap(nil))
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTPAddr != "127.0.0.1:8000" || c.ReadHeaderTimeout != 5*time.Second || c.MaxBodyBytes != 1048576 {
		t.Fatalf("unexpected safe defaults: %#v", c)
	}
	if c.DatabaseURL != "" || c.LicenseFile != "/etc/mile/mile_enterprise.lic" || !strings.HasSuffix(c.LicenseFile, ".lic") {
		t.Fatalf("unsafe database default or unexpected license path: %q", c.LicenseFile)
	}
}

func TestSecretFileLoading(t *testing.T) {
	path := filepath.Join(t.TempDir(), "database-secret")
	secret := "postgres://mile:secret@localhost:5432/mile"
	if err := os.WriteFile(path, []byte(secret+"\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := load(lookupMap(map[string]string{
		"MILE_DATABASE_URL_FILE":     path,
		"MILE_HTTP_ADDR":             "localhost:9000",
		"MILE_HTTP_READ_TIMEOUT_SEC": "12",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if c.DatabaseURL != secret || c.HTTPAddr != "localhost:9000" || c.ReadTimeout != 12*time.Second {
		t.Fatal("secret file loading failed")
	}
}

// Nothing has shipped under an earlier name, so unprefixed and IMS_*/MINTERN_*
// variables must be inert: an ambient DATABASE_URL or HTTP_ADDR from an unrelated
// process must neither reconfigure MILE nor collide with its own MILE_* setting.
func TestLegacyAndUnprefixedVariablesAreIgnored(t *testing.T) {
	c, err := load(lookupMap(map[string]string{
		"DATABASE_URL":              "postgres://other-app@db.invalid/other",
		"MILE_DATABASE_URL":         "postgres://mile@localhost:5432/mile",
		"HTTP_ADDR":                 "0.0.0.0:9999",
		"APP_DOMAIN":                "other.example.org",
		"REDIS_ADDR":                "redis.invalid:6379",
		"IMS_ENV":                   "production",
		"IMS_HTTP_ADDR":             "0.0.0.0:7777",
		"MINTERN_LICENSE_FILE":      "/tmp/stale.lic",
		"HTTP_READ_TIMEOUT_SEC":     "99",
		"IMS_HTTP_IDLE_TIMEOUT_SEC": "99",
	}))
	if err != nil {
		t.Fatalf("ambient non-MILE variables must not affect startup: %v", err)
	}
	if c.DatabaseURL != "postgres://mile@localhost:5432/mile" {
		t.Fatalf("MILE_DATABASE_URL was overridden: %q", c.DatabaseURL)
	}
	if c.Env != "development" || c.HTTPAddr != "127.0.0.1:8000" || c.AppDomain != "localhost" ||
		c.LicenseFile != "/etc/mile/mile_enterprise.lic" ||
		c.ReadTimeout != 10*time.Second || c.IdleTimeout != 60*time.Second {
		t.Fatalf("legacy or unprefixed variable leaked into configuration: %#v", c)
	}
}

func TestInvalidExplicitConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values map[string]string
		want   string
	}{
		{"malformed duration", map[string]string{"MILE_HTTP_READ_TIMEOUT_SEC": "ten"}, "MILE_HTTP_READ_TIMEOUT_SEC"},
		{"overflow", map[string]string{"MILE_HTTP_READ_TIMEOUT_SEC": "9223372037"}, "MILE_HTTP_READ_TIMEOUT_SEC"},
		{"zero", map[string]string{"MILE_HTTP_IDLE_TIMEOUT_SEC": "0"}, "MILE_HTTP_IDLE_TIMEOUT_SEC"},
		{"empty", map[string]string{"MILE_HTTP_ADDR": ""}, "must not be empty"},
		{"unknown env", map[string]string{"MILE_ENV": "prod"}, "MILE_ENV"},
		{"unknown log", map[string]string{"MILE_LOG_LEVEL": "verbose"}, "MILE_LOG_LEVEL"},
		{"bad port", map[string]string{"MILE_HTTP_ADDR": "localhost:65536"}, "MILE_HTTP_ADDR"},
		{"file conflict", map[string]string{"MILE_DATABASE_URL": "postgres://localhost/mile", "MILE_DATABASE_URL_FILE": "secret"}, "mutually exclusive"},
		{"missing file", map[string]string{"MILE_METRICS_TOKEN_FILE": filepath.Join(t.TempDir(), "absent")}, "could not be read"},
		{"body limit", map[string]string{"MILE_HTTP_MAX_BODY_BYTES": "-1"}, "MILE_HTTP_MAX_BODY_BYTES"},
		{"public metrics", map[string]string{"MILE_HTTP_ADDR": ":8000"}, "MILE_METRICS_TOKEN"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := load(lookupMap(tc.values))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestProductionRequiresIdentityAndMetricsProtection(t *testing.T) {
	values := map[string]string{"MILE_ENV": "production"}
	if _, err := load(lookupMap(values)); err == nil {
		t.Fatal("production defaults unexpectedly accepted")
	}
	values["MILE_APP_DOMAIN"] = "mile.example.org"
	values["MILE_METRICS_TOKEN"] = strings.Repeat("x", 32)
	values["MILE_HTTP_ADDR"] = ":8000"
	if _, err := load(lookupMap(values)); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsDoNotAppearInErrorsOrStructuredLogs(t *testing.T) {
	secret := "secret-that-must-not-be-logged"
	_, err := load(lookupMap(map[string]string{"MILE_DATABASE_URL": "://" + secret}))
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("unsafe validation error: %v", err)
	}
	c, err := load(lookupMap(map[string]string{"MILE_DATABASE_URL": "postgres://mile:" + secret + "@localhost/mile", "MILE_METRICS_TOKEN": secret}))
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	slog.New(slog.NewJSONHandler(&output, nil)).Info("configuration", "config", c)
	if strings.Contains(output.String(), secret) || strings.Contains(output.String(), "postgres://") {
		t.Fatal("configuration leaked secrets into logs")
	}
}

func TestSecretFileBounds(t *testing.T) {
	for _, data := range []string{"\r\n", strings.Repeat("x", 65537)} {
		path := filepath.Join(t.TempDir(), "secret")
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := load(lookupMap(map[string]string{"MILE_METRICS_TOKEN_FILE": path})); err == nil {
			t.Fatal("invalid secret file accepted")
		}
	}
}
