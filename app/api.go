package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	_ "net/http/pprof"

	"github.com/gorilla/mux"
)

type Api struct {
	Port int

	server *http.Server
	router *mux.Router
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errorCount uint64
}

func NewApi(ctx context.Context, port int) *Api {
	router := mux.NewRouter()
	server := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%d", port),
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
		router: router,
		ctx:    ctx,
		cancel: cancel,
	}
	return api
}

func (a *Api) Start() {
	a.router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	if list, err := a.ActiveRoutes(a.router); err == nil {
		log.Println("active routes:", list)
	}

	go a.handleShutdown()
	go a.listenAndServe()
}

func (a *Api) Stop() {
	a.cancel()
}

func (a *Api) AddRoute(path string, handler http.HandlerFunc) {
	a.router.HandleFunc(path, handler)
}

func (a *Api) ActiveRoutes(r *mux.Router) ([]string, error) {
	routes := make([]string, 0)

	fn := func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		if route.GetHandler() != nil { // only list routes with a handler
			tmpl, _ := route.GetPathTemplate()
			routes = append(routes, tmpl)
		}
		return nil
	}

	err := r.Walk(fn)
	return routes, err

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
