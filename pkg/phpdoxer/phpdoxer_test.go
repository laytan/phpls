package phpdoxer

import (
	"reflect"
	"testing"
)

func Test_phpdoxer_All(t *testing.T) {
	tests := []struct {
		name string
		p    *phpdoxer
		args string
		want map[string]string
	}{
		{
			name: "empty",
			p:    &phpdoxer{},
			args: "",
			want: map[string]string{},
		},
		{
			name: "empty doc",
			p:    &phpdoxer{},
			args: `
            /**
             *
             */
            `,
			want: map[string]string{},
		},
		{
			name: "@var no content",
			p:    &phpdoxer{},
			args: `
            /**
             * @var 
             */
            `,
			want: map[string]string{
				"var": "",
			},
		},
		{
			name: "@var name",
			p:    &phpdoxer{},
			args: `
            /**
             * @var \Test\Class
             */
            `,
			want: map[string]string{
				"var": "\\Test\\Class",
			},
		},
		{
			name: "@var union",
			p:    &phpdoxer{},
			args: `
            /**
             * @var \Test\Class|\Test\Class
             */
            `,
			want: map[string]string{
				"var": "\\Test\\Class|\\Test\\Class",
			},
		},
		{
			name: "@var descripted",
			p:    &phpdoxer{},
			args: `
            /**
             * @var \Test\Class|\Test\Class description
             */
            `,
			want: map[string]string{
				"var": "\\Test\\Class|\\Test\\Class description",
			},
		},
		{
			name: "@var and @method",
			p:    &phpdoxer{},
			args: `
            /**
             * @method 
             * @var \Test\Class|\Test\Class description
             */
            `,
			want: map[string]string{
				"method": "",
				"var":    "\\Test\\Class|\\Test\\Class description",
			},
		},
		{
			name: "inline",
			p:    &phpdoxer{},
			args: "// @var \\Test\\Class test",
			want: map[string]string{
				"var": "\\Test\\Class test",
			},
		},
		{
			name: "newlines",
			p:    &phpdoxer{},
			args: `
            /**
             * @method 
             * @var \Test\Class|\Test\Class description
             *
             * @return string
             */
            `,
			want: map[string]string{
				"method": "",
				"var":    "\\Test\\Class|\\Test\\Class description",
				"return": "string",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &phpdoxer{}
			if got := p.All(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("phpdoxer.All() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_phpdoxer_Var(t *testing.T) {
	tests := []struct {
		name string
		p    *phpdoxer
		args string
		want []string
	}{
		{
			name: "empty",
			p:    &phpdoxer{},
			args: "",
			want: []string{},
		},
		{
			name: "simple var",
			p:    &phpdoxer{},
			args: "// @var string",
			want: []string{"string"},
		},
		{
			name: "union var",
			p:    &phpdoxer{},
			args: `
            /**
             * @var \Test\Class|string|null
             */
            `,
			want: []string{"\\Test\\Class", "string", "null"},
		},
		{
			name: "description",
			p:    &phpdoxer{},
			args: `
            /**
             * @var \Test\Class description
             */
            `,
			want: []string{"\\Test\\Class"},
		},
		{
			name: "not a var",
			p:    &phpdoxer{},
			args: `
            /**
             * @method test()
             */
            `,
			want: []string{},
		},
		{
			name: "multiple gives last one",
			p:    &phpdoxer{},
			args: `
            /**
             * @var \Test\Class desc
             * @method
             * @var \Testie\Classie descie
             */
            `,
			want: []string{"\\Testie\\Classie"},
		},
		{
			name: "single null",
			p:    &phpdoxer{},
			args: `
            /**
             * @var \Test|null
             */
            `,
			want: []string{"\\Test", "null"},
		},
		{
			name: "single short null",
			p:    &phpdoxer{},
			args: `
            /**
             * @var ?\Test\Null
             */
            `,
			want: []string{"\\Test\\Null", "null"},
		},
		{
			name: "two nulls",
			p:    &phpdoxer{},
			args: `
            /**
             * @var null|?Test
             */
            `,
			want: []string{"null", "Test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &phpdoxer{}
			if got := p.Var(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("phpdoxer.Var() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_phpdoxer_Return(t *testing.T) {
	tests := []struct {
		name string
		p    *phpdoxer
		args string
		want []string
	}{
		{
			name: "empty",
			p:    &phpdoxer{},
			args: "",
			want: []string{},
		},
		{
			name: "simple var",
			p:    &phpdoxer{},
			args: "// @return string",
			want: []string{"string"},
		},
		{
			name: "union var",
			p:    &phpdoxer{},
			args: `
            /**
             * @return \Test\Class|string|null
             */
            `,
			want: []string{"\\Test\\Class", "string", "null"},
		},
		{
			name: "description",
			p:    &phpdoxer{},
			args: `
            /**
             * @return \Test\Class description
             */
            `,
			want: []string{"\\Test\\Class"},
		},
		{
			name: "not a var",
			p:    &phpdoxer{},
			args: `
            /**
             * @method test()
             */
            `,
			want: []string{},
		},
		{
			name: "multiple gives last one",
			p:    &phpdoxer{},
			args: `
            /**
             * @return \Test\Class desc
             * @method
             * @return \Testie\Classie descie
             */
            `,
			want: []string{"\\Testie\\Classie"},
		},
		{
			name: "single null",
			p:    &phpdoxer{},
			args: `
            /**
             * @return \Test|null
             */
            `,
			want: []string{"\\Test", "null"},
		},
		{
			name: "single short null",
			p:    &phpdoxer{},
			args: `
            /**
             * @return ?\Test\Null
             */
            `,
			want: []string{"\\Test\\Null", "null"},
		},
		{
			name: "two nulls",
			p:    &phpdoxer{},
			args: `
            /**
             * @return null|?Test
             */
            `,
			want: []string{"null", "Test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &phpdoxer{}
			if got := p.Return(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("phpdoxer.Return() = %v, want %v", got, tt.want)
			}
		})
	}
}
