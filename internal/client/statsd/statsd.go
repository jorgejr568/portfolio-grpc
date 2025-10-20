package statsd

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type Client interface {
	Increment(name string, tags ...string) error
	Count(name string, value int64, tags ...string) error
	Timing(name string, duration time.Duration, tags ...string) error
	Start(serviceName, methodName string) RequestTracker
	Close() error
}

type RequestTracker interface {
	Succeeded() error
	Failed() error
	FailedWithError(err error) error
	Finished() error
}

// client represents a StatsD client that sends metrics over UDP
type client struct {
	conn   net.Conn
	prefix string
}

// Config holds the configuration for the StatsD client
type Config struct {
	Host   string // StatsD server host (e.g., "localhost")
	Port   int    // StatsD server port (e.g., 8125)
	Prefix string // Metric prefix (e.g., "portfolio-grpc")
}

// New creates a new StatsD client with UDP connection
func New(config Config) (Client, error) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := net.DialTimeout("udp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to StatsD server at %s: %w", addr, err)
	}

	return &client{
		conn:   conn,
		prefix: config.Prefix,
	}, nil
}

// Close closes the UDP connection
func (c *client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// send sends a metric to StatsD over UDP
func (c *client) send(metric string) error {
	_, err := c.conn.Write([]byte(metric))
	return err
}

// Increment increments a counter metric by 1 with optional tags
func (c *client) Increment(name string, tags ...string) error {
	return c.Count(name, 1, tags...)
}

// Count increments a counter metric by a specific value with optional tags
func (c *client) Count(name string, value int64, tags ...string) error {
	metric := fmt.Sprintf("%s.%s:%d|c", c.prefix, name, value)
	if len(tags) > 0 {
		metric += "|#" + strings.Join(tags, ",")
	}
	metric += "\n"
	return c.send(metric)
}

// Timing sends a timing metric in milliseconds with optional tags
func (c *client) Timing(name string, duration time.Duration, tags ...string) error {
	ms := duration.Milliseconds()
	metric := fmt.Sprintf("%s.%s:%d|ms", c.prefix, name, ms)
	if len(tags) > 0 {
		metric += "|#" + strings.Join(tags, ",")
	}
	metric += "\n"
	return c.send(metric)
}

// requestTracker tracks metrics for a single request
type requestTracker struct {
	client      *client
	serviceName string
	methodName  string
	startTime   time.Time
}

// StartRequest creates a new request tracker and sends the "started" metric
func (c *client) Start(serviceName, methodName string) RequestTracker {
	tracker := &requestTracker{
		client:      c,
		serviceName: serviceName,
		methodName:  methodName,
		startTime:   time.Now(),
	}

	// Send started metric
	_ = c.Increment(tracker.metricName("started"))

	return tracker
}

// metricName builds the full metric name
func (rt *requestTracker) metricName(suffix string) string {
	return fmt.Sprintf("api.%s.%s.%s", rt.serviceName, rt.methodName, suffix)
}

// Succeeded marks the request as succeeded and sends timing
func (rt *requestTracker) Succeeded() error {
	duration := time.Since(rt.startTime)

	// Send succeeded counter
	if err := rt.client.Increment(rt.metricName("succeeded")); err != nil {
		return err
	}

	// Send timing with success tag
	return rt.client.Timing(rt.metricName("timing"), duration, "status:success")
}

// Failed marks the request as failed and sends timing
func (rt *requestTracker) Failed() error {
	duration := time.Since(rt.startTime)

	// Send failed counter
	if err := rt.client.Increment(rt.metricName("failed")); err != nil {
		return err
	}

	// Send timing with failure tag
	return rt.client.Timing(rt.metricName("timing"), duration, "status:failed")
}

// FailedWithError marks the request as failed with error details as tags
func (rt *requestTracker) FailedWithError(err error) error {
	duration := time.Since(rt.startTime)

	// Sanitize error for tag
	errorType := "unknown"
	if err != nil {
		errorType = sanitizeTag(err.Error())
	}

	// Send failed counter with error tag
	if err := rt.client.Increment(rt.metricName("failed"), fmt.Sprintf("error:%s", errorType)); err != nil {
		return err
	}

	// Send timing with failure and error tags
	return rt.client.Timing(rt.metricName("timing"), duration, "status:failed", fmt.Sprintf("error:%s", errorType))
}

// Finished marks the request as finished (neutral completion) and sends timing
func (rt *requestTracker) Finished() error {
	duration := time.Since(rt.startTime)
	return rt.client.Timing(rt.metricName("timing"), duration, "status:finished")
}

// sanitizeTag converts strings to safe tag values
func sanitizeTag(input string) string {
	// Take first 50 characters
	if len(input) > 50 {
		input = input[:50]
	}

	// Replace invalid characters with underscores
	result := make([]byte, 0, len(input))
	for i := 0; i < len(input); i++ {
		c := input[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			result = append(result, c)
		} else if c == ' ' || c == ':' {
			result = append(result, '_')
		}
	}

	if len(result) == 0 {
		return "unknown"
	}
	return string(result)
}
