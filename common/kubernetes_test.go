package common

import (
	"reflect"
	"testing"
)

func TestNewKubernetesConnection(t *testing.T) {
	type args struct {
		name    string
		context string
	}
	tests := []struct {
		name string
		args args
		want Connection
	}{
		{
			name: "Test NewKubernetesConnection",
			args: args{
				name:    "test",
				context: "test-context",
			},
			want: Connection{
				Name: "test",
				Type: KubernetesConnectionType{},
				Details: KubernetesConnectionDetails{
					SelectedContext: "test-context",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewKubernetesConnection(tt.args.name, tt.args.context); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKubernetesConnection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetKubernetesConnectionDetails(t *testing.T) {
	type args struct {
		conn Connection
	}
	tests := []struct {
		name    string
		args    args
		want    KubernetesConnectionDetails
		wantErr bool
	}{
		{
			name: "Test GetKubernetesConnectionDetails",
			args: args{
				conn: Connection{
					Name: "test",
					Type: KubernetesConnectionType{},
					Details: KubernetesConnectionDetails{
						SelectedContext: "test-context",
					},
				},
			},
			want: KubernetesConnectionDetails{
				SelectedContext: "test-context",
			},
			wantErr: false,
		},
		{
			name: "Test GetKubernetesConnectionDetails with wrong type",
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
			want:    KubernetesConnectionDetails{},
			wantErr: true,
		},
		{
			name: "Test GetKubernetesConnectionDetails with wrong connection details",
			args: args{
				conn: Connection{
					Name: "test",
					Type: KubernetesConnectionType{},
					Details: DatabaseConnectionDetails{
						ConnectionString: "host=localhost user=test password=test dbname=test",
						Driver:           "postgres",
					},
				},
			},
			want:    KubernetesConnectionDetails{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetKubernetesConnectionDetails(tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKubernetesConnectionDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetKubernetesConnectionDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}
