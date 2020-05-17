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

type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

var _ Server = &http.Server{}

type Api struct {
	Port int

	server Server
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	logging    bool
	errorCount uint64
}

func NewApi(ctx context.Context, port int, logging bool) *Api {
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
	api := &Api{
		Port: port,

		server:  server,
		ctx:     ctx,
		cancel:  cancel,
		logging: logging,
	}
	return api
}

var rpcServer RpcServer = rpc.DefaultServer

type RpcServer interface {
	Register(interface{}) error
	HandleHTTP(string, string)
}

func (a *Api) Start(rcvr interface{}) error {
	if err := rpcServer.Register(rcvr); err != nil {
		return err
	}
	rpcServer.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	go a.handleShutdown()
	go a.listenAndServe()
	return nil
}

func (a *Api) Stop() {
	a.cancel()
	a.wg.Wait()
}

func (a *Api) handleShutdown() {
	a.wg.Add(1)
	defer a.wg.Done()
	<-a.ctx.Done()
	a.log(nil, "server shutting down")
	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Second)
	defer cancel()
	if err := a.server.Shutdown(ctx); err != nil {
		a.log(err, "server shutdown failed:")
	}
}

func (a *Api) listenAndServe() {
	a.log(nil, fmt.Sprint("server active on port:", a.Port))
	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		a.log(err)
		a.cancel()
	}
}

var Println = log.Println

func (a *Api) log(err error, msg ...string) {
	if !a.logging {
		return
	}
	if err == nil {
		Println(msg)
	} else {
		atomic.AddUint64(&a.errorCount, 1)
		Println("[ERROR]", msg, "err:", err)
	}
}
