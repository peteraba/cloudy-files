package repo

import "context"

type Store interface {
	Read(ctx context.Context) ([]byte, error)
	ReadForWrite(ctx context.Context) ([]byte, error)
	WriteLocked(ctx context.Context, data []byte) error
	Unlock(ctx context.Context) error
	Write(ctx context.Context, data []byte) error
}
