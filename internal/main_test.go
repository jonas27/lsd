package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mockYAML = `apiVersion: v1
data:
  password: c2VjcmV0
  app: a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==
kind: Secret
metadata:
  name: "kubernetes secret decoder"
  namespace: ksd
type: Opaque`
	mockJSON = `{
    "apiVersion": "v1",
    "data": {
        "password": "c2VjcmV0",
        "app": "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg=="
    },
    "kind": "Secret",
    "metadata": {
        "name": "kubernetes secret decoder",
        "namespace": "ksd"
    },
    "type": "Opaque"
}`
	mockJSONDecoded = `{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
        "name": "kubernetes secret decoder",
        "namespace": "ksd"
    },
    "stringData": {
        "app": "kubernetes secret decoder",
        "password": "secret"
    },
    "type": "Opaque"
}`
	mockYAMLList = `apiVersion: v1
items:
- data:
    password: "c2VjcmV0"
    app: "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg=="
- data:
    password: "c2VjcmV0"
    app: "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg=="`
	mockYAMLListDecode = `apiVersion: v1
items:
- stringData:
    app: kubernetes secret decoder
    password: secret
- stringData:
    app: kubernetes secret decoder
    password: secret
`
	mockJsonList = `{
apiVersion: "v1",
items: [
  {data: {foo: bar}, apiVersion: "v1"}, {data: {foo: bar}, apiVersion: "v1"}
]}`
)

var plainTexts = []string{
	"this is a plain text",
	`{"this": "is", "a": "text",\n"with": "multiple", "line": "s"}`,
	`version:"2"\ndata:\n\tplain: "yml"`,
	"",
	"text",
	"\t",
	"\n",
	"\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
	"0",
	"0x00000",
}

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		args     string
		expected string
	}{
		// {"no args", []string{}, "the command is intended to work with pipes.\nusage: kubectl get secret <secret"},
		{"json secret", mockJSON, mockJSONDecoded},
    {"yaml list", mockYAMLList, mockYAMLListDecode},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := bytes.NewBufferString(tt.args)
			out, err := run(b)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestRead(t *testing.T) {
	for _, text := range plainTexts {
		reader := strings.NewReader(text)
		assert.Equal(t, text, string(read(reader)))
	}
}

func BenchmarkRead(b *testing.B) {
	for _, text := range plainTexts {
		reader := strings.NewReader(text)
		for n := 0; n < b.N; n++ {
			read(reader)
		}
	}
	b.ReportAllocs()
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name         string
		data         map[string]string
		expected     string
		expectedYAML string
	}{
		{
			name: "json data", data: map[string]string{"password": "c2VjcmV0", "app": "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg=="},
			expected: "{\n    \"app\": \"a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==\",\n    \"password\": \"c2VjcmV0\"\n}",
			expectedYAML: `app: a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==
password: c2VjcmV0
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byt, err := marshal(tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(byt))

			byt, err = marshalYAML(tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedYAML, string(byt))
		})
	}
}

func BenchmarkMarshal(b *testing.B) {
	test := map[string]string{
		"password": "c2VjcmV0",
		"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
	}
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_, _ = marshal(test)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	var j map[string]interface{}
	expected := map[string]interface{}{
		"apiVersion": "v1",
		"data": map[string]interface{}{
			"password": "c2VjcmV0",
			"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
		},
		"kind": "Secret",
		"metadata": map[string]interface{}{
			"name":      "kubernetes secret decoder",
			"namespace": "ksd",
		},
		"type": "Opaque",
	}

	err := unmarshal([]byte(mockJSON), &j, true)
	assert.NoError(t, err)
	assert.Equal(t, expected, j)
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	var j map[string]interface{}
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = unmarshal(nil, &j, true)
	}
}

func TestUnmarshalYaml(t *testing.T) {
	var y map[string]interface{}
	expected := map[string]interface{}{
		"apiVersion": "v1",
		"data": map[string]interface{}{
			"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
			"password": "c2VjcmV0",
		},
		"kind": "Secret",
		"metadata": map[string]interface{}{
			"name":      "kubernetes secret decoder",
			"namespace": "ksd",
		},
		"type": "Opaque",
	}
	err := unmarshal([]byte(mockYAML), &y, false)
	assert.NoError(t, err)
	assert.Equal(t, expected, y)
}

func BenchmarkUnmarshalYaml(b *testing.B) {
	var y map[string]interface{}
	yamlCase, _ := os.ReadFile("./mock.yml")
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = unmarshal(yamlCase, &y, false)
	}
}

func TestIsList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		isList bool
	}{
		{"json no list", mockJSON, false}, {"json list", mockJsonList, true}, {"yaml no list", mockYAML, false}, {"yaml list", mockYAMLList, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var y secret
			b := []byte(tt.input)
			err := unmarshal(b, &y, false)
			assert.NoError(t, err)
		})
	}
}

func TestSecret_Decode(t *testing.T) {
	data := map[string]interface{}{
		"password": "c2VjcmV0",
		"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
	}
	result := decode(data)
	expected := map[string]string{
		"password": "secret",
		"app":      "kubernetes secret decoder",
	}
	assert.Equal(t, expected, result)
}

func BenchmarkSecret_Decode(b *testing.B) {
	data := map[string]interface{}{
		"password": "c2VjcmV0",
		"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
	}

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		decode(data)
	}
}

func TestIsJSONString(t *testing.T) {
	wrongTests := [...][]byte{
		nil,
		[]byte(""),
		[]byte("k"),
		[]byte("-"),
		[]byte(`"test": "case"`),
		[]byte(mockYAML),
	}
	for _, test := range wrongTests {
		if isJSON(test) {
			t.Errorf("%v must not be a json string", string(test))
		}
	}
	successCases := [...][]byte{
		[]byte("null"),
		[]byte(`{"valid":"json"}`),
		[]byte(`{"nested": {"json": "string"}}`),
		[]byte(mockJSON),
	}
	for _, test := range successCases {
		assert.True(t, isJSON(test))
	}
}

func BenchmarkIsJSONString(b *testing.B) {
	jsonCase, _ := os.ReadFile("./mock.json")
	successCases := [...][]byte{
		[]byte("null"),
		[]byte(`{"valid":"json"}`),
		[]byte(`{"nested": {"json": "string"}}`),
		jsonCase,
	}

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		for _, test := range successCases {
			isJSON(test)
		}
	}
}
