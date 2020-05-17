package app

import (
	"context"
	"fmt"
	"net/rpc"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/Setheck/tweetstreem/app/mocks"
)

type TestReceiver struct {
	ResponseFmt string
}

func (tr TestReceiver) Ping(in, out *string) error {
	*out = fmt.Sprintf(tr.ResponseFmt, *in)
	return nil
}

func TestNewApi(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	api := NewApi(ctx, 5000, true)
	cancel()
	select {
	case <-api.ctx.Done():
	default:
		t.Fail()
	}
}

func TestApi_Rpc(t *testing.T) {
	port := 5000
	api := NewApi(nil, port, true)
	msgFormat := "pong %s"
	if err := api.Start(TestReceiver{msgFormat}); err != nil {
		t.Fatal(err)
	}
	defer api.Stop()
	<-time.After(time.Millisecond * 500) // wait for server to start, TODO:(smt) replace with backoff
	client, err := rpc.DialHTTP("tcp", fmt.Sprint(":", port))
	if err != nil {
		t.Fatal(err)
	}
	input := "test123"
	var got string
	if err := client.Call("TestReceiver.Ping", &input, &got); err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf(msgFormat, input)
	if got != want {
		fmt.Println("output mismatch", want, "!=", got)
		t.Fail()
	}
}

func TestApi_StartStop(t *testing.T) {
	mockServer := new(mocks.Server)
	mockServer.On("ListenAndServe").Return(nil)
	mockServer.On("Shutdown", mock.Anything).
		Return(nil)

	mockRpcServer := new(mocks.RpcServer)
	mockRpcServer.On("Register", mock.Anything).Return(nil)
	mockRpcServer.On("HandleHTTP", mock.AnythingOfType("string"), mock.AnythingOfType("string"))
	rpcServer = mockRpcServer

	api := NewApi(nil, 5000, true)
	api.server = mockServer

	msgFormat := "pong %s"
	if err := api.Start(TestReceiver{msgFormat}); err != nil {
		t.Fatal(err)
	}
	<-time.After(time.Millisecond * 500) // wait for server to start, TODO:(smt) replace with backoff
	api.Stop()

	mockServer.AssertCalled(t, "ListenAndServe")
	mockServer.AssertCalled(t, "Shutdown", mock.Anything)
	mockRpcServer.AssertCalled(t, "Register", mock.Anything)
	mockRpcServer.AssertCalled(t, "HandleHTTP", mock.AnythingOfType("string"), mock.AnythingOfType("string"))
}
