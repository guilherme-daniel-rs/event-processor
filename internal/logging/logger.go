package logging

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type contextKey string

const traceKey contextKey = "trace"

type Trace struct {
	MessageID string
	Logs      []string
	mu        sync.Mutex
}

func WithTrace(ctx context.Context, messageID string) context.Context {
	return context.WithValue(ctx, traceKey, &Trace{
		MessageID: messageID,
		Logs:      []string{},
	})
}

func Append(ctx context.Context, format string, args ...interface{}) {
	t, ok := ctx.Value(traceKey).(*Trace)
	if !ok {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.Logs = append(t.Logs, fmt.Sprintf(format, args...))
}

func Flush(ctx context.Context, err error) {
	t, ok := ctx.Value(traceKey).(*Trace)
	if !ok {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	level := "INFO"
	if err != nil {
		level = "ERROR"
	}

	traceOutput := strings.Join(t.Logs, " | ")
	if err != nil {
		fmt.Printf("[%s] ID: %s | Trace: %s | Error: %v\n", level, t.MessageID, traceOutput, err)
		return
	}
	fmt.Printf("[%s] ID: %s | Trace: %s\n", level, t.MessageID, traceOutput)
}
