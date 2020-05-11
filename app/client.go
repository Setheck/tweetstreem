package app

import (
	"fmt"
	"net/rpc"
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

func (t RemoteClient) RpcCall(str string) error {
	client, err := rpc.DialHTTP("tcp", t.addr)
	if err != nil {
		return err
	}
	args := &Arguments{Input: str}
	output := &Output{}
	err = client.Call("TweetStreem.RpcProcessCommand", args, output)
	if len(output.Result) > 0 && t.TweetStreem.EnableClientLinks {
		fmt.Println("opening in browser:", output.Result)
		return OpenBrowser(output.Result)
	}
	return err
}
