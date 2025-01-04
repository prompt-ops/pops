package common

import (
	"reflect"
	"testing"
)

func TestNewCloudConnection(t *testing.T) {
	type args struct {
		name     string
		provider AvailableCloudConnectionType
	}
	tests := []struct {
		name string
		args args
		want Connection
	}{
		{
			name: "Test NewCloudConnection",
			args: args{
				name:     "test",
				provider: AWSCloudConnection,
			},
			want: Connection{
				Name: "test",
				Type: CloudConnectionType{
					MainType: ConnectionTypeCloud,
					Subtype:  "AWS",
				},
				Details: CloudConnectionDetails{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCloudConnection(tt.args.name, tt.args.provider); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCloudConnection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCloudConnectionDetails(t *testing.T) {
	type args struct {
		conn Connection
	}
	tests := []struct {
		name    string
		args    args
		want    CloudConnectionDetails
		wantErr bool
	}{
		{
			name: "Test GetCloudConnectionDetails",
			args: args{
				conn: Connection{
					Name: "test",
					Type: CloudConnectionType{
						Subtype: "aws",
					},
					Details: CloudConnectionDetails{},
				},
			},
			want:    CloudConnectionDetails{},
			wantErr: false,
		},
		{
			name: "Test GetCloudConnectionDetails with wrong type",
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
			want:    CloudConnectionDetails{},
			wantErr: true,
		},
		{
			name: "Test GetCloudConnectionDetails with wrong connection details",
			args: args{
				conn: Connection{
					Name: "test",
					Type: CloudConnectionType{
						Subtype: "aws",
					},
					Details: DatabaseConnectionDetails{
						ConnectionString: "host=localhost user=test password=test dbname=test",
						Driver:           "postgres",
					},
				},
			},
			want:    CloudConnectionDetails{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCloudConnectionDetails(tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCloudConnectionDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCloudConnectionDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}
