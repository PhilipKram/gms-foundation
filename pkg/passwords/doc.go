// Package passwords provides bcrypt hashing and verification helpers.
//
// Note: bcrypt silently truncates passwords longer than 72 bytes.
// Callers that accept arbitrary-length passwords should enforce or
// document this limit.
package passwords
