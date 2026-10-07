package repository

import (
	"errors"
	"sync"

	"device-registry-api/internal/domain"
)

var ErrConflict = errors.New("device already exists")

type DeviceRepository interface {
	GetAll() []domain.Device
	GetFiltered(status, deviceType string) []domain.Device
	GetByID(id string) (domain.Device, bool)
	Create(device domain.Device) error
	Update(id string, device domain.Device) bool
	Delete(id string) bool
}

type MemoryDeviceRepository struct {
	mu      sync.RWMutex
	storage map[string]domain.Device
}

func NewMemoryDeviceRepository() *MemoryDeviceRepository {
	return &MemoryDeviceRepository{storage: make(map[string]domain.Device)}
}

func (r *MemoryDeviceRepository) GetAll() []domain.Device {
	return r.GetFiltered("", "")
}

func (r *MemoryDeviceRepository) GetFiltered(status, deviceType string) []domain.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()

	devices := make([]domain.Device, 0, len(r.storage))
	for _, d := range r.storage {
		if status != "" && d.Status != status {
			continue
		}
		if deviceType != "" && d.Type != deviceType {
			continue
		}
		devices = append(devices, d)
	}
	return devices
}

func (r *MemoryDeviceRepository) GetByID(id string) (domain.Device, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.storage[id]
	return d, ok
}

func (r *MemoryDeviceRepository) Create(device domain.Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.storage[device.ID]; exists {
		return ErrConflict
	}
	r.storage[device.ID] = device
	return nil
}

func (r *MemoryDeviceRepository) Update(id string, device domain.Device) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.storage[id]; !exists {
		return false
	}
	device.ID = id
	r.storage[id] = device
	return true
}

func (r *MemoryDeviceRepository) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.storage[id]; !exists {
		return false
	}
	delete(r.storage, id)
	return true
}
