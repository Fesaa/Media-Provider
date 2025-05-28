package utils

import "testing"

func TestExt(t *testing.T) {
	type args struct {
		uri        string
		defaultExt []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Simple",
			args: args{
				uri:        "https://www.example.com",
				defaultExt: nil,
			},
			want: ".com",
		},
		{
			name: "query",
			args: args{
				uri:        "https://www.example.com/file.webp?q=40",
				defaultExt: nil,
			},
			want: ".webp",
		},
		{
			name: "Anchor",
			args: args{
				uri:        "https://www.example.com/file.webp?q=40#80",
				defaultExt: nil,
			},
			want: ".xwebp",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Ext(tt.args.uri, tt.args.defaultExt...); got != tt.want {
				t.Errorf("Ext() = %v, want %v", got, tt.want)
			}
		})
	}
}
