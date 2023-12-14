package otel

import "testing"

func TestExtractMethodNameFromFullMethod(t *testing.T) {
	type args struct {
		FullMethod string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test-1",
			args: args{
				FullMethod: "/main.ResourceAccess/GetAll",
			},
			want: "main.ResourceAccess/GetAll",
		},
		{
			name: "test-2",
			args: args{
				FullMethod: "/main.UserStruct/GetCurrent",
			},
			want: "main.UserStruct/GetCurrent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractFullMethodNameFromInfoFullMethod(tt.args.FullMethod); got != tt.want {
				t.Errorf("ExtractFullMethodNameFromInfoFullMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetServiceAndMethodFromInfoFullMethod(t *testing.T) {
	type args struct {
		FullMethod string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			name: "test-1",
			args: args{
				FullMethod: "/main.ResourceAccess/GetAll",
			},
			want:  "main.ResourceAccess",
			want1: "GetAll",
		},
		{
			name: "test-2",
			args: args{
				FullMethod: "/main.UserStruct/GetCurrent",
			},
			want:  "main.UserStruct",
			want1: "GetCurrent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetServiceAndMethodFromInfoFullMethod(tt.args.FullMethod)
			if got != tt.want {
				t.Errorf("GetServiceAndMethodFromInfoFullMethod() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetServiceAndMethodFromInfoFullMethod() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
