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

type Api struct {
	Port int

	server *http.Server
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errorCount uint64
}

func NewApi(ctx context.Context, port int) *Api {
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

		server: server,
		ctx:    ctx,
		cancel: cancel,
	}
	return api
}

func (a *Api) Start(rcvr interface{}) error {
	if err := rpc.Register(rcvr); err != nil {
		return err
	}
	rpc.HandleHTTP()
	go a.handleShutdown()
	go a.listenAndServe()
	return nil
}

func (a *Api) Stop() {
	a.cancel()
}

func (a *Api) handleShutdown() {
	a.wg.Add(1)
	defer a.wg.Done()
	<-a.ctx.Done()
	log.Println("server shutting down")
	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Second)
	defer cancel()
	if err := a.server.Shutdown(ctx); err != nil {
		a.logError(err, "server shutdown failed:")
	}
}

func (a *Api) listenAndServe() {
	log.Println("server active on:", a.server.Addr)
	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		a.logError(err)
		a.cancel()
	}
}

func (a *Api) logError(err error, msg ...string) {
	atomic.AddUint64(&a.errorCount, 1)
	log.Println("[ERROR]", msg, "err:", err)
}
