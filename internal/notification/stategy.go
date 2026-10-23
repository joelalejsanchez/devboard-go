// Package notification - interfaces de notificacion
package notification

import (
	"context"
	"log/slog"
)

// Notifier contrato de notificacion
type Notifier interface {
	Notify(ctx context.Context, userID string, message string) error
}

// LogNotifier Notifier de logs
type LogNotifier struct {
	logger *slog.Logger
}

// Notify envio de notificaciones mediante logs
func (n *LogNotifier) Notify(ctx context.Context, userID string, message string) error {
	n.logger.InfoContext(ctx, "notificacion enviada",
		slog.String("user_id", userID),
		slog.String("message", message),
	)

	return nil
}

// NewLogNotifier constructor de LogNotifier
func NewLogNotifier(logger *slog.Logger) *LogNotifier {
	return &LogNotifier{logger: logger}
}

// NoOpNotifier para pruebas
type NoOpNotifier struct{}

// Notify envio nulo de notificacion
func (n *NoOpNotifier) Notify(_ context.Context, _ string, _ string) error {
	return nil
}
