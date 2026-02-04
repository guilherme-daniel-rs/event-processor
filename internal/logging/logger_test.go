package logging

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraceLogging(t *testing.T) {
	t.Run("should store logs in context and append correctly", func(t *testing.T) {
		ctx := context.Background()
		msgID := "test-msg-123"

		ctx = WithTrace(ctx, msgID)

		Append(ctx, "Log one")
		Append(ctx, "Log two: %d", 2)

		trace, ok := ctx.Value(traceKey).(*Trace)
		assert.True(t, ok)
		assert.Equal(t, msgID, trace.MessageID)
		assert.Len(t, trace.Logs, 2)
		assert.Equal(t, "Log one", trace.Logs[0])
		assert.Equal(t, "Log two: 2", trace.Logs[1])
	})

	t.Run("should do nothing when appending without trace in context", func(t *testing.T) {
		ctx := context.Background()
		assert.NotPanics(t, func() {
			Append(ctx, "This should not crash")
		})
	})

	t.Run("should flush logs correctly with info level", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		ctx := WithTrace(context.Background(), "id-info")
		Append(ctx, "step 1")
		Append(ctx, "step 2")

		Flush(ctx, nil)

		w.Close()
		var buf strings.Builder
		io.Copy(&buf, r)
		os.Stdout = old

		output := buf.String()
		assert.Contains(t, output, "[INFO]")
		assert.Contains(t, output, "ID: id-info")
		assert.Contains(t, output, "step 1 | step 2")
	})

	t.Run("should flush logs correctly with error level", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		ctx := WithTrace(context.Background(), "id-err")
		Append(ctx, "attempt failed")
		testErr := errors.New("something went wrong")

		Flush(ctx, testErr)

		w.Close()
		var buf strings.Builder
		io.Copy(&buf, r)
		os.Stdout = old

		output := buf.String()
		assert.Contains(t, output, "[ERROR]")
		assert.Contains(t, output, "ID: id-err")
		assert.Contains(t, output, "attempt failed")
		assert.Contains(t, output, "Error: something went wrong")
	})

	t.Run("should do nothing when flushing context without trace", func(t *testing.T) {
		ctx := context.Background()
		assert.NotPanics(t, func() {
			Flush(ctx, nil)
		})
	})
}
