package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnection_Creation(t *testing.T) {
	tests := []struct {
		name string
		conn Connection
		want bool
	}{
		{
			name: "valid database connection",
			conn: Connection{
				ID:      "conn-1",
				Name:    "test-db",
				Type:    ConnectionTypeDatabase,
				Subtype: "postgresql",
				Config: ConnectionConfig{
					Host:     "localhost",
					Port:     5432,
					Username: "user",
					Database: "testdb",
				},
				Status: ConnectionStatusActive,
			},
			want: true,
		},
		{
			name: "valid kubernetes connection",
			conn: Connection{
				ID:      "conn-2",
				Name:    "test-k8s",
				Type:    ConnectionTypeKubernetes,
				Subtype: "minikube",
				Config: ConnectionConfig{
					Extra: map[string]string{
						"kubeconfig": "/path/to/config",
					},
				},
				Status: ConnectionStatusActive,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.conn.ID)
			assert.NotEmpty(t, tt.conn.Name)
			assert.NotEmpty(t, tt.conn.Type)
			assert.NotEmpty(t, tt.conn.Status)
		})
	}
}

func TestConnection_Validation(t *testing.T) {
	tests := []struct {
		name    string
		conn    Connection
		wantErr bool
	}{
		{
			name: "empty name should fail",
			conn: Connection{
				ID:   "conn-1",
				Name: "",
				Type: ConnectionTypeDatabase,
			},
			wantErr: true,
		},
		{
			name: "invalid connection type should fail",
			conn: Connection{
				ID:   "conn-1",
				Name: "test",
				Type: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid connection should pass",
			conn: Connection{
				ID:   "conn-1",
				Name: "test",
				Type: ConnectionTypeDatabase,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConnection(tt.conn)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConnection_LastUsedUpdate(t *testing.T) {
	conn := Connection{
		ID:   "conn-1",
		Name: "test",
		Type: ConnectionTypeDatabase,
	}

	assert.Nil(t, conn.LastUsed)

	now := time.Now()
	conn.LastUsed = &now

	assert.NotNil(t, conn.LastUsed)
	assert.Equal(t, now, *conn.LastUsed)
}

// Helper function for validation (would be in the actual domain logic)
func validateConnection(conn Connection) error {
	if conn.Name == "" {
		return assert.AnError
	}
	if conn.Type != ConnectionTypeDatabase &&
		conn.Type != ConnectionTypeKubernetes &&
		conn.Type != ConnectionTypeCloud {
		return assert.AnError
	}
	return nil
}
