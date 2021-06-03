package program

import "testing"

func Test_isTypeNameString(t *testing.T) {
	type args struct {
		typeStr  string
		typeName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "typeStr equals to typeName",
			args: args{
				typeStr:  "Foo",
				typeName: "Foo",
			},
			want: true,
		},
		{
			name: "typeStr prefix with path",
			args: args{
				typeStr:  "foo.Foo",
				typeName: "Foo",
			},
			want: true,
		},
		{
			name: "typeName has suffix of typeStr",
			args: args{
				typeStr:  "foo.FooFoo",
				typeName: "Foo",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		if got := IsTypeNameString(tt.args.typeStr, tt.args.typeName); got != tt.want {
			t.Errorf("%q. IsTypeNameString() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
