package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	analyticsH "sales-tracker/internal/http-server/handler/analytics"
	itemsH "sales-tracker/internal/http-server/handler/items"
	"sales-tracker/internal/http-server/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/wb-go/wbf/zlog"
)

func NewRouter(itemsH *itemsH.ItemsHandler, analyticsH *analyticsH.AnalyticsHandler, logger *zlog.Zerolog) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RecoveryMiddleware)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/static/") {
				middleware.LoggingMiddleware(next).ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	})
	workDir, _ := os.Getwd()
	staticDir := http.Dir(filepath.Join(workDir, "static"))
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(staticDir)))
	r.Route("/items", func(r chi.Router) {
		r.Get("/", itemsH.GetItems)
		r.Post("/", itemsH.CreateItem)
		r.Get("/export", itemsH.ExportCSV)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", itemsH.GetItemByID)
			r.Put("/", itemsH.UpdateItem)
			r.Delete("/", itemsH.DeleteItem)
		})
	})
	r.Get("/analytics", analyticsH.GetAnalytics)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		serveHTML(w, r, workDir)
	})
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/static/") &&
			!strings.HasPrefix(r.URL.Path, "/items") &&
			r.URL.Path != "/analytics" {
			serveHTML(w, r, workDir)
		} else {
			http.NotFound(w, r)
		}
	})
	logger.Info().Msg("Routes registered")
	return r
}

func serveHTML(w http.ResponseWriter, r *http.Request, workDir string) {
	indexPath := filepath.Join(workDir, "static", "templates", "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		http.Error(w, "HTML template not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, indexPath)
}
