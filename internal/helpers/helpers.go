package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func CheckIDParam(w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	idString := chi.URLParam(r, "id")
	if idString == "" {
		return uuid.Nil, fmt.Errorf("id parameter is required")
	}

	id, err := uuid.Parse(idString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid id parameter:%w", err)
	}

	return id, nil
}

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	js, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		fmt.Println("failed to marshal json:", err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)

	return err
}

func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}
