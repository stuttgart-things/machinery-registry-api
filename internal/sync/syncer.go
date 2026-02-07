package sync

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/stuttgart-things/machinery-registry-api/internal/registry"
)

// Config holds syncer configuration.
type Config struct {
	Repo     string        // GitHub repo slug, e.g. "stuttgart-things/harvester"
	Path     string        // Path to registry file in repo
	Branch   string        // Git branch
	Token    string        // GitHub token (optional, for private repos)
	Interval time.Duration // Polling interval
	BaseURL  string        // Override base URL (for testing); defaults to https://raw.githubusercontent.com
}

// Syncer periodically fetches registry.yaml from GitHub and maintains
// a thread-safe in-memory snapshot.
type Syncer struct {
	cfg      Config
	registry *registry.ClaimRegistry
	mu       sync.RWMutex
	client   *http.Client
	cancel   context.CancelFunc
	done     chan struct{}
}

// NewSyncer creates a new Syncer with the given configuration.
func NewSyncer(cfg Config) *Syncer {
	if cfg.Path == "" {
		cfg.Path = "claims/registry.yaml"
	}
	if cfg.Branch == "" {
		cfg.Branch = "main"
	}
	if cfg.Interval == 0 {
		cfg.Interval = 60 * time.Second
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://raw.githubusercontent.com"
	}
	return &Syncer{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
		done:   make(chan struct{}),
	}
}

// rawURL returns the GitHub raw content URL for the registry file.
func (s *Syncer) rawURL() string {
	return fmt.Sprintf("%s/%s/%s/%s", s.cfg.BaseURL, s.cfg.Repo, s.cfg.Branch, s.cfg.Path)
}

// fetch downloads and parses the registry file.
func (s *Syncer) fetch(ctx context.Context) (*registry.ClaimRegistry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.rawURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if s.cfg.Token != "" {
		req.Header.Set("Authorization", "token "+s.cfg.Token)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, s.rawURL())
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return registry.ParseData(data)
}

// InitialSync performs the first sync. Returns an error if the fetch fails
// (fail-fast on startup).
func (s *Syncer) InitialSync(ctx context.Context) error {
	reg, err := s.fetch(ctx)
	if err != nil {
		return fmt.Errorf("initial sync failed: %w", err)
	}

	s.mu.Lock()
	s.registry = reg
	s.mu.Unlock()

	log.Printf("Initial sync complete: %d claims loaded from %s", len(reg.Claims), s.rawURL())
	return nil
}

// Start begins the background polling loop. Call Stop() or cancel the
// context to terminate.
func (s *Syncer) Start(ctx context.Context) {
	ctx, s.cancel = context.WithCancel(ctx)

	go func() {
		defer close(s.done)
		ticker := time.NewTicker(s.cfg.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				reg, err := s.fetch(ctx)
				if err != nil {
					log.Printf("Sync error: %v", err)
					continue
				}
				s.mu.Lock()
				s.registry = reg
				s.mu.Unlock()
				log.Printf("Sync complete: %d claims", len(reg.Claims))
			}
		}
	}()
}

// Stop terminates the background polling loop and waits for it to finish.
func (s *Syncer) Stop() {
	if s.cancel != nil {
		s.cancel()
		<-s.done
	}
}

// GetRegistry returns the current registry snapshot (thread-safe).
func (s *Syncer) GetRegistry() *registry.ClaimRegistry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.registry
}
