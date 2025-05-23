package render

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing/fstest"
)

func ExampleNewEngine() {
	// Create a simple filesystem for templates
	templates := fstest.MapFS{
		"hello.html": &fstest.MapFile{
			Data: []byte("Hello, <%= name %>!"),
		},
	}

	// Create a new render engine
	engine := NewEngine(templates)

	// Render a template to string
	result, err := engine.RenderHTML("hello.html", map[string]any{
		"name": "World",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
	// Output: Hello, World!
}

func ExampleEngine_HTML() {
	templates := fstest.MapFS{
		"greeting.html": &fstest.MapFile{
			Data: []byte("Greetings, <%= user %>!"),
		},
		"app/layouts/application.html": &fstest.MapFile{
			Data: []byte(`<html><body><%= yield %></body></html>`),
		},
	}

	engine := NewEngine(templates)
	var buf bytes.Buffer
	page := engine.HTML(&buf)

	page.Set("user", "Alice")
	err := page.Render("greeting.html")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(buf.String())
	// Output: <html><body>Greetings, Alice!</body></html>
}

func ExampleWithDefaultLayout() {
	templates := fstest.MapFS{
		"page.html": &fstest.MapFile{
			Data: []byte("Content here"),
		},
		"layouts/custom.html": &fstest.MapFile{
			Data: []byte(`<div><%= yield %></div>`),
		},
	}

	// Create engine with custom default layout
	engine := NewEngine(templates, WithDefaultLayout("layouts/custom.html"))

	var buf bytes.Buffer
	page := engine.HTML(&buf)
	page.Render("page.html")

	fmt.Println(buf.String())
	// Output: <div>Content here</div>
}

func ExampleWithHelpers() {
	templates := fstest.MapFS{
		"text.html": &fstest.MapFile{
			Data: []byte("Result: <%= shout(message) %>"),
		},
	}

	// Create engine with helper functions
	engine := NewEngine(templates, WithHelpers(map[string]any{
		"shout": func(text string) string {
			return strings.ToUpper(text)
		},
	}))

	result, err := engine.RenderHTML("text.html", map[string]any{
		"message": "hello world",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
	// Output: Result: HELLO WORLD
}

func ExamplePage_RenderClean() {
	templates := fstest.MapFS{
		"snippet.html": &fstest.MapFile{
			Data: []byte("Name: <%= name %>, Age: <%= age %>"),
		},
	}

	engine := NewEngine(templates)
	var buf bytes.Buffer
	page := engine.HTML(&buf)

	page.Set("name", "Bob")
	page.Set("age", 30)

	// Render without any layout
	err := page.RenderClean("snippet.html")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(buf.String())
	// Output: Name: Bob, Age: 30
}

func ExamplePage_RenderWithLayout() {
	templates := fstest.MapFS{
		"content.html": &fstest.MapFile{
			Data: []byte("Main content"),
		},
		"special_layout.html": &fstest.MapFile{
			Data: []byte(`<article><%= yield %></article>`),
		},
	}

	engine := NewEngine(templates)
	var buf bytes.Buffer
	page := engine.HTML(&buf)

	// Render with a specific layout
	err := page.RenderWithLayout("content.html", "special_layout.html")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(buf.String())
	// Output: <article>Main content</article>
}

func ExampleTemplateFS() {
	// Create embedded templates
	embedTemplates := fstest.MapFS{
		"prod.html": &fstest.MapFile{
			Data: []byte("Production template"),
		},
	}

	// Create a template filesystem that can fallback between local and embedded
	fs := TemplateFS(embedTemplates, "/path/to/templates")

	// Use the filesystem with render engine
	engine := NewEngine(fs)
	
	result, err := engine.RenderHTML("prod.html", map[string]any{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
	// Output: Production template
}

func ExampleEngine_Set() {
	templates := fstest.MapFS{
		"welcome.html": &fstest.MapFile{
			Data: []byte("Welcome to <%= site_name %>!"),
		},
	}

	engine := NewEngine(templates)
	// Set global values available to all templates
	engine.Set("site_name", "My Website")

	result, err := engine.RenderHTML("welcome.html", map[string]any{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
	// Output: Welcome to My Website!
}

func ExampleEngine_SetHelper() {
	templates := fstest.MapFS{
		"math.html": &fstest.MapFile{
			Data: []byte("Result: <%= multiply(x, y) %>"),
		},
	}

	engine := NewEngine(templates)
	// Set a global helper function
	engine.SetHelper("multiply", func(a, b int) int {
		return a * b
	})

	result, err := engine.RenderHTML("math.html", map[string]any{
		"x": 6,
		"y": 7,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
	// Output: Result: 42
}