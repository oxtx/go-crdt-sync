// internal/api/handlers.go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/oxtx/go-crdt-sync/internal/service"
)

type Service interface {
	PutDoc(docID string, req service.PutDocRequest) (service.PutDocResponse, error)
	PostOps(docID string, req service.PostOpsRequest) (service.PostOpsResponse, error)
	GetDoc(docID string) (service.GetDocResponse, error)
}

func Attach(r *chi.Mux, svc Service) {
	r.Route("/v1", func(r chi.Router) {
		r.Put("/docs/{docID}", func(w http.ResponseWriter, req *http.Request) {
			id := chi.URLParam(req, "docID")
			var body service.PutDocRequest
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				render.Status(req, http.StatusBadRequest)
				render.PlainText(w, req, "invalid json")
				return
			}
			res, err := svc.PutDoc(id, body)
			if err != nil {
				render.Status(req, http.StatusBadRequest)
				render.PlainText(w, req, err.Error())
				return
			}
			render.JSON(w, req, res)
		})
		r.Get("/docs/{docID}", func(w http.ResponseWriter, req *http.Request) {
			id := chi.URLParam(req, "docID")
			res, err := svc.GetDoc(id)
			if err != nil {
				render.Status(req, http.StatusNotFound)
				render.PlainText(w, req, err.Error())
				return
			}
			render.JSON(w, req, res)
		})
		r.Post("/docs/{docID}/ops", func(w http.ResponseWriter, req *http.Request) {
			id := chi.URLParam(req, "docID")
			var body service.PostOpsRequest
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				render.Status(req, http.StatusBadRequest)
				render.PlainText(w, req, "invalid json")
				return
			}
			res, err := svc.PostOps(id, body)
			if err != nil {
				render.Status(req, http.StatusBadRequest)
				render.PlainText(w, req, err.Error())
				return
			}
			render.JSON(w, req, res)
		})
	})
}
