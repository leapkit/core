---
title: Environment Variables
index: 2
---

Leapkit provides a handy tool to load environment variables from a `.env` file. This feature is useful when you want to load your environment variables from a file instead of setting them directly in your system, like when you're in development mode.

To use it you can create a `.env` file in the root of your project and add your environment variables in the following format:

```env
# .env
PORT=8080
```

Then, you can load the environment variables in your by doing an underscore import on your `main.go` file.

```go
// main.go
import _ "github.com/leapkit/core/envload"
```

This will load the environment variables from the `.env` file into your application.
