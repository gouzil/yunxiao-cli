package output

import (
	"bytes"
	"testing"
)

type sampleValue struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestSelectFieldsRejectsUnknownField(t *testing.T) {
	_, err := SelectFields(sampleValue{ID: "1", Name: "api"}, []string{"missing"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRendererJSONWithJQ(t *testing.T) {
	buffer := bytes.Buffer{}
	renderer := NewRenderer(&buffer)
	err := renderer.Render(sampleValue{ID: "1", Name: "api"}, Table{}, Options{JSONFields: []string{"name"}, JQ: ".name"})
	if err != nil {
		t.Fatal(err)
	}
	if got := buffer.String(); got != "\"api\"\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestRenderTableEmpty(t *testing.T) {
	buffer := bytes.Buffer{}
	renderer := NewRenderer(&buffer)
	if err := renderer.Render(nil, Table{Empty: "No repositories found."}, Options{}); err != nil {
		t.Fatal(err)
	}
	if got := buffer.String(); got != "No repositories found.\n" {
		t.Fatalf("output = %q", got)
	}
}
