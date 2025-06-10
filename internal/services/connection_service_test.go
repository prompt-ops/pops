package services

import (
	"context"
	"testing"

	"github.com/prompt-ops/pops/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockConnectionRepository struct {
	mock.Mock
}

func (m *MockConnectionRepository) Save(ctx context.Context, conn *domain.Connection) error {
	args := m.Called(ctx, conn)
	return args.Error(0)
}

func (m *MockConnectionRepository) FindByName(ctx context.Context, name string) (*domain.Connection, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*domain.Connection), args.Error(1)
}

func (m *MockConnectionRepository) FindAll(ctx context.Context) ([]*domain.Connection, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Connection), args.Error(1)
}

func (m *MockConnectionRepository) Delete(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

type MockConnectionAdapter struct {
	mock.Mock
}

func (m *MockConnectionAdapter) TestConnection(ctx context.Context, conn *domain.Connection) error {
	args := m.Called(ctx, conn)
	return args.Error(0)
}

func (m *MockConnectionAdapter) FetchContext(ctx context.Context, conn *domain.Connection) (string, error) {
	args := m.Called(ctx, conn)
	return args.String(0), args.Error(1)
}

func (m *MockConnectionAdapter) ExecuteCommand(ctx context.Context, conn *domain.Connection, command string) (string, error) {
	args := m.Called(ctx, conn, command)
	return args.String(0), args.Error(1)
}

func TestConnectionService_CreateConnection(t *testing.T) {
	tests := []struct {
		name          string
		request       CreateConnectionRequest
		setupMocks    func(*MockConnectionRepository, *MockConnectionAdapter)
		expectError   bool
		expectedCalls int
	}{
		{
			name: "successful connection creation",
			request: CreateConnectionRequest{
				Name:     "test-db",
				Type:     domain.ConnectionTypeDatabase,
				Subtype:  "postgresql",
				Host:     "localhost",
				Port:     5432,
				Username: "user",
				Database: "testdb",
			},
			setupMocks: func(repo *MockConnectionRepository, adapter *MockConnectionAdapter) {
				adapter.On("TestConnection", mock.Anything, mock.Anything).Return(nil)
				adapter.On("FetchContext", mock.Anything, mock.Anything).Return("Tables: users, orders", nil)
				repo.On("Save", mock.Anything, mock.Anything).Return(nil)
			},
			expectError:   false,
			expectedCalls: 1,
		},
		{
			name: "connection test failure",
			request: CreateConnectionRequest{
				Name:     "test-db",
				Type:     domain.ConnectionTypeDatabase,
				Subtype:  "postgresql",
				Host:     "invalid-host",
				Port:     5432,
				Username: "user",
				Database: "testdb",
			},
			setupMocks: func(repo *MockConnectionRepository, adapter *MockConnectionAdapter) {
				adapter.On("TestConnection", mock.Anything, mock.Anything).Return(assert.AnError)
				repo.On("Save", mock.Anything, mock.Anything).Return(nil)
			},
			expectError:   true,
			expectedCalls: 1, // Save is called to store error status
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockConnectionRepository)
			adapter := new(MockConnectionAdapter)

			tt.setupMocks(repo, adapter)

			service := NewConnectionService(repo, adapter)

			ctx := context.Background()
			conn, err := service.CreateConnection(ctx, tt.request)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, conn)
				assert.Equal(t, tt.request.Name, conn.Name)
				assert.Equal(t, tt.request.Type, conn.Type)
				assert.Equal(t, tt.request.Subtype, conn.Subtype)
			}

			repo.AssertNumberOfCalls(t, "Save", tt.expectedCalls)
			adapter.AssertExpectations(t)
		})
	}
}

func TestConnectionService_GetConnection(t *testing.T) {
	tests := []struct {
		name           string
		connectionName string
		setupMocks     func(*MockConnectionRepository)
		expectError    bool
	}{
		{
			name:           "existing connection",
			connectionName: "test-db",
			setupMocks: func(repo *MockConnectionRepository) {
				conn := &domain.Connection{
					ID:       "conn-1",
					Name:     "test-db",
					Type:     domain.ConnectionTypeDatabase,
					Status:   domain.ConnectionStatusActive,
					LastUsed: nil,
				}
				repo.On("FindByName", mock.Anything, "test-db").Return(conn, nil)
				repo.On("Save", mock.Anything, mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name:           "non-existing connection",
			connectionName: "missing-db",
			setupMocks: func(repo *MockConnectionRepository) {
				repo.On("FindByName", mock.Anything, "missing-db").Return((*domain.Connection)(nil), assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockConnectionRepository)
			adapter := new(MockConnectionAdapter)

			tt.setupMocks(repo)

			service := NewConnectionService(repo, adapter)

			ctx := context.Background()
			conn, err := service.GetConnection(ctx, tt.connectionName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, conn)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, conn)
				assert.Equal(t, tt.connectionName, conn.Name)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestConnectionService_ListConnections(t *testing.T) {
	repo := new(MockConnectionRepository)
	adapter := new(MockConnectionAdapter)

	connections := []*domain.Connection{
		{
			ID:   "conn-1",
			Name: "db1",
			Type: domain.ConnectionTypeDatabase,
		},
		{
			ID:   "conn-2",
			Name: "k8s1",
			Type: domain.ConnectionTypeKubernetes,
		},
	}

	repo.On("FindAll", mock.Anything).Return(connections, nil)

	service := NewConnectionService(repo, adapter)

	ctx := context.Background()
	result, err := service.ListConnections(ctx, ConnectionFilter{})

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "db1", result[0].Name)
	assert.Equal(t, "k8s1", result[1].Name)

	repo.AssertExpectations(t)
}

func TestConnectionService_DeleteConnection(t *testing.T) {
	repo := new(MockConnectionRepository)
	adapter := new(MockConnectionAdapter)

	repo.On("Delete", mock.Anything, "test-db").Return(nil)

	service := NewConnectionService(repo, adapter)

	ctx := context.Background()
	err := service.DeleteConnection(ctx, "test-db")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
