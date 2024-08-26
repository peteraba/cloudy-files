package util

import "fmt"

// FileSize represents a file size with units.
type FileSize struct {
	Unit string
	Size int
}

// String returns the string representation of the file size.
func (f FileSize) String() string {
	return fmt.Sprintf("%d %s", f.Size, f.Unit)
}

// FileSizeFromSize returns a FileSize instance from the given size.
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
