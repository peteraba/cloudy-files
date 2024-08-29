package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peteraba/cloudy-files/util"
)

func TestFileSize_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sut  util.FileSize
		want string
	}{
		{
			name: "empty",
			sut:  util.FileSize{Unit: "", Size: 0},
			want: "0",
		},
		{
			name: "1 KB",
			sut:  util.FileSize{Unit: "KB", Size: 1},
			want: "1 KB",
		},
		{
			name: "1023 KB",
			sut:  util.FileSize{Unit: "KB", Size: 1023},
			want: "1023 KB",
		},
		{
			name: "99 MB",
			sut:  util.FileSize{Unit: "MB", Size: 99},
			want: "99 MB",
		},
		{
			name: "479 GB",
			sut:  util.FileSize{Unit: "GB", Size: 479},
			want: "479 GB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.sut.String())
		})
	}
}

func TestFileSizeFromSize(t *testing.T) {
	t.Parallel()

	type args struct {
		size int
	}
	tests := []struct {
		name string
		args args
		want util.FileSize
	}{
		{
			name: "empty",
			args: args{size: 0},
			want: util.FileSize{Unit: "", Size: 0},
		},
		{
			name: "1 KB",
			args: args{size: 1024},
			want: util.FileSize{Unit: "KB", Size: 1},
		},
		{
			name: "1023 KB",
			args: args{size: 1024*1024 - 1},
			want: util.FileSize{Unit: "KB", Size: 1023},
		},
		{
			name: "1 MB",
			args: args{size: 1024 * 1024},
			want: util.FileSize{Unit: "MB", Size: 1},
		},
		{
			name: "1 GB",
			args: args{size: 1024 * 1024 * 1024},
			want: util.FileSize{Unit: "GB", Size: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := util.FileSizeFromSize(tt.args.size)

			assert.Equal(t, tt.want.Unit, fs.Unit)
			assert.Equal(t, tt.want.Size, fs.Size)
		})
	}
}
