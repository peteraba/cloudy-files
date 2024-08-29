package util_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peteraba/cloudy-files/util"
)

func TestSpy_GetError(t *testing.T) {
	t.Parallel()

	t.Run("panic when calling method with few arguments", func(t *testing.T) {
		t.Parallel()

		// setup
		s := util.NewSpy()
		s.Register("foo", 0, assert.AnError, "bar")

		// execute
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("GetError() did not panic")
			}
		}()

		_ = s.GetError("foo")
	})

	t.Run("panic when calling method with too many arguments", func(t *testing.T) {
		t.Parallel()

		// setup
		s := util.NewSpy()
		s.Register("foo", 0, assert.AnError)

		// execute
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("GetError() did not panic")
			}
		}()

		_ = s.GetError("foo", "bar")
	})

	type fields struct {
		Methods map[string][]util.InAndOut
	}
	type args struct {
		method string
		args   []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		calls   int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "empty",
			fields:  fields{},
			args:    args{method: "foo"},
			calls:   1,
			wantErr: assert.NoError,
		},
		{
			name: "simple method call resulting in error",
			fields: fields{
				Methods: map[string][]util.InAndOut{
					"foo": {
						{In: []interface{}{}, Out: assert.AnError},
					},
				},
			},
			args:    args{method: "foo"},
			calls:   1,
			wantErr: assert.Error,
		},
		{
			name: "method call with concrete argument expected not resulting in error",
			fields: fields{
				Methods: map[string][]util.InAndOut{
					"foo": {
						{Skip: 0, In: []interface{}{"bar"}, Out: assert.AnError},
					},
				},
			},
			args:    args{method: "foo", args: []interface{}{"baz"}},
			calls:   1,
			wantErr: assert.NoError,
		},
		{
			name: "method call with concrete argument expected resulting in error",
			fields: fields{
				Methods: map[string][]util.InAndOut{
					"foo": {
						{Skip: 0, In: []interface{}{"bar"}, Out: assert.AnError},
					},
				},
			},
			args:    args{method: "foo", args: []interface{}{"bar"}},
			calls:   1,
			wantErr: assert.Error,
		},
		{
			name: "method call with any argument expected resulting in error",
			fields: fields{
				Methods: map[string][]util.InAndOut{
					"foo": {
						{Skip: 0, In: []interface{}{util.Any}, Out: assert.AnError},
					},
				},
			},
			args:    args{method: "foo", args: []interface{}{"bar"}},
			calls:   1,
			wantErr: assert.Error,
		},
		{
			name: "method call with any argument expected not resulting in error on first call",
			fields: fields{
				Methods: map[string][]util.InAndOut{
					"foo": {
						{Skip: 1, In: []interface{}{util.Any}, Out: assert.AnError},
					},
				},
			},
			args:    args{method: "foo", args: []interface{}{"bar"}},
			calls:   1,
			wantErr: assert.NoError,
		},
		{
			name: "method call with any argument expected resulting in error on second call",
			fields: fields{
				Methods: map[string][]util.InAndOut{
					"foo": {
						{Skip: 1, In: []interface{}{util.Any}, Out: assert.AnError},
					},
				},
			},
			args:    args{method: "foo", args: []interface{}{"bar"}},
			calls:   2,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// setup
			sut := util.NewSpy()

			for method, inAndOuts := range tt.fields.Methods {
				for _, inAndOut := range inAndOuts {
					sut.Register(method, inAndOut.Skip, inAndOut.Out, inAndOut.In...)
				}
			}

			// execute
			var actual error
			for range tt.calls {
				actual = sut.GetError(tt.args.method, tt.args.args...)
			}

			// assert
			tt.wantErr(t, actual, fmt.Sprintf("GetError(%v, %v)", tt.args.method, tt.args.args))
		})
	}
}
