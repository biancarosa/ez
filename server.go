package ez

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ServerOptions contains configuration options for the EZServer
type ServerOptions struct {
	Addr           string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
}

// DefaultServerOptions returns default server options
func DefaultServerOptions() *ServerOptions {
	return &ServerOptions{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
}

type EZServer[T any, U any] struct {
	s        *http.Server
	r        []Route[T, U]
	capacity int
	shutdown chan struct{}
}

// New creates a new EZServer with default options
func New[T any, U any]() *EZServer[T, U] {
	return NewWithOptions[T, U](DefaultServerOptions())
}

// NewWithOptions creates a new EZServer with custom options
func NewWithOptions[T any, U any](opts *ServerOptions) *EZServer[T, U] {
	const defaultRouteCapacity = 10
	server := &http.Server{
		Addr:           opts.Addr,
		ReadTimeout:    opts.ReadTimeout,
		WriteTimeout:   opts.WriteTimeout,
		IdleTimeout:    opts.IdleTimeout,
		MaxHeaderBytes: opts.MaxHeaderBytes,
	}

	return &EZServer[T, U]{
		s:        server,
		r:        make([]Route[T, U], 0, defaultRouteCapacity),
		capacity: defaultRouteCapacity,
		shutdown: make(chan struct{}),
	}
}

// WithCapacity sets the initial capacity for routes
func (ez *EZServer[T, U]) WithCapacity(capacity int) *EZServer[T, U] {
	ez.r = make([]Route[T, U], 0, capacity)
	ez.capacity = capacity
	return ez
}

func (ez *EZServer[T, U]) NotFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func matchesPattern(path, pattern string) bool {
	return path == pattern
}

// Handler returns an http.Handler that processes the route
func (ez *EZServer[T, U]) Handler(route Route[T, U]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !contains(route.Method, r.Method) || !matchesPattern(r.URL.Path, route.Pattern) {
			ez.NotFound(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), RouteKey, route)
		route.Handler(w, r.WithContext(ctx))
	})
}

// RegisterRoute registers a new route. If the slice needs to grow, it will double in capacity.
func (ez *EZServer[T, U]) RegisterRoute(route Route[T, U]) {
	if len(ez.r) == cap(ez.r) {
		// Double capacity when we need to grow
		newCap := cap(ez.r) * 2
		if newCap == 0 {
			newCap = ez.capacity
		}
		newRoutes := make([]Route[T, U], len(ez.r), newCap)
		copy(newRoutes, ez.r)
		ez.r = newRoutes
	}
	ez.r = append(ez.r, route)
	http.Handle(route.Pattern, ez.Handler(route))
}

// RegisterRoutes registers multiple routes at once, preallocating the necessary capacity
func (ez *EZServer[T, U]) RegisterRoutes(routes []Route[T, U]) {
	if needed := len(ez.r) + len(routes); needed > cap(ez.r) {
		newRoutes := make([]Route[T, U], len(ez.r), needed)
		copy(newRoutes, ez.r)
		ez.r = newRoutes
	}
	for _, route := range routes {
		ez.RegisterRoute(route)
	}
}

func (ez *EZServer[T, U]) GetRoutes() []Route[T, U] {
	return ez.r
}

// ListenAndServe starts the server and blocks until it's shut down
func (ez *EZServer[T, U]) ListenAndServe() error {
	fmt.Println("Running server on", ez.s.Addr)

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := ez.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for shutdown signal or error
	select {
	case err := <-serverErr:
		return err
	case <-ez.shutdown:
		fmt.Println("Shutdown signal received")
		return nil
	}
}

// Shutdown gracefully shuts down the server
func (ez *EZServer[T, U]) Shutdown(ctx context.Context) error {
	close(ez.shutdown)
	return ez.s.Shutdown(ctx)
}

func (ez *EZServer[T, U]) GenerateDocs() error {
	generator := DocsGenerator[T, U]{
		server:   ez,
		metadata: DefaultDocMetadata(),
	}
	return generator.GenerateDocs()
}
