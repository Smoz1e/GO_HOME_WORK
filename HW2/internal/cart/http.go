package cart

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func Handler(service *Service, logger *log.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /user/{userID}/cart/{skuID}", func(w http.ResponseWriter, r *http.Request) {
		userID, err := positiveID(r.PathValue("userID"))
		if err != nil {
			writeError(w, http.StatusBadRequest, ErrInvalidUser.Error())
			return
		}
		sku, err := positiveID(r.PathValue("skuID"))
		if err != nil {
			writeError(w, http.StatusBadRequest, ErrInvalidSKU.Error())
			return
		}
		var input struct {
			Count *uint32 `json:"count"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "некорректное тело запроса: укажите положительное count")
			return
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			writeError(w, http.StatusBadRequest, "тело запроса должно содержать один объект JSON")
			return
		}
		if input.Count == nil || *input.Count == 0 {
			writeError(w, http.StatusBadRequest, ErrInvalidCount.Error())
			return
		}
		if err := service.Add(r.Context(), userID, sku, *input.Count); err != nil {
			switch {
			case errors.Is(err, ErrInvalidUser), errors.Is(err, ErrInvalidSKU), errors.Is(err, ErrInvalidCount), errors.Is(err, ErrOverflow):
				writeError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, ErrProductNotFound):
				writeError(w, http.StatusPreconditionFailed, err.Error())
			default:
				writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервиса корзины")
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /user/{userID}/cart", func(w http.ResponseWriter, r *http.Request) {
		userID, err := positiveID(r.PathValue("userID"))
		if err != nil {
			writeError(w, http.StatusBadRequest, ErrInvalidUser.Error())
			return
		}
		view, err := service.Get(userID)
		if err != nil {
			if errors.Is(err, ErrCartNotFound) {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			if errors.Is(err, ErrInvalidUser) {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервиса корзины")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(view); err != nil && logger != nil {
			logger.Printf("не удалось вывести корзину: %v", err)
		}
	})
	mux.HandleFunc("GET /swagger/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		_, _ = io.WriteString(w, openAPISpec)
	})
	mux.HandleFunc("GET /swagger/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, swaggerHTML)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "маршрут не найден")
	})
	return Logging(mux, logger)
}

func positiveID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("разобрать ID: %w", err)
	}
	if id < 1 {
		return 0, errors.New("ID должен быть положительным")
	}
	return id, nil
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"ошибка": message})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) { w.status = code; w.ResponseWriter.WriteHeader(code) }

func Logging(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		if logger != nil {
			logger.Printf("%s %s %d %s", r.Method, r.URL.Path, wrapped.status, time.Since(start))
		}
	})
}
