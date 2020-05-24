package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Setheck/tweetstreem/app/mocks"
	"github.com/stretchr/testify/mock"
)

func Test(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

		})
	}
}

func TestRemoteClient_RpcCall(t *testing.T) {
	tests := []struct {
		name         string
		clientLinks  bool
		input        string
		rpcDialError bool
		clientError  bool
	}{
		{"happy call", false, "testing", false, false},
		{"happy call local open", true, "testing", false, false},
		{"rpc requestError", true, "testing", true, false},
		{"client requestError", true, "testing", false, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expectedNetwork := "tcp"
			expectedAddress := "localhost:9999"

			var clientError error
			if test.clientError {
				clientError = assert.AnError
			}
			mockRpcClient := new(mocks.RpcClient)
			mockRpcClient.On("Call",
				"TweetStreem.RpcProcessCommand",
				mock.AnythingOfType("*app.Arguments"),
				mock.AnythingOfType("*app.Output")).
				Return(clientError).
				Run(func(args mock.Arguments) {
					output := args.Get(2).(*Output)
					output.Result = test.input
				})

			rpcDialHTTP = func(network, address string) (RpcClient, error) {
				assert.Equal(t, expectedNetwork, network)
				assert.Equal(t, expectedAddress, address)
				if test.rpcDialError {
					return nil, assert.AnError
				}
				return mockRpcClient, nil
			}
			openUrlLocal = func(url string) error {
				assert.Equal(t, url, test.input)
				return nil
			}

			ts := &TweetStreem{EnableClientLinks: test.clientLinks}
			client := NewRemoteClient(ts, expectedAddress)
			err := client.RpcCall("input")
			if test.rpcDialError || test.clientError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
