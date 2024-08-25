package util

import "fmt"

type FileSize struct {
	Unit string
	Size int
}

func (f FileSize) String() string {
	return fmt.Sprintf("%d %s", f.Size, f.Unit)
}

// sdkfsd.
func FileSizeFromSize(size int) FileSize {
	units := []string{"bytes", "KB", "MB", "GB", "TB", "PB"}

	var unit string

	i := 0
	for size >= 1024 && i < len(units) {
		size /= 1024
		i++
		unit = units[i]
	}

	return FileSize{
		Size: size,
		Unit: unit,
	}
}
