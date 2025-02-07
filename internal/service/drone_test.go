package service

import (
	"github.com/dronesphere/internal/model/entity"
	"github.com/stretchr/testify/mock"
)

type MockDroneRepo struct {
	mock.Mock
}

func (m *MockDroneRepo) ListAll() ([]entity.Drone, error) {
	args := m.Called()
	return args.Get(0).([]entity.Drone), args.Error(1)
}
