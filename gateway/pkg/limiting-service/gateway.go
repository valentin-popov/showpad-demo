package limiter

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gateway/pkg/config"
	errorlog "gateway/pkg/error-log"
	"gateway/pkg/strategy"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Limiter represents a rate limiter structure.
type Limiter struct {
	address string

	logger      *errorlog.Logger
	routeLimits map[string]strategy.LimitStrategy
	sqlDb       *sql.DB
	userIdCache *UserCache

	apiAddress string
	apiKey     string
}

// New creates a new Limiter instance with the provided configuration.
func New(ctx context.Context, cfg *config.Config) (*Limiter, error) {
	logger, err := errorlog.New(cfg.LogFile)
	if err != nil {
		return nil, err
	}

	db, err := NewDB(ctx, cfg.DBFile)
	if err != nil {
		return nil, err
	}

	lim := &Limiter{
		address: cfg.Address,

		sqlDb:  db,
		logger: logger,
		userIdCache: &UserCache{
			data: make(map[string]userData),
			ttl:  cfg.UserCacheTTL * time.Minute,
		},
		apiAddress: cfg.Api.Address,
		apiKey:     cfg.Api.Key,
	}

	routeLimits := map[string]strategy.LimitStrategy{}

	for path, route := range cfg.Routes {

		switch route.Strategy {
		case "token_bucket":
			routeLimits[path] = &strategy.TokenBucket{
				Capacity:      route.BucketCap,
				Created:       time.Now(),
				Mu:            sync.Mutex{},
				LastRefill:    map[string]map[string]time.Time{},
				CurrentTokens: map[string]map[string]int{},
			}

		case "fixed_window":
			routeLimits[path] = &strategy.FixedWindow{
				LengthSeconds: route.WindowLength,
				SqlDb:         db,
				Logger:        logger,
				SqlTable:      route.SqlTable,
			}
		}
	}

	if len(routeLimits) != 0 {
		lim.routeLimits = routeLimits
	}

	return lim, nil

}

// Run starts the HTTP server and listens for incoming requests.
func (l *Limiter) Run(ctx context.Context) error {
	srv := http.Server{
		Addr:    l.address,
		Handler: l,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	fmt.Println("API gateway running on " + l.address)

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return srv.Shutdown(shutdownCtx)

}

// ServeHTTP is the main handler that processes incoming HTTP requests and applies rate limiting.
func (l *Limiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		l.logger.WriteError(errUnauthorized)
		http.Error(w, respUnauthorized, http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		l.logger.WriteError(errUnauthorized)
		http.Error(w, respUnauthorized, http.StatusUnauthorized)
		return
	}

	userId := strings.TrimPrefix(authHeader, "Bearer ")
	if userId == "" {
		l.logger.WriteError(errUnauthorized)
		http.Error(w, respUnauthorized, http.StatusUnauthorized)
		return
	}

	if !l.isValidUser(userId) {
		l.logger.WriteError(errUnauthorized)
		http.Error(w, respUnauthorized, http.StatusUnauthorized)
		return
	}

	// user quota update
	if r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/users/") {

		if !isAdmin(userId) {
			http.Error(w, respUnauthorized, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			http.Error(w, respBadRequest, http.StatusBadRequest)
			return
		}
		victimId := parts[2]

		byteSlc, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, respBadRequest, http.StatusBadRequest)
			return
		}

		data := struct {
			Rate float64 `json:"rate"`
		}{}

		if err := json.Unmarshal(byteSlc, &data); err != nil {
			http.Error(w, respBadRequest, http.StatusBadRequest)
			return
		}

		if err := l.updateUserQuota(victimId, data.Rate); err != nil {
			http.Error(w, respInternalServer, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "{userId: %s, rate: %.3f}", victimId, data.Rate)
		return
	}

	algo := l.routeLimits[r.URL.Path]

	if algo == nil {
		l.logger.WriteError(errNotFound)
		http.Error(w, respNotFound, http.StatusNotFound)
		return
	}

	if !algo.Accept(userId, l.userIdCache.GetRate(userId), r.URL.Path) {
		l.logger.WriteError(errRateLimitExceeded)
		http.Error(w, respRateLimitExceeded, http.StatusTooManyRequests)
		return
	}

	l.sendToAPI(w, r)

}

// Stop performs any necessary cleanup for the Limiter.
func (l *Limiter) Stop() {
	l.logger.WriteInfo("Shutting down limiter...")
	l.logger.Close()
	l.sqlDb.Close()
}

func (l *Limiter) sendToAPI(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(r.Method, l.apiAddress, r.Body)
	if err != nil {
		l.logger.WriteError(fmt.Errorf("internal server error: %w", err))
		http.Error(w, respInternalServer, http.StatusInternalServerError)
		return
	}

	client := &http.Client{}

	gatewayToken := r.Header.Get("Authorization")

	apiToken := gatewayToken + ":" + l.apiKey

	req.Header.Add("Authorization", apiToken)

	resp, err := client.Do(req)
	if err != nil {
		l.logger.WriteError(fmt.Errorf("internal server error: %w", err))
		http.Error(w, respInternalServer, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for headerName, headerValues := range resp.Header {
		for _, headerValue := range headerValues {
			w.Header().Add(headerName, headerValue)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// First, the user is looked up in the cache.
// If not found, the user is looked up in the persistent database.
// If found in the database, the user is added to the cache for future requests.
func (l *Limiter) isValidUser(userId string) bool {
	quota := l.userIdCache.GetRate(userId)
	if quota > 0 {
		return true
	}

	if err := l.sqlDb.QueryRow("SELECT quota FROM users WHERE id = ?", userId).Scan(&quota); err != nil {
		if err != sql.ErrNoRows {
			l.logger.WriteError(fmt.Errorf("database error: %w", err))
		}
		return false
	}

	l.userIdCache.Add(userId, quota)

	return true
}

func isAdmin(userId string) bool {
	return userId == "0"
}
