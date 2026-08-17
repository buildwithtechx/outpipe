package engine

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
)

func CopyWithLimit(ctx context.Context, destination io.Writer, source io.Reader, limit int64) (int64, error) {

	if limit < 0 {
		return 0, fmt.Errorf("stream limit cannot be negative")
	}

	buffer := make([]byte, 32*1024)
	var copied int64

	for {

		select {
		case <-ctx.Done():
			return copied, ctx.Err()
		default:
		}

		read, readErr := source.Read(buffer)

		if read > 0 {

			if limit > 0 && copied+int64(read) > limit {
				return copied, fmt.Errorf("stream limit exceeded")
			}

			written, writeErr := destination.Write(buffer[:read])
			copied += int64(written)

			if writeErr != nil {
				return copied, writeErr
			}

			if written != read {
				return copied, io.ErrShortWrite
			}
		}

		if readErr == io.EOF {
			return copied, nil
		}

		if readErr != nil {
			return copied, readErr
		}
	}
}

func PipeConnections(ctx context.Context, left net.Conn, right net.Conn, limit int64) error {

	if left == nil || right == nil {
		return fmt.Errorf("both connections are required")
	}

	results := make(chan error, 2)
	var once sync.Once
	closeConnections := func() {
		once.Do(func() {
			left.Close()
			right.Close()
		})
	}
	go func() {
		_, err := CopyWithLimit(ctx, left, right, limit)
		closeConnections()
		results <- err
	}()
	go func() {
		_, err := CopyWithLimit(ctx, right, left, limit)
		closeConnections()
		results <- err
	}()
	first := <-results
	second := <-results

	if first != nil && first != context.Canceled && first != context.DeadlineExceeded {
		return first
	}

	return second
}
