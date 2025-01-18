package conn

import (
	"reflect"
	"testing"
)

func TestNewDatabaseConnection(t *testing.T) {
	type args struct {
		name                        string
		availableDatabaseConnection AvailableDatabaseConnectionType
		connectionString            string
	}
	tests := []struct {
		name string
		args args
		want Connection
	}{
		{
			name: "Test NewDatabaseConnection",
			args: args{
				name:                        "test",
				availableDatabaseConnection: PostgreSQLDatabaseConnection,
				connectionString:            "host=localhost user=test password=test dbname=test",
			},
			want: Connection{
				Name: "test",
				Type: DatabaseConnectionType{
					MainType: ConnectionTypeDatabase,
					Subtype:  "PostgreSQL",
				},
				Details: DatabaseConnectionDetails{
					ConnectionString: "host=localhost user=test password=test dbname=test",
					Driver:           "postgres",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDatabaseConnection(tt.args.name, tt.args.availableDatabaseConnection, tt.args.connectionString); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDatabaseConnection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDatabaseConnectionDetails(t *testing.T) {
	type args struct {
		conn Connection
	}
	tests := []struct {
		name    string
		args    args
		want    DatabaseConnectionDetails
		wantErr bool
	}{
		{
			name: "Test GetDatabaseConnectionDetails",
			args: args{
				conn: Connection{
					Name: "test",
					Type: DatabaseConnectionType{
						Subtype: "postgres",
					},
					Details: DatabaseConnectionDetails{
						ConnectionString: "host=localhost user=test password=test dbname=test",
						Driver:           "postgres",
					},
				},
			},
			want: DatabaseConnectionDetails{
				ConnectionString: "host=localhost user=test password=test dbname=test",
				Driver:           "postgres",
			},
			wantErr: false,
		},
		{
			name: "Test GetDatabaseConnectionDetails with wrong type",
			args: args{
				conn: Connection{
					Name: "test",
					Type: CloudConnectionType{
						Subtype: "aws",
					},
					Details: CloudConnectionDetails{},
				},
			},
			want:    DatabaseConnectionDetails{},
			wantErr: true,
		},
		{
			name: "Test GetDatabaseConnectionDetails with wrong connection details",
			args: args{
				conn: Connection{
					Name: "test",
					Type: DatabaseConnectionType{
						Subtype: "postgres",
					},
					Details: CloudConnectionDetails{},
				},
			},
			want:    DatabaseConnectionDetails{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDatabaseConnectionDetails(tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDatabaseConnectionDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDatabaseConnectionDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}
