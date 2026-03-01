// Package uploads handles file upload persistence with MIME-type validation,
// size limits, and content-type verification via magic-byte detection.
//
// Files are read entirely into memory before being written to disk. This
// simplifies size checking and content validation but means memory usage
// scales with the configured MaxSize per category.
package uploads
