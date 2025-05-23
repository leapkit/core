package render

import (
	"bytes"
	"testing"
	"testing/fstest"
)

var benchTemplates = fstest.MapFS{
	"simple.html": &fstest.MapFile{
		Data: []byte("Hello, <%= name %>!"),
	},
	"with_data.html": &fstest.MapFile{
		Data: []byte("Name: <%= name %>, Age: <%= age %>, Active: <%= active %>"),
	},
	"app/layouts/application.html": &fstest.MapFile{
		Data: []byte(`<!DOCTYPE html>
<html>
<head>
    <title><%= title %></title>
</head>
<body>
    <div class="container">
        <%= yield %>
    </div>
</body>
</html>`),
	},
	"complex.html": &fstest.MapFile{
		Data: []byte(`
<div class="user-profile">
    <h1><%= user.name %></h1>
    <p>Email: <%= user.email %></p>
    <p>Joined: <%= user.created_at %></p>
    <% if (user.active) { %>
        <span class="status active">Active User</span>
    <% } else { %>
        <span class="status inactive">Inactive User</span>
    <% } %>
    <ul>
    <% for (item in items) { %>
        <li><%= item.name %> - $<%= item.price %></li>
    <% } %>
    </ul>
</div>`),
	},
}

func BenchmarkEngineCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewEngine(benchTemplates)
	}
}

func BenchmarkEngineHTML(b *testing.B) {
	engine := NewEngine(benchTemplates)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		engine.HTML(&buf)
	}
}

func BenchmarkSimpleRender(b *testing.B) {
	engine := NewEngine(benchTemplates)
	data := map[string]any{"name": "World"}
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		engine.RenderHTML("simple.html", data)
	}
}

func BenchmarkComplexRender(b *testing.B) {
	engine := NewEngine(benchTemplates)
	data := map[string]any{
		"user": map[string]any{
			"name":       "John Doe",
			"email":      "john@example.com", 
			"created_at": "2023-01-01",
			"active":     true,
		},
		"items": []map[string]any{
			{"name": "Product 1", "price": 19.99},
			{"name": "Product 2", "price": 29.99},
			{"name": "Product 3", "price": 39.99},
		},
	}
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		engine.RenderHTML("complex.html", data)
	}
}

func BenchmarkPageRenderWithLayout(b *testing.B) {
	engine := NewEngine(benchTemplates)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		page := engine.HTML(&buf)
		page.Set("name", "World")
		page.Set("title", "Test Page")
		page.Render("simple.html")
	}
}

func BenchmarkPageRenderClean(b *testing.B) {
	engine := NewEngine(benchTemplates)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		page := engine.HTML(&buf)
		page.Set("name", "World")
		page.RenderClean("simple.html")
	}
}

func BenchmarkEngineWithHelpers(b *testing.B) {
	engine := NewEngine(benchTemplates, WithHelpers(map[string]any{
		"uppercase": func(s string) string { return s },
		"lowercase": func(s string) string { return s },
		"multiply":  func(a, b int) int { return a * b },
	}))
	data := map[string]any{"name": "test"}
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		engine.RenderHTML("simple.html", data)
	}
}

func BenchmarkConcurrentRender(b *testing.B) {
	engine := NewEngine(benchTemplates)
	data := map[string]any{"name": "World"}
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			engine.RenderHTML("simple.html", data)
		}
	})
}

func BenchmarkSetAndGet(b *testing.B) {
	engine := NewEngine(benchTemplates)
	var buf bytes.Buffer
	page := engine.HTML(&buf)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		page.Set("key", "value")
		page.Value("key")
	}
}