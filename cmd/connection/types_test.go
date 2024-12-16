package connection

import (
	"reflect"
	"sort"
	"testing"
)

func TestGetAvailableConnectionTypes(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "All available connection types",
			want: []string{
				"cloud",
				"kubernetes",
				"db",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAvailableConnectionTypes()
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAvailableConnectionTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}
