package ai

import (
	"strings"
	"testing"
)

func TestBuildPromptAllTasks(t *testing.T) {
	cases := []struct{ task, contains string }{
		{"generate_bio", "tagline"},
		{"generate_tagline", "tagline"},
		{"generate_project_description", "description"},
		{"generate_experience", "description"},
		{"generate_education", "description"},
	}
	for _, tc := range cases {
		prompt := BuildPrompt(tc.task, map[string]string{"title": "Backend Dev"}, "extra notes")
		if prompt == "" {
			t.Errorf("%s: expected non-empty prompt", tc.task)
		}
		if !strings.Contains(prompt, tc.contains) {
			t.Errorf("%s: prompt missing %q", tc.task, tc.contains)
		}
		if !strings.Contains(prompt, "Backend Dev") {
			t.Errorf("%s: prompt missing context value", tc.task)
		}
		if !strings.Contains(prompt, "extra notes") {
			t.Errorf("%s: prompt missing extra notes", tc.task)
		}
	}
}

func TestBuildPromptUnknownTask(t *testing.T) {
	if prompt := BuildPrompt("bogus_task", nil, ""); prompt != "" {
		t.Errorf("expected empty prompt for unknown task, got %q", prompt)
	}
}

func TestKnownTask(t *testing.T) {
	if !KnownTask("generate_bio") {
		t.Error("expected generate_bio to be known")
	}
	if KnownTask("bogus_task") {
		t.Error("expected bogus_task to be unknown")
	}
}

func TestFormatContextOmitsEmptyValues(t *testing.T) {
	out := formatContext(map[string]string{"title": "Dev", "location": ""})
	if strings.Contains(out, "location") {
		t.Error("empty context value should be omitted")
	}
	if !strings.Contains(out, "title: Dev") {
		t.Error("non-empty context value missing from output")
	}
}

func TestParseFields(t *testing.T) {
	fields, err := ParseFields(`{"tagline":"hi","bio":"world"}`)
	if err != nil {
		t.Fatal(err)
	}
	if fields["tagline"] != "hi" || fields["bio"] != "world" {
		t.Errorf("unexpected fields: %v", fields)
	}
}

func TestParseFieldsFenced(t *testing.T) {
	fields, err := ParseFields("```json\n{\"description\":\"x\"}\n```")
	if err != nil {
		t.Fatal(err)
	}
	if fields["description"] != "x" {
		t.Errorf("unexpected fields: %v", fields)
	}
}

func TestParseFieldsInvalid(t *testing.T) {
	if _, err := ParseFields("not json at all"); err == nil {
		t.Error("expected error for non-JSON input")
	}
}
