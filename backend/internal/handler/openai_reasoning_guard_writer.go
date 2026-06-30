package handler

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

type openAIReasoningGuardWriter struct {
	underlying gin.ResponseWriter
	headers    http.Header
	body       bytes.Buffer
	status     int
	written    bool
	size       int
}

func newOpenAIReasoningGuardWriter(underlying gin.ResponseWriter) *openAIReasoningGuardWriter {
	cloned := make(http.Header)
	if underlying != nil {
		for k, values := range underlying.Header() {
			cp := append([]string(nil), values...)
			cloned[k] = cp
		}
	}
	return &openAIReasoningGuardWriter{
		underlying: underlying,
		headers:    cloned,
		status:     http.StatusOK,
		size:       -1,
	}
}

func (w *openAIReasoningGuardWriter) Header() http.Header {
	return w.headers
}

func (w *openAIReasoningGuardWriter) WriteHeader(code int) {
	if code <= 0 {
		code = http.StatusOK
	}
	w.status = code
	w.written = true
	if w.size < 0 {
		w.size = 0
	}
}

func (w *openAIReasoningGuardWriter) WriteHeaderNow() {
	if !w.written {
		w.WriteHeader(w.status)
	}
}

func (w *openAIReasoningGuardWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.WriteHeader(w.status)
	}
	n, err := w.body.Write(data)
	if w.size < 0 {
		w.size = 0
	}
	w.size += n
	return n, err
}

func (w *openAIReasoningGuardWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *openAIReasoningGuardWriter) Status() int {
	return w.status
}

func (w *openAIReasoningGuardWriter) Size() int {
	return w.size
}

func (w *openAIReasoningGuardWriter) Written() bool {
	return w.written
}

func (w *openAIReasoningGuardWriter) Flush() {}

func (w *openAIReasoningGuardWriter) Unwrap() http.ResponseWriter {
	if w == nil {
		return nil
	}
	return w.underlying
}

func (w *openAIReasoningGuardWriter) Pusher() http.Pusher {
	if p, ok := w.underlying.(http.Pusher); ok {
		return p
	}
	return nil
}

func (w *openAIReasoningGuardWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.underlying.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func (w *openAIReasoningGuardWriter) CloseNotify() <-chan bool {
	if cn, ok := w.underlying.(http.CloseNotifier); ok {
		return cn.CloseNotify()
	}
	ch := make(chan bool, 1)
	return ch
}

func (w *openAIReasoningGuardWriter) replayToUnderlying() error {
	if w.underlying == nil {
		return nil
	}
	dst := w.underlying.Header()
	for k := range dst {
		delete(dst, k)
	}
	for k, values := range w.headers {
		dst[k] = append([]string(nil), values...)
	}
	if w.written {
		w.underlying.WriteHeader(w.status)
	}
	if w.body.Len() == 0 {
		return nil
	}
	_, err := io.Copy(w.underlying, bytes.NewReader(w.body.Bytes()))
	return err
}
