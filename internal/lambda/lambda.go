// Package lambda defines top level Lambda handlers
package lambda

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
)

// EchoHandler the Echo specific Lambda handler.
type EchoHandler struct {
	echoAdapter *echoadapter.EchoLambda
}

// NewEchoHandler creates a neww echo handler.
func NewEchoHandler(echoAdapter *echoadapter.EchoLambda) *EchoHandler {
	return &EchoHandler{
		echoAdapter: echoAdapter,
	}
}

// Handle proxies an API Gateway request.
func (h *EchoHandler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return h.echoAdapter.ProxyWithContext(ctx, request)
}
