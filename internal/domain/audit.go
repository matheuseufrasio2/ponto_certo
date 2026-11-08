package domain

import (
	"context"
)

type auditContextKey struct{}

// AuditMetadata holds client networking telemetry for security and compliance audits.
type AuditMetadata struct {
	IPAddress string
	UserAgent string
}

// GetAuditMetadata retrieves client IP and User-Agent from context, or empty values if not present.
func GetAuditMetadata(ctx context.Context) AuditMetadata {
	if ctx == nil {
		return AuditMetadata{}
	}
	if meta, ok := ctx.Value(auditContextKey{}).(AuditMetadata); ok {
		return meta
	}
	return AuditMetadata{}
}

// WithAuditMetadata manually sets audit metadata on context.
func WithAuditMetadata(ctx context.Context, meta AuditMetadata) context.Context {
	return context.WithValue(ctx, auditContextKey{}, meta)
}
