package ai

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func KnownTask(task string) bool {
	switch task {
	case "generate_bio",
		"generate_tagline",
		"generate_project_description",
		"generate_experience",
		"generate_education":
		return true
	}
	return false
}

func BuildPrompt(task string, context map[string]string, extra string) string {
	ctx := formatContext(context)
	notes := strings.TrimSpace(extra)
	if notes == "" {
		notes = "(none provided)"
	}

	switch task {
	case "generate_bio":
		return fmt.Sprintf("You are a professional resume writer helping a developer craft portfolio copy.\n\nContext:\n%s\n\nUser's extra notes:\n%s\n\nRespond with JSON in exactly this shape (no extra keys):\n{\"tagline\": \"one sentence, at most 20 words\", \"bio\": \"first-person paragraph, 40 to 80 words\"}", ctx, notes)
	case "generate_tagline":
		return fmt.Sprintf("You are a professional resume writer.\n\nContext:\n%s\n\nUser's extra notes:\n%s\n\nRespond with JSON in exactly this shape (no extra keys):\n{\"tagline\": \"one sentence, at most 20 words\"}", ctx, notes)
	case "generate_project_description":
		return fmt.Sprintf("You are a technical writer crafting project descriptions for a developer portfolio.\n\nContext:\n%s\n\nUser's extra notes:\n%s\n\nRespond with JSON in exactly this shape (no extra keys):\n{\"description\": \"3 to 5 sentences about what the project does, the problem it solves, and the value it provides\"}", ctx, notes)
	case "generate_experience":
		return fmt.Sprintf("You are a professional resume writer.\n\nContext:\n%s\n\nUser's extra notes:\n%s\n\nRespond with JSON in exactly this shape (no extra keys):\n{\"description\": \"3 to 6 bullet points of achievements in first-person past tense, using numbers and measurable outcomes where possible\"}", ctx, notes)
	case "generate_education":
		return fmt.Sprintf("You are a professional resume writer.\n\nContext:\n%s\n\nUser's extra notes:\n%s\n\nRespond with JSON in exactly this shape (no extra keys):\n{\"description\": \"2 to 4 sentences about the degree, relevant coursework, or achievements\"}", ctx, notes)
	}
	return ""
}

func formatContext(context map[string]string) string {
	keys := make([]string, 0, len(context))
	for k := range context {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		if v := strings.TrimSpace(context[k]); v != "" {
			fmt.Fprintf(&b, "- %s: %s\n", k, v)
		}
	}
	if b.Len() == 0 {
		return "(none provided)"
	}
	return b.String()
}

var codeFenceRE = regexp.MustCompile("(?s)^```[a-zA-Z]*\\s*|\\s*```$")

func ParseFields(text string) (map[string]string, error) {
	cleaned := codeFenceRE.ReplaceAllString(strings.TrimSpace(text), "")
	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON object found in model response")
	}
	var fields map[string]string
	if err := json.Unmarshal([]byte(cleaned[start:end+1]), &fields); err != nil {
		return nil, fmt.Errorf("parse model JSON: %w", err)
	}
	return fields, nil
}
