package template_helper_test

import (
    "os"
    "testing"

    template_helper "crypto_rates_auth/internal/template"
)

const validTemplate = `
<html>
<body>
<h1>{{.Title}}</h1>
</body>
</html>
`

const invalidTemplate = `
<html>
<body>
<h1>{{.Title</h1> <!-- missing closing brace -->
</body>
</html>
`

func setupTemplate(content string) {
    _ = os.MkdirAll("web/templates/auth", 0755)
    _ = os.WriteFile("web/templates/auth/index.html", []byte(content), 0644)
}

func teardownTemplate() {
    _ = os.Remove("web/templates/auth/index.html")
}

func TestLoad_ValidTemplate(t *testing.T) {
    setupTemplate(validTemplate)
    defer teardownTemplate()

    tmpl, err := template_helper.Load()
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if tmpl == nil {
        t.Error("expected template to be loaded, got nil")
    }
}

func TestLoad_MissingFile(t *testing.T) {
    teardownTemplate() // ensure file is gone

    _, err := template_helper.Load()
    if err == nil {
        t.Error("expected error for missing file, got nil")
    }
}

func TestLoad_InvalidTemplate(t *testing.T) {
    setupTemplate(invalidTemplate)
    defer teardownTemplate()

    _, err := template_helper.Load()
    if err == nil {
        t.Error("expected parse error, got nil")
    }
}

