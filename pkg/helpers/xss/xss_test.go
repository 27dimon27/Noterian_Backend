package xss

import (
	"testing"
)

type TestStruct struct {
	Name     string
	Age      int
	Email    string
	Nested   *NestedStruct
	Tags     []string
	Metadata map[string]string
}

type NestedStruct struct {
	Content string
	Value   int
}

func TestSanitizeStruct_String(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes script tag",
			input:    "<script>alert('xss')</script>Hello",
			expected: "Hello",
		},
		{
			name:     "removes onclick attribute",
			input:    `<div onclick="alert('xss')">Click</div>`,
			expected: "Click",
		},
		{
			name:     "removes img onerror",
			input:    `<img src=x onerror=alert('xss')>`,
			expected: "",
		},
		{
			name:     "unescapes HTML entities",
			input:    "&lt;div&gt;Hello&lt;/div&gt;",
			expected: "<div>Hello</div>",
		},
		{
			name:     "normal text unchanged",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "removes javascript protocol",
			input:    `<a href="javascript:alert('xss')">Link</a>`,
			expected: "Link",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &TestStruct{
				Name: tt.input,
			}
			SanitizeStruct(obj)
			if obj.Name != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, obj.Name)
			}
		})
	}
}

func TestSanitizeStruct_NilPointer(t *testing.T) {
	var obj *TestStruct = nil
	SanitizeStruct(obj)
}

func TestSanitizeStruct_NonPointer(t *testing.T) {
	obj := TestStruct{Name: "<script>test</script>"}
	SanitizeStruct(obj)
	if obj.Name != "<script>test</script>" {
		t.Errorf("non-pointer should not be modified")
	}
}

func TestSanitizeStruct_NestedStruct(t *testing.T) {
	obj := &TestStruct{
		Name: "<script>name</script>",
		Nested: &NestedStruct{
			Content: "<script>nested</script>",
			Value:   42,
		},
	}

	SanitizeStruct(obj)

	if obj.Name != "" {
		t.Errorf("expected %q, got %q", "", obj.Name)
	}
	if obj.Nested.Content != "" {
		t.Errorf("expected %q, got %q", "", obj.Nested.Content)
	}
	if obj.Nested.Value != 42 {
		t.Errorf("expected %d, got %d", 42, obj.Nested.Value)
	}
}

func TestSanitizeStruct_Slice(t *testing.T) {
	obj := &TestStruct{
		Tags: []string{"<script>tag1</script>", "normal", "<img src=x onerror=alert(1)>"},
	}

	SanitizeStruct(obj)

	expected := []string{"", "normal", ""}
	for i := range obj.Tags {
		if obj.Tags[i] != expected[i] {
			t.Errorf("index %d: expected %q, got %q", i, expected[i], obj.Tags[i])
		}
	}
}

func TestSanitizeStruct_Map(t *testing.T) {
	obj := &TestStruct{
		Metadata: map[string]string{
			"<script>key</script>": "value",
			"normal_key":           "<script>value</script>",
			"safe":                 "safe",
		},
	}

	SanitizeStruct(obj)

	for key, value := range obj.Metadata {
		if key == "safe" && value != "safe" {
			t.Errorf("safe key/value should remain unchanged")
		}
		if key == "normal_key" && value != "" {
			t.Errorf("expected value '', got %q", value)
		}
	}
}

func TestSanitizeStruct_Interface(t *testing.T) {
	var data interface{} = &TestStruct{
		Name: "<script>interface</script>",
	}

	SanitizeStruct(&data)

	obj := data.(*TestStruct)
	if obj.Name != "" {
		t.Errorf("expected %q, got %q", "", obj.Name)
	}
}

func TestSanitizeStruct_PreservesNonStringFields(t *testing.T) {
	obj := &TestStruct{
		Age:   25,
		Email: "test@example.com",
	}

	SanitizeStruct(obj)

	if obj.Age != 25 {
		t.Errorf("age should remain 25, got %d", obj.Age)
	}
	if obj.Email != "test@example.com" {
		t.Errorf("email should remain unchanged, got %q", obj.Email)
	}
}

func BenchmarkSanitizeStruct(b *testing.B) {
	obj := &TestStruct{
		Name:     "<script>alert('xss')</script>John",
		Email:    "john@example.com",
		Tags:     []string{"<b>tag1</b>", "<i>tag2</i>"},
		Metadata: map[string]string{"key": "<script>value</script>"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeStruct(obj)
	}
}
