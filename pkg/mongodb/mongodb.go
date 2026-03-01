package mongodb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const defaultPingTimeout = 5 * time.Second

// AuthConfig holds MongoDB authentication credentials.
type AuthConfig struct {
	Username  string
	Password  string
	Mechanism string // e.g. "SCRAM-SHA-256"
	Source    string // auth database; defaults to Config.Database if empty
}

// CSFLEConfig holds Client-Side Field Level Encryption settings.
type CSFLEConfig struct {
	KeyVaultNamespace string // e.g. "encryption.__keyVault"
	DEKName           string // logical name of the data encryption key
	MasterKey         []byte // 96-byte local master key
}

// DataEncryptionKeyName returns the logical name of the data encryption key
// associated with this CSFLE configuration. Callers can use this when
// performing explicit encryption or decryption via Encryption().
func (c *CSFLEConfig) DataEncryptionKeyName() string {
	if c == nil {
		return ""
	}
	return c.DEKName
}

// Config holds the settings needed to connect to MongoDB.
type Config struct {
	Host     string       // URI string, e.g. "mongodb://host:27017"
	Database string       // primary database name
	Auth     AuthConfig   // authentication credentials
	CSFLE    *CSFLEConfig // nil = CSFLE disabled
}

// clientConfig holds internal options applied via Option functions.
type clientConfig struct {
	PingTimeout      time.Duration
	AppName          string
	DirectConnection *bool
}

func defaultClientConfig() clientConfig {
	return clientConfig{
		PingTimeout: defaultPingTimeout,
	}
}

// Option customizes client configuration.
type Option func(*clientConfig)

// WithPingTimeout sets the timeout for the initial ping.
func WithPingTimeout(d time.Duration) Option {
	return func(c *clientConfig) { c.PingTimeout = d }
}

// WithAppName sets the application name in the MongoDB connection metadata.
func WithAppName(name string) Option {
	return func(c *clientConfig) { c.AppName = name }
}

// WithDirectConnection bypasses topology discovery when true.
func WithDirectConnection(direct bool) Option {
	return func(c *clientConfig) { c.DirectConnection = &direct }
}

// Client wraps a MongoDB client, database, and optional CSFLE encryption.
// When CSFLE is configured, PlainDB holds a second connection without
// auto-encryption so queries can fall back when decryption fails.
type Client struct {
	client      *mongo.Client
	db          *mongo.Database
	plainClient *mongo.Client
	plainDB     *mongo.Database
	encryption  *mongo.ClientEncryption
}

// DB returns the primary database handle.
func (c *Client) DB() *mongo.Database {
	return c.db
}

// PlainDB returns the database handle without CSFLE auto-encryption.
// If CSFLE is not configured, it returns the primary database handle.
func (c *Client) PlainDB() *mongo.Database {
	if c.plainDB != nil {
		return c.plainDB
	}
	return c.db
}

// Encryption returns the ClientEncryption handle for manual encrypt/decrypt.
// Returns nil if CSFLE is not configured.
func (c *Client) Encryption() *mongo.ClientEncryption {
	return c.encryption
}

// Close disconnects both clients and closes the encryption handle.
// All errors are collected and returned as a single joined error.
func (c *Client) Close(ctx context.Context) error {
	var errs []error
	if c.encryption != nil {
		if err := c.encryption.Close(ctx); err != nil {
			errs = append(errs, fmt.Errorf("closing encryption handle: %w", err))
		}
	}
	if c.plainClient != nil {
		if err := c.plainClient.Disconnect(ctx); err != nil {
			errs = append(errs, fmt.Errorf("disconnecting plain client: %w", err))
		}
	}
	if c.client != nil {
		if err := c.client.Disconnect(ctx); err != nil {
			errs = append(errs, fmt.Errorf("disconnecting client: %w", err))
		}
	}
	return errors.Join(errs...)
}

// buildClientOpts creates the base MongoDB client options from a Config.
func buildClientOpts(cfg Config, cc clientConfig) *options.ClientOptions {
	opts := options.Client().ApplyURI(cfg.Host)

	if cfg.Auth.Username != "" {
		authSource := cfg.Auth.Source
		if authSource == "" {
			authSource = cfg.Database
		}
		opts.SetAuth(options.Credential{
			Username:      cfg.Auth.Username,
			Password:      cfg.Auth.Password,
			AuthMechanism: cfg.Auth.Mechanism,
			AuthSource:    authSource,
		})
	}

	if cc.AppName != "" {
		opts.SetAppName(cc.AppName)
	}
	if cc.DirectConnection != nil {
		opts.SetDirect(*cc.DirectConnection)
	}

	return opts
}

// kmsProviders returns the KMS providers map for local CSFLE.
func kmsProviders(masterKey []byte) map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"local": {
			"key": bson.Binary{
				Subtype: 0x00,
				Data:    masterKey,
			},
		},
	}
}

// Connect creates a new MongoDB connection, pings it, and returns a Client.
// When cfg.CSFLE is non-nil, auto-encryption (bypass mode) is configured and a
// second plain connection is created for fallback reads without decryption.
func Connect(ctx context.Context, cfg Config, opts ...Option) (*Client, error) {
	cc := defaultClientConfig()
	for _, o := range opts {
		o(&cc)
	}

	clientOpts := buildClientOpts(cfg, cc)
	csfleEnabled := cfg.CSFLE != nil

	if csfleEnabled && len(cfg.CSFLE.MasterKey) != 96 {
		return nil, fmt.Errorf("CSFLE MasterKey must be exactly 96 bytes, got %d", len(cfg.CSFLE.MasterKey))
	}

	if csfleEnabled {
		autoEncryptionOpts := options.AutoEncryption().
			SetKeyVaultNamespace(cfg.CSFLE.KeyVaultNamespace).
			SetKmsProviders(kmsProviders(cfg.CSFLE.MasterKey)).
			SetBypassAutoEncryption(true)
		clientOpts.SetAutoEncryptionOptions(autoEncryptionOpts)
	}

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("connecting to MongoDB: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cc.PingTimeout)
	defer cancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("pinging MongoDB: %w", err)
	}

	c := &Client{
		client: client,
		db:     client.Database(cfg.Database),
	}

	if csfleEnabled {
		plainOpts := buildClientOpts(cfg, cc)
		plainClient, err := mongo.Connect(plainOpts)
		if err != nil {
			_ = c.Close(ctx)
			return nil, fmt.Errorf("connecting plain MongoDB client: %w", err)
		}
		c.plainClient = plainClient
		c.plainDB = plainClient.Database(cfg.Database)

		ceOpts := options.ClientEncryption().
			SetKeyVaultNamespace(cfg.CSFLE.KeyVaultNamespace).
			SetKmsProviders(kmsProviders(cfg.CSFLE.MasterKey))

		ce, err := mongo.NewClientEncryption(client, ceOpts)
		if err != nil {
			_ = c.Close(ctx)
			return nil, fmt.Errorf("creating client encryption: %w", err)
		}
		c.encryption = ce
	}

	return c, nil
}

// IsMongocryptError returns true if the error is a mongocrypt decryption failure.
func IsMongocryptError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "mongocrypt")
}
