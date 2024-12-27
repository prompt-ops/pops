package config

import (
	"reflect"
	"testing"
)

func TestNewCloudConnection(t *testing.T) {
	type args struct {
		name     string
		provider string
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
				provider: "aws",
			},
			want: Connection{
				Type:    "cloud",
				Name:    "test",
				SubType: "aws",
				ConnectionDetails: CloudConnectionDetails{
					Provider: "aws",
				},
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
				Type:    "kubernetes",
				Name:    "test",
				SubType: "test-context",
				ConnectionDetails: KubernetesConnectionDetails{
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

func TestNewDatabaseConnection(t *testing.T) {
	type args struct {
		name             string
		driver           string
		connectionString string
	}
	tests := []struct {
		name string
		args args
		want Connection
	}{
		{
			name: "Test NewDatabaseConnection",
			args: args{
				name:             "test",
				driver:           "postgres",
				connectionString: "host=localhost user=test password=test dbname=test",
			},
			want: Connection{
				Type:    "database",
				Name:    "test",
				SubType: "postgres",
				ConnectionDetails: DatabaseConnectionDetails{
					ConnectionString: "host=localhost user=test password=test dbname=test",
					Driver:           "postgres",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDatabaseConnection(tt.args.name, tt.args.driver, tt.args.connectionString); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDatabaseConnection() = %v, want %v", got, tt.want)
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
					Type:    "cloud",
					Name:    "test",
					SubType: "aws",
					ConnectionDetails: CloudConnectionDetails{
						Provider: "aws",
					},
				},
			},
			want: CloudConnectionDetails{
				Provider: "aws",
			},
			wantErr: false,
		},
		{
			name: "Test GetCloudConnectionDetails with wrong type",
			args: args{
				conn: Connection{
					Type:    "database",
					Name:    "test",
					SubType: "postgres",
					ConnectionDetails: DatabaseConnectionDetails{
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
					Type:    "cloud",
					Name:    "test",
					SubType: "aws",
					ConnectionDetails: DatabaseConnectionDetails{
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
					Type:    "kubernetes",
					Name:    "test",
					SubType: "test-context",
					ConnectionDetails: KubernetesConnectionDetails{
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
					Type:    "database",
					Name:    "test",
					SubType: "postgres",
					ConnectionDetails: DatabaseConnectionDetails{
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
					Type:    "kubernetes",
					Name:    "test",
					SubType: "test-context",
					ConnectionDetails: DatabaseConnectionDetails{
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
					Type:    "database",
					Name:    "test",
					SubType: "postgres",
					ConnectionDetails: DatabaseConnectionDetails{
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
					Type:    "cloud",
					Name:    "test",
					SubType: "aws",
					ConnectionDetails: CloudConnectionDetails{
						Provider: "aws",
					},
				},
			},
			want:    DatabaseConnectionDetails{},
			wantErr: true,
		},
		{
			name: "Test GetDatabaseConnectionDetails with wrong connection details",
			args: args{
				conn: Connection{
					Type:    "database",
					Name:    "test",
					SubType: "postgres",
					ConnectionDetails: CloudConnectionDetails{
						Provider: "aws",
					},
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
