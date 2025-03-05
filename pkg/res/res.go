package res

import (
	"encoding/json"
	"net/http"
)

func Json(w http.ResponseWriter, data any, statuscode int) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statuscode)
	json.NewEncoder(w).Encode(data)
}
