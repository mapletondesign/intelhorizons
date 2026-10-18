package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"intelhorizons/internal/cache"
	"intelhorizons/internal/mailer"
	"intelhorizons/internal/store"
)

type Handler struct {
	store   *store.Store
	cache   *cache.Cache
	mailer  *mailer.Mailer
	siteURL string
}

func New(st *store.Store, ca *cache.Cache, ml *mailer.Mailer, siteURL string) *Handler {
	return &Handler{store: st, cache: ca, mailer: ml, siteURL: siteURL}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(corsMiddleware(envOr("CORS_ORIGIN", "https://intelhorizons.com")))
	r.Use(securityHeaders)

	r.Get("/api/feed", h.feed)
	r.Get("/api/digest", h.digest)
	r.Get("/api/sources", h.sources)
	r.Get("/healthz", h.healthz)
	r.Post("/api/subscribe", h.subscribe)
	r.Get("/api/subscribe/confirm", h.confirmSubscription)
	r.Get("/api/unsubscribe", h.unsubscribe)

	r.Handle("/*", spaHandler("web/dist"))

	return r
}

func (h *Handler) feed(w http.ResponseWriter, r *http.Request) {
	q := store.FeedQuery{
		Limit:    queryInt(r, "limit", 20),
		Offset:   queryInt(r, "offset", 0),
		SourceID: queryInt(r, "source", 0),
		Search:   r.URL.Query().Get("q"),
		Category: r.URL.Query().Get("category"),
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	cacheKey := h.cache.FeedKey(q.SourceID, q.Limit, q.Offset, q.Search, q.Category)
	var articles []store.Article
	if hit, err := h.cache.Get(r.Context(), cacheKey, &articles); err != nil {
		log.Printf("[%s] cache get feed: %v", chimw.GetReqID(r.Context()), err)
	} else if hit {
		writeJSON(w, http.StatusOK, articles)
		return
	}

	articles, err := h.store.Feed(r.Context(), q)
	if err != nil {
		log.Printf("[%s] store.Feed: %v", chimw.GetReqID(r.Context()), err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if articles == nil {
		articles = []store.Article{}
	}

	if err := h.cache.Set(r.Context(), cacheKey, articles, h.cache.FeedTTL()); err != nil {
		log.Printf("[%s] cache set feed: %v", chimw.GetReqID(r.Context()), err)
	}

	writeJSON(w, http.StatusOK, articles)
}

func (h *Handler) digest(w http.ResponseWriter, r *http.Request) {
	key := h.cache.DigestKey()
	var articles []store.Article
	if hit, err := h.cache.Get(r.Context(), key, &articles); err != nil {
		log.Printf("[%s] cache get digest: %v", chimw.GetReqID(r.Context()), err)
	} else if hit {
		writeJSON(w, http.StatusOK, articles)
		return
	}

	articles, err := h.store.DailyDigest(r.Context(), 20)
	if err != nil {
		log.Printf("[%s] store.DailyDigest: %v", chimw.GetReqID(r.Context()), err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := h.cache.Set(r.Context(), key, articles, h.cache.DigestTTL()); err != nil {
		log.Printf("[%s] cache set digest: %v", chimw.GetReqID(r.Context()), err)
	}

	writeJSON(w, http.StatusOK, articles)
}

func (h *Handler) sources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.store.AllSources(r.Context())
	if err != nil {
		log.Printf("[%s] store.AllSources: %v", chimw.GetReqID(r.Context()), err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if sources == nil {
		sources = []store.Source{}
	}
	writeJSON(w, http.StatusOK, sources)
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	dbOk := h.store.Ping(r.Context()) == nil
	redisOk := h.cache.Ping(r.Context()) == nil

	status := "ok"
	code := http.StatusOK
	if !dbOk || !redisOk {
		status = "degraded"
		code = http.StatusServiceUnavailable
	}

	writeJSON(w, code, map[string]any{
		"status": status,
		"db":     dbOk,
		"redis":  redisOk,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func corsMiddleware(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowedOrigin != "" && origin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

func queryInt(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// spaHandler serves static files and falls back to index.html for unknown
// paths so that the Vue router can handle client-side navigation.
func spaHandler(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	})
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
