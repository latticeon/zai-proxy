package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"zai-proxy/internal/logger"
	"zai-proxy/internal/proxy"
	"zai-proxy/internal/separatorrule"
)

const useProxyHeader = "X-Use-Proxy"

type responseMetricsWriter struct {
	http.ResponseWriter
	statusCode int
	charCount  int
}

func newResponseMetricsWriter(w http.ResponseWriter) *responseMetricsWriter {
	return &responseMetricsWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (w *responseMetricsWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseMetricsWriter) Write(data []byte) (int, error) {
	w.charCount += utf8.RuneCount(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseMetricsWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func shouldUseProxy(r *http.Request) bool {
	return strings.TrimSpace(r.Header.Get(useProxyHeader)) == "1"
}

func shouldEnableSeparatorRule(r *http.Request) bool {
	return separatorrule.IsEnabled(r.Header.Get(separatorrule.Header))
}

func routeLabel(useProxy bool) string {
	if useProxy && proxy.HasAvailableProxies() {
		return "PROXY"
	}
	return "DIRECT"
}

func streamModeLabel(stream bool) string {
	if stream {
		return "stream"
	}
	return "non-stream"
}

func proxyModeLabel(useProxy bool) string {
	if useProxy && proxy.HasAvailableProxies() {
		return "proxy"
	}
	return "direct"
}

func separatorRuleLabel(enabled bool) string {
	if enabled {
		return "on"
	}
	return "off"
}

func logRequestLifecycleStart(label string, r *http.Request, useProxy, stream, separatorRuleEnabled bool, requestModel string) {
	logger.LogInfo("[%s] %s 请求开始 method=%s path=%s model=%s proxy=%s client=%s separator_rule=%s", routeLabel(useProxy), label, r.Method, r.URL.Path, requestModel, proxyModeLabel(useProxy), streamModeLabel(stream), separatorRuleLabel(separatorRuleEnabled))
}

func logRequestLifecycleFinish(label string, r *http.Request, useProxy, stream, separatorRuleEnabled, truncated bool, requestModel string) {
	result := "完成响应"
	if truncated {
		result = "响应被z.ai截断"
	}
	logger.LogInfo("[%s] %s 请求结束 method=%s path=%s model=%s proxy=%s client=%s separator_rule=%s result=%s", routeLabel(useProxy), label, r.Method, r.URL.Path, requestModel, proxyModeLabel(useProxy), streamModeLabel(stream), separatorRuleLabel(separatorRuleEnabled), result)
}

func logRequestSummary(label string, startedAt time.Time, statusCode, charCount int, r *http.Request, useProxy bool) {
	logger.LogInfo("[%s] %s %s %s status=%d chars=%d duration=%s", routeLabel(useProxy), label, r.Method, r.URL.Path, statusCode, charCount, formatDuration(time.Since(startedAt)))
}

func WithRequestLogging(label string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		metricsWriter := newResponseMetricsWriter(w)
		useProxy := shouldUseProxy(r)
		defer logRequestSummary(label, startedAt, metricsWriter.statusCode, metricsWriter.charCount, r, useProxy)
		next(metricsWriter, r)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return d.String()
	}
	ms := float64(d) / float64(time.Millisecond)
	return fmt.Sprintf("%.2fms", ms)
}
