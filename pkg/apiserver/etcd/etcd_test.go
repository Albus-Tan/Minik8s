package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"reflect"
	"testing"
)

func TestAll(t *testing.T) {
	t.Run("TestInit", TestInit)
	t.Run("TestClear", TestClear)
	t.Run("TestPut", TestPut)
	t.Run("TestHas", TestHas)
	t.Run("TestGet", TestGet)
	t.Run("TestGetAllWithPrefix", TestGetAllWithPrefix)
	t.Run("TestDelete", TestDelete)
	t.Run("TestClear", TestClear)
	t.Run("TestClose", TestClose)
}

func TestInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "success"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init()
		})
	}
}

func TestPut(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success1",
			args: args{key: "123", value: "11111"},
		},
		{
			name: "success2",
			args: args{key: "abc", value: "base"},
		},
		{
			name: "success3",
			args: args{key: "abg", value: "base1"},
		},
		{
			name: "success4",
			args: args{key: "at", value: "base2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err, _ := Put(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Put() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name      string
		args      args
		wantValue string
		wantErr   bool
	}{
		{
			name:      "success1",
			args:      args{key: "123"},
			wantValue: "11111",
		},
		{
			name:      "success2",
			args:      args{key: "abc"},
			wantValue: "base",
		},
		{
			name:      "no_value1",
			args:      args{key: "a"},
			wantValue: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, err := Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("Get() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestGetAllWithPrefix(t *testing.T) {
	type args struct {
		keyPrefix string
	}
	tests := []struct {
		name       string
		args       args
		wantValues []string
		wantErr    bool
	}{
		{
			name:       "success1",
			args:       args{keyPrefix: "123"},
			wantValues: []string{"11111"},
		},
		{
			name:       "success2",
			args:       args{keyPrefix: "12"},
			wantValues: []string{"11111"},
		},
		{
			name:       "success3",
			args:       args{keyPrefix: "abc"},
			wantValues: []string{"base"},
		},
		{
			name:       "success4",
			args:       args{keyPrefix: "ab"},
			wantValues: []string{"base", "base1"},
		},
		{
			name:       "success5",
			args:       args{keyPrefix: "a"},
			wantValues: []string{"base", "base1", "base2"},
		},
		{
			name:       "success all",
			args:       args{keyPrefix: ""},
			wantValues: []string{"11111", "base", "base1", "base2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValues, err := GetAllWithPrefix(tt.args.keyPrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotValues, tt.wantValues) {
				t.Errorf("GetAllWithPrefix() gotValues = %v, want %v", gotValues, tt.wantValues)
			}
		})
	}
}

func TestHas(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name      string
		args      args
		wantValue bool
		wantErr   bool
	}{
		{
			name:      "success1",
			args:      args{key: "123"},
			wantValue: true,
		},
		{
			name:      "success2",
			args:      args{key: "abc"},
			wantValue: true,
		},
		{
			name:      "fail1",
			args:      args{key: "abd"},
			wantValue: false,
		},
		{
			name:      "fail2",
			args:      args{key: "ab"},
			wantValue: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, err := Has(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Has() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("Has() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{key: "123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Delete(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteAllWithPrefix(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success1",
			args: args{key: "a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteAllWithPrefix(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("DeleteAllWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClear(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "success"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Clear(); (err != nil) != tt.wantErr {
				t.Errorf("Clear() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWatch(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name  string
		args  args
		want  context.CancelFunc
		want1 chan *Event
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := Watch(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Watch() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Watch() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestWatchAllWithPrefix(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name  string
		args  args
		want  context.CancelFunc
		want1 chan *Event
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := WatchAllWithPrefix(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WatchAllWithPrefix() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("WatchAllWithPrefix() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_doWatch(t *testing.T) {
	type args struct {
		rch clientv3.WatchChan
		ch  chan *Event
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doWatch(tt.args.rch, tt.args.ch)
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "success"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Close()
		})
	}
}
