package repository

import (
	"device-registry-api/internal/domain"
	"testing"
)

func TestMemoryRepositoryCRUD(t *testing.T) {
	r := NewMemoryDeviceRepository()
	d := domain.Device{ID: "dev-001", Name: "Sensor", IPAddress: "127.0.0.1", Type: "sensor", Status: "online"}
	if err := r.Create(d); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.GetByID("dev-001"); !ok {
		t.Fatal("expected device")
	}
	d.Status = "offline"
	if !r.Update("dev-001", d) {
		t.Fatal("update failed")
	}
	if got, _ := r.GetByID("dev-001"); got.Status != "offline" {
		t.Fatal("status not updated")
	}
	if !r.Delete("dev-001") {
		t.Fatal("delete failed")
	}
	if _, ok := r.GetByID("dev-001"); ok {
		t.Fatal("device still exists")
	}
}
