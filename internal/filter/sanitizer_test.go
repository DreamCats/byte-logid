package filter

import "testing"

func TestNewMessageSanitizer_ValidPatterns(t *testing.T) {
	patterns := []string{"foo", "bar", `\d+`}
	s, err := NewMessageSanitizer(patterns)
	if err != nil {
		t.Fatalf("NewMessageSanitizer() error: %v", err)
	}
	if s == nil {
		t.Fatal("NewMessageSanitizer() returned nil")
	}
}

func TestNewMessageSanitizer_InvalidPattern(t *testing.T) {
	patterns := []string{"[invalid"}
	_, err := NewMessageSanitizer(patterns)
	if err == nil {
		t.Error("NewMessageSanitizer() should return error for invalid regex")
	}
}

func TestSanitize_RemovesPatterns(t *testing.T) {
	s, _ := NewMessageSanitizer([]string{"_compliance_nlp_log", `"LogID":\s*"[^"]*"`})

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "remove compliance log",
			input: "hello _compliance_nlp_log world",
			want:  "hello world",
		},
		{
			name:  "remove LogID field",
			input: `some text "LogID": "abc123" more text`,
			want:  "some text more text",
		},
		{
			name:  "no match leaves text unchanged",
			input: "normal log message",
			want:  "normal log message",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.Sanitize(tt.input)
			if got != tt.want {
				t.Errorf("Sanitize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitize_CleansMultipleSpaces(t *testing.T) {
	s, _ := NewMessageSanitizer([]string{"REMOVE"})
	got := s.Sanitize("before  REMOVE  after")
	// "before" + "  " + "" + "  " + "after" → cleaned to single spaces
	if got != "before after" {
		t.Errorf("Sanitize() = %q, want %q", got, "before after")
	}
}
