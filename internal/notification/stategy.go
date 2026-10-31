// Package notification defines notification interfaces and implementations.
package notification

import (
	"context"
	"log/slog"
)

// Notifier sends a notification to a user.
type Notifier interface {
	Notify(ctx context.Context, userID string, message string) error
}

// LogNotifier writes notifications to the application logger.
type LogNotifier struct {
	logger *slog.Logger
}

// Notify logs a notification for the specified user.
func (n *LogNotifier) Notify(ctx context.Context, userID string, message string) error {
	n.logger.InfoContext(ctx, "notificacion enviada",
		slog.String("user_id", userID),
		slog.String("message", message),
	)

	return nil
}

// NewLogNotifier creates a logger-backed notifier.
func NewLogNotifier(logger *slog.Logger) *LogNotifier {
	return &LogNotifier{logger: logger}
}

// NoOpNotifier discards notifications.
type NoOpNotifier struct{}

// Notify discards the notification and returns no error.
func (n *NoOpNotifier) Notify(_ context.Context, _ string, _ string) error {
	return nil
}
