// Package dtx contains a variety of patterns in the field of distributed transaction.
package dtx

type correlationContextKeyType string

// CorrelationID is an identifier to correlate transactions in context.
const CorrelationID correlationContextKeyType = "CorrelationID"
