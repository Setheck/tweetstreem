package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"sync"
	"sync/atomic"
	"time"

	_ "net/http/pprof"
)

var rpcServer RpcServer = rpc.DefaultServer

type RPCListener interface {
	Start(interface{}) error
	Stop()
}

type RpcServer interface {
	Register(interface{}) error
	HandleHTTP(string, string)
}

type Server interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

var _ RPCListener = &HttpRPCListener{}
var _ Server = &http.Server{}

type HttpRPCListener struct {
	Port int

	server Server
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	logging    bool
	errorCount uint64
}

// NewHttpRPCListener creates a new instance, defaults context to context.Background() if nil is provided.
func NewHttpRPCListener(ctx context.Context, port int, logging bool) *HttpRPCListener {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout:   time.Second * 100,
		ReadTimeout:    time.Second * 100,
		IdleTimeout:    time.Second * 600,
		MaxHeaderBytes: 1 << 20,
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	listener := &HttpRPCListener{
		Port: port,

		server:  server,
		ctx:     ctx,
		cancel:  cancel,
		logging: logging,
	}
	return listener
}

// Start starts the http rpc listener, registering the given receiver.
func (a *HttpRPCListener) Start(receiver interface{}) error {
	if err := rpcServer.Register(receiver); err != nil {
		return err
	}
	rpcServer.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	go a.handleShutdown()
	go a.listenAndServe()
	return nil
}

// Stop will cancel the internal context and wait for the listener to shutdown cleanly.
func (a *HttpRPCListener) Stop() {
	a.cancel()
	a.wg.Wait()
}

func (a *HttpRPCListener) handleShutdown() {
	a.wg.Add(1)
	defer a.wg.Done()
	<-a.ctx.Done()
	a.log(nil, "server shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.server.Shutdown(ctx); err != nil {
		a.log(err, "server shutdown failed:")
	}
}

func (a *HttpRPCListener) listenAndServe() {
	a.log(nil, fmt.Sprint("server active on port:", a.Port))
	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		a.log(err)
		a.cancel()
	}
}

func (a *HttpRPCListener) log(err error, msg ...string) {
	if !a.logging {
		return
	}
	if err == nil {
		log.Println(msg)
	} else {
		atomic.AddUint64(&a.errorCount, 1)
		log.Println("[ERROR]", msg, "err:", err)
	}
}
