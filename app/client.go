package app

import (
	"fmt"
	"net/rpc"
)

type RemoteClient struct {
	addr string
}

func NewRemoteClient(addr string) RemoteClient {
	return RemoteClient{
		addr: addr,
	}
}

type Arguments struct {
	Input string
}

func (t RemoteClient) RpcCall(str string) error {
	client, err := rpc.DialHTTP("tcp", t.addr)
	if err != nil {
		return err
	}
	args := &Arguments{Input: str}
	var output *string
	err = client.Call("TweetStreem.RpcProcessCommand", args, &output)
	if output == nil {
		*output = ""
	}
	fmt.Println(*output)
	return err
}
