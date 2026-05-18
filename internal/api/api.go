package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"intelhorizons/internal/cache"
	"intelhorizons/internal/store"
)

type Handler struct {
	store *store.Store
	cache *cache.Cache
}

func New(st *store.Store, ca *cache.Cache) *Handler {
	return &Handler{store: st, cache: ca}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Get("/api/feed", h.feed)
	r.Get("/api/digest", h.digest)
	r.Get("/api/sources", h.sources)
	r.Get("/healthz", h.healthz)

	// Serve Vue frontend from embedded or local dist
	r.Handle("/*", http.FileServer(http.Dir("web/dist")))

	return r
}

func (h *Handler) feed(w http.ResponseWriter, r *http.Request) {
	q := store.FeedQuery{
		Limit:    queryInt(r, "limit", 20),
		Offset:   queryInt(r, "offset", 0),
		SourceID: queryInt(r, "source", 0),
		Search:   r.URL.Query().Get("q"),
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	cacheKey := h.cache.FeedKey(q.SourceID, q.Limit, q.Offset, q.Search)
	var articles []store.Article
	if hit, err := h.cache.Get(r.Context(), cacheKey, &articles); err != nil {
		log.Printf("cache get: %v", err)
	} else if hit {
		writeJSON(w, articles)
		return
	}

	articles, err := h.store.Feed(r.Context(), q)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		log.Printf("store.Feed: %v", err)
		return
	}

	if err := h.cache.Set(r.Context(), cacheKey, articles, h.cache.FeedTTL()); err != nil {
		log.Printf("cache set: %v", err)
	}

	writeJSON(w, articles)
}

func (h *Handler) digest(w http.ResponseWriter, r *http.Request) {
	key := h.cache.DigestKey()
	var articles []store.Article
	if hit, err := h.cache.Get(r.Context(), key, &articles); err != nil {
		log.Printf("cache get: %v", err)
	} else if hit {
		writeJSON(w, articles)
		return
	}

	articles, err := h.store.DailyDigest(r.Context(), 20)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		log.Printf("store.DailyDigest: %v", err)
		return
	}

	if err := h.cache.Set(r.Context(), key, articles, h.cache.DigestTTL()); err != nil {
		log.Printf("cache set: %v", err)
	}

	writeJSON(w, articles)
}

func (h *Handler) sources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.store.AllSources(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		log.Printf("store.AllSources: %v", err)
		return
	}
	writeJSON(w, sources)
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json encode: %v", err)
	}
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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://intelhorizons.com")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
