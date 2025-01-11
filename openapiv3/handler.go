package openapiv3

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"

	"github.com/go-kratos/kratos/v2/api/metadata"
	"github.com/go-kratos/kratos/v2/transport/http/binding"
)

func NewHandler(opts ...Option) http.Handler {
	service := New(opts...)
	r := mux.NewRouter()

	r.HandleFunc("/q/services", func(w http.ResponseWriter, r *http.Request) {
		services, err := service.ListServices(r.Context())
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(services)
	}).Methods("GET")

	r.HandleFunc("/q/service/group/{name}", func(w http.ResponseWriter, r *http.Request) {
		raws := mux.Vars(r)
		vars := make(url.Values, len(raws))
		for k, v := range raws {
			vars[k] = []string{v}
		}
		var in metadata.GetServiceDescRequest
		if err := binding.BindQuery(vars, &in); err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}

		content, err := service.GetServiceGroupOpenAPI(r.Context(), in.Name)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(200)
		w.Write([]byte(content))
	}).Methods("GET")

	r.HandleFunc("/q/service/{name}", func(w http.ResponseWriter, r *http.Request) {
		raws := mux.Vars(r)
		vars := make(url.Values, len(raws))
		for k, v := range raws {
			vars[k] = []string{v}
		}
		var in metadata.GetServiceDescRequest
		if err := binding.BindQuery(vars, &in); err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}

		content, err := service.GetServiceOpenAPI(r.Context(), in.Name)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(200)
		w.Write([]byte(content))
	}).Methods("GET")

	return r
}
