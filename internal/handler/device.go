package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"device-registry-api/internal/domain"
	"device-registry-api/internal/repository"
	"github.com/go-chi/chi/v5"
)

var idCounter uint64

type DeviceHandler struct {
	repo repository.DeviceRepository
}

func NewDeviceHandler(repo repository.DeviceRepository) *DeviceHandler {
	return &DeviceHandler{repo: repo}
}

func (h *DeviceHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.Get)
		r.Put("/", h.Update)
		r.Delete("/", h.Delete)
	})
	return r
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeProblem(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, domain.Problem{
		Error: http.StatusText(status), Message: message, Status: status, Timestamp: time.Now().UTC(),
	})
}

func decodeDevice(r *http.Request) (domain.Device, error) {
	var input struct {
		Name      string `json:"name"`
		IPAddress string `json:"ip_address"`
		Type      string `json:"type"`
		Status    string `json:"status"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		return domain.Device{}, fmt.Errorf("invalid JSON body: %w", err)
	}
	input.Name = strings.TrimSpace(input.Name)
	input.IPAddress = strings.TrimSpace(input.IPAddress)
	input.Type = strings.TrimSpace(input.Type)
	input.Status = strings.TrimSpace(input.Status)
	if input.Name == "" || input.IPAddress == "" || input.Type == "" {
		return domain.Device{}, fmt.Errorf("fields name, ip_address and type are required")
	}
	if input.Status == "" {
		input.Status = "offline"
	}
	if input.Status != "online" && input.Status != "offline" {
		return domain.Device{}, fmt.Errorf("status must be online or offline")
	}
	return domain.Device{Name: input.Name, IPAddress: input.IPAddress, Type: input.Type, Status: input.Status}, nil
}

func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	deviceType := strings.TrimSpace(r.URL.Query().Get("type"))
	if status != "" && status != "online" && status != "offline" {
		writeProblem(w, http.StatusBadRequest, "query parameter 'status' must be online or offline")
		return
	}
	writeJSON(w, http.StatusOK, h.repo.GetFiltered(status, deviceType))
}

func (h *DeviceHandler) Create(w http.ResponseWriter, r *http.Request) {
	device, err := decodeDevice(r)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, err.Error())
		return
	}
	device.ID = fmt.Sprintf("dev-%03d", atomic.AddUint64(&idCounter, 1))
	device.UpdatedAt = time.Now().UTC()
	if err := h.repo.Create(device); err != nil {
		writeProblem(w, http.StatusConflict, "device ID already exists")
		return
	}
	writeJSON(w, http.StatusCreated, device)
}

func (h *DeviceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	device, ok := h.repo.GetByID(id)
	if !ok {
		writeProblem(w, http.StatusNotFound, fmt.Sprintf("Device with ID '%s' does not exist in registry", id))
		return
	}
	writeJSON(w, http.StatusOK, device)
}

func (h *DeviceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	device, err := decodeDevice(r)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, err.Error())
		return
	}
	device.ID = id
	device.UpdatedAt = time.Now().UTC()
	if !h.repo.Update(id, device) {
		writeProblem(w, http.StatusNotFound, fmt.Sprintf("Device with ID '%s' does not exist in registry", id))
		return
	}
	writeJSON(w, http.StatusOK, device)
}

func (h *DeviceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !h.repo.Delete(id) {
		writeProblem(w, http.StatusNotFound, fmt.Sprintf("Device with ID '%s' does not exist in registry", id))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func WriteProblemForMiddleware(w http.ResponseWriter, status int, message string) {
	writeProblem(w, status, message)
}
