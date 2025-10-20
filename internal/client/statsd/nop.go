package statsd

import (
	"time"
)

type nopClient struct{}

func (n nopClient) Increment(_ string, _ ...string) error {
	return nil
}

func (n nopClient) Count(_ string, _ int64, _ ...string) error {
	return nil
}

func (n nopClient) Timing(_ string, _ time.Duration, _ ...string) error {
	return nil
}

func (n nopClient) Start(_ string, _ string) RequestTracker {
	return &nopRequestTracker{}
}

func (n nopClient) Close() error {
	return nil
}

func Nop() Client {
	return &nopClient{}
}

type nopRequestTracker struct{}

func (n nopRequestTracker) Finished() error {
	return nil
}

func (n nopRequestTracker) Succeeded() error {
	return nil
}

func (n nopRequestTracker) Failed() error {
	return nil
}

func (n nopRequestTracker) FailedWithError(_ error) error {
	return nil
}
