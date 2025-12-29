package observability

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// Provider provides access to all observability components.
type Provider struct {
	config Config
	logger log.Logger

	mu     sync.RWMutex
	closed bool
}

// New creates a new observability provider with the given configuration.
func New(cfg Config) *Provider {
	p := &Provider{
		config: cfg,
	}

	// Initialize logger
	p.logger = p.createLogger(cfg.Log)

	// Set as global logger
	log.SetGlobal(p.logger)

	return p
}

// NewWithDefaults creates a provider with default configuration.
func NewWithDefaults() *Provider {
	return New(DefaultConfig())
}

// NewDevelopment creates a provider configured for development.
func NewDevelopment() *Provider {
	return New(DevelopmentConfig())
}

// NewProduction creates a provider configured for production.
func NewProduction() *Provider {
	return New(ProductionConfig())
}

// createLogger creates a logger based on configuration.
func (p *Provider) createLogger(cfg LogConfig) log.Logger {
	opts := cfg.ToLogOptions()

	switch cfg.Format {
	case "json":
		return log.NewSlogLogger(opts)
	case "console", "":
		return log.NewConsoleLogger(opts)
	default:
		return log.NewConsoleLogger(opts)
	}
}

// ============================================================================
// Accessors
// ============================================================================

// Logger returns the logger instance.
func (p *Provider) Logger() log.Logger {
	return p.logger
}

// ComponentLogger returns a logger with the component field set.
func (p *Provider) ComponentLogger(component string) log.Logger {
	return p.logger.With(log.Component(component))
}

// Config returns the observability configuration.
func (p *Provider) Config() Config {
	return p.config
}

// ============================================================================
// Lifecycle
// ============================================================================

// Start starts all observability components.
func (p *Provider) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.logger.Info("observability provider started",
		log.String("log_level", p.config.Log.Level),
		log.String("log_format", p.config.Log.Format),
	)

	return nil
}

// Stop gracefully stops all observability components.
func (p *Provider) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.logger.Info("observability provider stopping")

	// Flush any buffered logs
	// Future: flush metrics, close trace exporters

	p.closed = true
	return nil
}

// ============================================================================
// Builder Pattern
// ============================================================================

// Builder builds an observability provider with custom options.
type Builder struct {
	config      Config
	output      io.Writer
	errorOutput io.Writer
}

// NewBuilder creates a new provider builder.
func NewBuilder() *Builder {
	return &Builder{
		config:      DefaultConfig(),
		output:      os.Stdout,
		errorOutput: os.Stderr,
	}
}

// WithConfig sets the full configuration.
func (b *Builder) WithConfig(cfg Config) *Builder {
	b.config = cfg
	return b
}

// WithLogLevel sets the log level.
func (b *Builder) WithLogLevel(level string) *Builder {
	b.config.Log.Level = level
	return b
}

// WithLogFormat sets the log format.
func (b *Builder) WithLogFormat(format string) *Builder {
	b.config.Log.Format = format
	return b
}

// WithColorize enables/disables colorized output.
func (b *Builder) WithColorize(colorize bool) *Builder {
	b.config.Log.Colorize = colorize
	return b
}

// WithCaller enables/disables caller information.
func (b *Builder) WithCaller(include bool) *Builder {
	b.config.Log.IncludeCaller = include
	return b
}

// WithComponent sets the default component name.
func (b *Builder) WithComponent(component string) *Builder {
	b.config.Log.Component = component
	return b
}

// WithOutput sets the output writer.
func (b *Builder) WithOutput(w io.Writer) *Builder {
	b.output = w
	return b
}

// WithErrorOutput sets the error output writer.
func (b *Builder) WithErrorOutput(w io.Writer) *Builder {
	b.errorOutput = w
	return b
}

// Build creates the provider.
func (b *Builder) Build() *Provider {
	p := &Provider{
		config: b.config,
	}

	// Create logger with custom outputs
	opts := b.config.Log.ToLogOptions().
		WithOutput(b.output).
		WithErrorOutput(b.errorOutput)

	switch b.config.Log.Format {
	case "json":
		p.logger = log.NewSlogLogger(opts)
	default:
		p.logger = log.NewConsoleLogger(opts)
	}

	log.SetGlobal(p.logger)

	return p
}

// ============================================================================
// Global Provider
// ============================================================================

var (
	globalProvider     *Provider
	globalProviderOnce sync.Once
)

// Global returns the global observability provider.
// Initializes with defaults if not already set.
func Global() *Provider {
	globalProviderOnce.Do(func() {
		if globalProvider == nil {
			globalProvider = NewWithDefaults()
		}
	})
	return globalProvider
}

// SetGlobal sets the global observability provider.
func SetGlobal(p *Provider) {
	globalProvider = p
	globalProviderOnce.Do(func() {}) // Mark as initialized
}

// L is a shortcut for Global().Logger().
func L() log.Logger {
	return Global().Logger()
}
