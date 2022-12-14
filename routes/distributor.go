package routes

import (
	"net/http"
)

func (hw *HandleWrapper) Process2D(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:
		hw.Process2DGet(w, r)

	case http.MethodPost:
		hw.Process2DPost(w, r)

	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed) //TODO custom error message
	}
}

func (hw *HandleWrapper) Process3D(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:
		hw.Process3DGet(w, r)

	case http.MethodPost:
		hw.Process3DPost(w, r)

	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed) //TODO custom error message
	}
}
