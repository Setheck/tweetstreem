package app

import (
	"fmt"
	"net/rpc"
)

type RemoteClient struct {
}

type Arguments struct {
	Input string
}

type Response struct {
}

func (t RemoteClient) RpcCall(str string) error {
	client, err := rpc.DialHTTP("tcp", "localhost:8080")
	if err != nil {
		return err
	}
	args := &Arguments{
		Input: str,
	}
	var output *string
	fmt.Println(args.Input)
	err = client.Call("TweetStreem.RpcProcessCommand", args, &output)
	fmt.Println(*output)
	return err
}
