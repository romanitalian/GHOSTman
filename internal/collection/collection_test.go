package collection

import (
	"os"
	"testing"
)

func TestSubstituteVariables(t *testing.T) {
	vars := map[string]string{"foo": "bar", "name": "Alice"}
	input := "Hello, {{name}}! Foo is {{foo}}."
	want := "Hello, Alice! Foo is bar."
	got := SubstituteVariables(input, vars)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLoadPostmanCollection(t *testing.T) {
	// Создаём временный файл с минимальной коллекцией
	jsonData := `{"info":{"name":"Test"},"item":[],"variable":[{"key":"foo","value":"bar","type":"string"}]}`
	f, err := os.CreateTemp("", "col_test_*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(jsonData); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	f.Close()

	coll, err := LoadPostmanCollection(f.Name())
	if err != nil {
		t.Fatalf("LoadPostmanCollection error: %v", err)
	}
	if coll.Info.Name != "Test" {
		t.Errorf("unexpected Info.Name: %q", coll.Info.Name)
	}
	if len(coll.Variable) != 1 || coll.Variable[0].Key != "foo" || coll.Variable[0].Value != "bar" {
		t.Errorf("unexpected variable: %+v", coll.Variable)
	}
}
