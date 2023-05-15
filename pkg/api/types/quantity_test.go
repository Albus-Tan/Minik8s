package types

import "testing"

func TestParseQuantity(t *testing.T) {
	type args struct {
		name ResourceName
		q    Quantity
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "success1",
			args: args{name: ResourceCPU, q: "11111"},
			want: 11111000,
		},
		{
			name: "success2",
			args: args{name: ResourceCPU, q: "100"},
			want: 100000,
		},
		{
			name: "success3",
			args: args{name: ResourceCPU, q: "100m"},
			want: 100,
		},
		{
			name: "success4",
			args: args{name: ResourceCPU, q: "198m"},
			want: 198,
		},
		{
			name: "success5",
			args: args{name: ResourceCPU, q: "1M"},
			want: 1,
		},
		{
			name: "success6",
			args: args{name: ResourceCPU, q: "3"},
			want: 3000,
		},
		{
			name: "success7",
			args: args{name: ResourceMemory, q: "3M"},
			want: 3,
		},
		{
			name: "success8",
			args: args{name: ResourceMemory, q: "300m"},
			want: 300,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQuantity(tt.args.name, tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuantity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseQuantity() got = %v, want %v", got, tt.want)
			}
		})
	}
}
