package convert

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

type T1 struct {
	String string
	Int    int
	Bool   bool
	Array  []string
	Nil    *uint
	IntP   *int
}

type T2 struct {
	String string
	Int    int
	Bool   bool
}

type T3 struct {
	String *string
	Int    *int
	Bool   *bool
}

type T4 struct {
	IntP *int
}

type T5 struct {
	IntP int
}

type T6 struct {
	Foo string `convert:"Bar"`
}

type T7 struct {
	Bar string
}

type T8 struct {
	Bar int
}

type T9 struct {
	Time *time.Time `convert:"TimeUnix"`
}

type T10 struct {
	TimeUnix int64 `convert:"Time"`
}

var t1 = time.Date(2020, 11, 7, 0, 0, 0, 0, time.UTC)

func TestRegisterConverter(t *testing.T) {
	type args struct {
		from reflect.Type
		to   reflect.Type
		fn   Fn
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "simple",
			args: args{
				from: reflect.TypeOf(T1{}),
				to:   reflect.TypeOf(T2{}),
				fn:   func(from interface{}) (interface{}, error) { return nil, nil },
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Register(tt.args.from, tt.args.to, tt.args.fn)
		})
	}
}

type TFn struct {
	From, To reflect.Type
	Fn       Fn
}

var (
	TimeToInt = TFn{
		From: reflect.TypeOf(time.Time{}),
		To:   reflect.TypeOf(int64(0)),
		Fn: func(from interface{}) (interface{}, error) {
			t, _ := from.(time.Time)
			return t.Unix(), nil
		},
	}
	IntToTime = TFn{
		From: reflect.TypeOf(int64(0)),
		To:   reflect.TypeOf(time.Time{}),
		Fn: func(from interface{}) (interface{}, error) {
			t, _ := from.(int64)
			return time.Unix(t, 0), nil
		},
	}
	StringToInt = TFn{
		From: reflect.TypeOf(""),
		To:   reflect.TypeOf(0),
		Fn: func(from interface{}) (interface{}, error) {
			s := from.(string)
			return strconv.Atoi(s)
		},
	}
)

func TestConvert(t *testing.T) {
	type args struct {
		from interface{}
		to   interface{}
		Fns  []TFn
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "simple case",
			args: args{
				from: T1{String: "t1-id", Int: 42, Bool: true},
				to:   &T2{},
			},
			want:    &T2{String: "t1-id", Int: 42, Bool: true},
			wantErr: false,
		},
		{
			name: "errors out on non-pointer receiver",
			args: args{
				from: T1{String: "t1-id", Int: 42, Bool: true},
				to:   T2{},
			},
			want:    T2{},
			wantErr: true,
		},
		{
			name: "converts to corresponding pointer types",
			args: args{
				from: T1{String: "t1-id", Int: 42, Bool: true},
				to:   &T3{},
			},
			want:    &T3{String: String("t1-id"), Int: Int(42), Bool: Bool(true)},
			wantErr: false,
		},
		{
			name: "converts pointer to pointer",
			args: args{
				from: T1{IntP: Int(42)},
				to:   &T4{},
			},
			want:    &T4{IntP: Int(42)},
			wantErr: false,
		},
		{
			name: "converts pointer to non-pointer",
			args: args{
				from: T1{IntP: Int(42)},
				to:   &T5{},
			},
			want:    &T5{IntP: 42},
			wantErr: false,
		},
		{
			name: "converts pointer to non-pointer",
			args: args{
				from: T1{IntP: Int(42)},
				to:   &T5{},
			},
			want:    &T5{IntP: 42},
			wantErr: false,
		},
		{
			name: "copy to a field with a different name",
			args: args{
				from: T6{Foo: "yolo"},
				to:   &T7{},
			},
			want:    &T7{Bar: "yolo"},
			wantErr: false,
		},
		{
			name: "copy to a field with a different name and type",
			args: args{
				from: T6{Foo: "42"},
				to:   &T8{},
				Fns:  []TFn{StringToInt},
			},
			want:    &T8{Bar: 42},
			wantErr: false,
		},
		{
			name: "copy to a field with a different type " +
				"without registering doesn't copy the value",
			args: args{
				from: T6{Foo: "42"},
				to:   &T8{},
			},
			want:    &T8{},
			wantErr: false,
		},
		{
			name: "registered converter func errors are propagated up",
			args: args{
				from: T6{Foo: "yolo"},
				to:   &T8{},
				Fns:  []TFn{StringToInt},
			},
			want:    &T8{},
			wantErr: true,
		},
		{
			name: "covert custom type pointer to non-pointer",
			args: args{
				from: T9{
					Time: &t1,
				},
				to:  &T10{},
				Fns: []TFn{TimeToInt},
			},
			want:    &T10{TimeUnix: 1604707200},
			wantErr: false,
		},
		{
			name: "covert custom type non-pointer to pointer",
			args: args{
				from: T9{
					Time: &t1,
				},
				to:  &T10{},
				Fns: []TFn{TimeToInt},
			},
			want:    &T10{TimeUnix: 1604707200},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converters = map[reflect.Type]map[reflect.Type]Fn{}
			for _, fn := range tt.args.Fns {
				Register(fn.From, fn.To, fn.Fn)
			}

			err := Convert(tt.args.from, tt.args.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.args.to, tt.want) {
				t.Errorf("Convert() res = %v, want %v", tt.args.to, tt.want)
			}
		})
	}
}

func String(s string) *string {
	return &s
}

func Int(i int) *int {
	return &i
}

func Bool(b bool) *bool {
	return &b
}
