package handlers

import (
	"net/http"
)

// DocsHandler serves Swagger UI
type DocsHandler struct{}

// NewDocsHandler creates a new docs handler
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// ServeSwaggerUI serves the Swagger UI HTML page
func (h *DocsHandler) ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Inbox Allocation Service - API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui.css">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
        .topbar {
            display: none;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "/api/openapi.yaml",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                defaultModelsExpandDepth: 1,
                defaultModelExpandDepth: 1,
                docExpansion: "list",
                filter: true,
                showRequestHeaders: true,
                tryItOutEnabled: true
            });
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// ServeOpenAPISpec serves the OpenAPI YAML specification
func (h *DocsHandler) ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	// The openapi.yaml file will be served from the filesystem
	http.ServeFile(w, r, "./api/openapi.yaml")
}
