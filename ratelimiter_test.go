package ratelimiter

import (
	"testing"
	"time"
)

func TestPerDuration(t *testing.T) {
	type args struct {
		n        int
		duration time.Duration
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "100 per second",
			args: args{100, time.Second},
			want: 10 * time.Millisecond,
		},
		{
			name: "500 per second",
			args: args{500, time.Second},
			want: 2 * time.Millisecond,
		},
		{
			name: "60 per minute",
			args: args{60, time.Minute},
			want: 1 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PerDuration(tt.args.n, tt.args.duration); got != tt.want {
				t.Errorf("PerDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
