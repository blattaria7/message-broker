package server

import (
	"context"
	"message-broker/internal/broker"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	broker broker.Broker
}

func NewHandler(broker broker.Broker) *Handler {
	return &Handler{
		broker: broker,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		query := r.URL.Query()
		v, ok := query["v"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
		}

		data := broker.Data{
			Name:  strings.Trim(r.URL.Path, "/"),
			Value: v[0], // тут не делаю проверку количество элементов тк в задании не было про это сказано
		}

		h.broker.Write(data)

		w.WriteHeader(http.StatusOK)
	case http.MethodGet:
		ctx := r.Context()

		name := strings.Trim(r.URL.Path, "/")

		query := r.URL.Query()
		if data, ok := query["timeout"]; ok {
			timeout, err := strconv.ParseInt(data[0], 10, 32)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			ctx, _ = context.WithTimeout(ctx, time.Second*time.Duration(timeout))
		}

		val, err := h.broker.Read(ctx, name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(val))
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method is not supported.", http.StatusBadRequest)
		return
	}
}
