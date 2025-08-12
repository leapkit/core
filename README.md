<img width="300" alt="logo" src="https://github.com/LeapKit/template/assets/645522/d5bcb8ed-c763-4b39-8cfb-aed694b87646">

## Leapkit Core
**Leapkit Core** is a Go web framework library that provides essential building blocks for building web applications. It's designed to be a lightweight, modular foundation with clean APIs and sensible defaults.

## 🏗️ Core Components

### Server Module (`server/`)
- HTTP server with routing capabilities
- Middleware support
- Session management
- Error handling with templating support
- Built on top of standard Go `net/http`

### Database Module (`db/`)
- Database connection management
- Support for SQLite and PostgreSQL
- Database migration system
- Database creation utilities

### Rendering Engine (`render/`)
- Template rendering system using Plush templating engine
- Support for layouts and partials
- File system-based template loading
- Context-aware rendering

### Form Handling (`form/`)
- Form data decoding
- Form validation system
- HTTP request form processing

### Tools (`tools/`)
- Environment variable loading utilities

## 🎯 Purpose

This library serves as a foundational component for the Leapkit ecosystem - a Go-based web framework that aims to provide:

- **Simplicity**: Clean, straightforward APIs for common web development tasks
- **Modularity**: Each component can be used independently
- **Convention over Configuration**: Sensible defaults (like `app/layouts/application.html` for layouts)
- **Database Agnostic**: Support for multiple database systems
- **Modern Go Practices**: Uses Go 1.22+ and follows contemporary Go patterns

## 📦 Dependencies

The project uses minimal, well-established dependencies:
- **Plush**: For templating (from Buffalo ecosystem)
- **Gorilla Sessions**: For session management
- **SQLite3**: For database support

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


![Alt](https://repobeats.axiom.co/api/embed/96fe663d186f3135ee411891075e366b731aaa16.svg "Repobeats analytics image")
