package server

import "context"

// TODO:
// TODO(lifetime): split liveness from readiness, and fail readiness while
// draining, so the load balancer stops sending traffic before the process goes.
func Health(_ context.Context) error {
	return nil
}
