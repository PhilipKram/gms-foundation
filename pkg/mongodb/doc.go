// Package mongodb provides a wrapper for connecting to MongoDB with support
// for Client-Side Field Level Encryption (CSFLE). The Client type offers
// access to encrypted and unencrypted database handles, as well as manual
// encryption/decryption via the Encryption method. A plain fallback
// connection is created automatically when CSFLE is enabled so that queries
// can bypass decryption when needed.
package mongodb
