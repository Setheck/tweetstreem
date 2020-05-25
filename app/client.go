package app

import (
	"fmt"
	"net/rpc"

	"github.com/Setheck/tweetstreem/util"
)

type RemoteClient struct {
	addr string
	*TweetStreem
}

func NewRemoteClient(ts *TweetStreem, addr string) RemoteClient {
	return RemoteClient{
		addr:        addr,
		TweetStreem: ts,
	}
}

type Arguments struct {
	Input string
}

type Output struct {
	Result string
}

type RpcClient interface {
	Call(serviceMethod string, args interface{}, reply interface{}) error
}

type ClientProvider interface {
	DialHTTP(string, string) (RpcClient, error)
}

var rpcDialHTTP = func(network, address string) (RpcClient, error) { return rpc.DialHTTP(network, address) }
var openUrlLocal = util.OpenBrowser

func (t RemoteClient) RpcCall(str string) error {
	client, err := rpcDialHTTP("tcp", t.addr)
	if err != nil {
		return err
	}
	args := &Arguments{Input: str}
	output := &Output{}
	err = client.Call("TweetStreem.RpcProcessCommand", args, output)
	if err == nil && len(output.Result) > 0 && t.TweetStreem.EnableClientLinks {
		fmt.Println("opening in browser:", output.Result)
		return openUrlLocal(output.Result)
	}
	return err
}
