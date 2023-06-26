package parse

import (
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *Doc
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "enum",
			args: args{
				s: `
fdsafsdafsa
@enum("a:fdsafasd", "b:fsdafdd", "c:fsadfas")
`,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.s)
			spew.Dump(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}
