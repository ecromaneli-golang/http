# WebServer

A lightweight and flexible HTTP router for Go, designed to simplify web development with dynamic routing, parameter handling, and intuitive APIs.

[![Go Reference](https://pkg.go.dev/badge/github.com/ecromaneli-golang/http.svg)](https://pkg.go.dev/github.com/ecromaneli-golang/http)
[![Go Report Card](https://goreportcard.com/badge/github.com/ecromaneli-golang/http)](https://goreportcard.com/report/github.com/ecromaneli-golang/http)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Introduction

WebServer is a Go-based HTTP router that provides dynamic URL pattern matching, unified parameter collection, and a clean API for building web applications. It supports features like wildcard routing, optional parameters, and server-sent events (SSE), making it a powerful tool for creating modern web services.

## Features

- **Dynamic Routing**: Supports wildcards (`*`, `**`), named parameters (`{name}`), and optional parameters (`{name?}`).
- **Unified Parameter Handling**: Access path, query, and body parameters seamlessly.
- **Server-Sent Events (SSE)**: Built-in support for real-time updates.
- **Static File Serving**: Serve static files with ease.
- **Fluent Response API**: Chainable methods for setting headers, status codes, and writing responses.

## Installation

To install the library, use the following command:

```bash
go get github.com/ecromaneli-golang/http
```

## How to Use

### Creating a Server

```go
package main

import (
    "github.com/ecromaneli-golang/http/webserver"
)

func main() {
    // Create a new server instance
    server := webserver.NewServer()

    // Add routes
    server.Get("/hello/{name}", func(req *webserver.Request, res *webserver.Response) {
        name := req.Param("name")
        res.WriteText("Hello, " + name + "!")
    })

    // Start the server
    server.ListenAndServe(":8080")
}
```
## Examples

### Example 1: Basic Routing

```go
server.Get("/greet/{name}", func(req *webserver.Request, res *webserver.Response) {
    name := req.Param("name")
    res.WriteText("Hello, " + name + "!")
})
```

### Example 2: Handling Parameters

```go
server.Post("/submit", func(req *webserver.Request, res *webserver.Response) {
    name := req.Param("name")
    age := req.IntParam("age")
    res.WriteText("Received name: " + name + ", age: " + strconv.Itoa(age))
})
```

### Example 3: Server-Sent Events (SSE)

```go
server.Get("/events", func(req *webserver.Request, res *webserver.Response) {
    res.Headers(webserver.EventStreamHeader)
    event := &webserver.Event{
        Name: "update",
        Data: map[string]string{"message": "Hello, SSE!"},
    }
    res.FlushEvent(event)
})
```

## Request API

The `Request` object provides a unified interface for accessing HTTP request data, including headers, parameters, and body content.

### Methods

#### Headers

- **`Header(name string) string`**  
  Returns the first value of the specified header.  
  Example:
  ```go
  userAgent := req.Header("User-Agent")
  ```

- **`Headers(name string) []string`**  
  Returns all values of the specified header.  
  Example:
  ```go
  cookies := req.Headers("Cookie")
  ```

- **`AllHeaders() http.Header`**  
  Returns all headers as a map.  
  Example:
  ```go
  headers := req.AllHeaders()
  ```

#### Parameters

- **`Param(name string) string`**  
  Returns the first value of the specified parameter (path, query, or body).  
  Example:
  ```go
  id := req.Param("id")
  ```

- **`Params(name string) []string`**  
  Returns all values of the specified parameter.  
  Example:
  ```go
  tags := req.Params("tags")
  ```

- **`AllParams() map[string][]string`**  
  Returns all parameters as a map.  
  Example:
  ```go
  params := req.AllParams()
  ```

- **`UIntParam(name string) uint`**  
  Converts the parameter value to an unsigned integer.  
  Example:
  ```go
  age := req.UIntParam("age")
  ```

- **`IntParam(name string) int`**  
  Converts the parameter value to an integer.  
  Example:
  ```go
  count := req.IntParam("count")
  ```

- **`Float64Param(name string) float64`**  
  Converts the parameter value to a 64-bit floating-point number.  
  Example:
  ```go
  price := req.Float64Param("price")
  ```

#### Body

- **`Body() []byte`**  
  Returns the raw body of the request.  
  Example:
  ```go
  body := req.Body()
  ```

#### Files

- **`File(name string) *multipart.FileHeader`**  
  Returns the first uploaded file for the specified form field.  
  Example:
  ```go
  file := req.File("profilePicture")
  ```

- **`Files(name string) []*multipart.FileHeader`**  
  Returns all uploaded files for the specified form field.  
  Example:
  ```go
  files := req.Files("attachments")
  ```

- **`AllFiles() map[string][]*multipart.FileHeader`**  
  Returns all uploaded files as a map.  
  Example:
  ```go
  allFiles := req.AllFiles()
  ```

#### Other

- **`IsDone() bool`**  
  Checks if the request has been completed or canceled.  
  Example:
  ```go
  if req.IsDone() {
      return
  }
  ```

## Response API

The `Response` object provides a fluent interface for constructing HTTP responses.

### Methods

#### Headers

- **`Header(key, value string) *Response`**  
  Adds a header to the response.  
  Example:
  ```go
  res.Header("Content-Type", "application/json")
  ```

- **`Headers(headers map[string][]string) *Response`**  
  Adds multiple headers to the response.  
  Example:
  ```go
  res.Headers(map[string][]string{
      "X-Custom-Header": {"Value1", "Value2"},
  })
  ```

#### Status

- **`Status(status int) *Response`**  
  Sets the HTTP status code for the response.  
  Example:
  ```go
  res.Status(201)
  ```

#### Writing Content

- **`Write(data []byte)`**  
  Writes binary data to the response.  
  Example:
  ```go
  res.Write([]byte("Hello, world!"))
  ```

- **`WriteText(text string)`**  
  Writes plain text to the response.  
  Example:
  ```go
  res.WriteText("Hello, world!")
  ```

- **`WriteJSON(value any)`**  
  Serializes the value to JSON and writes it to the response.  
  Example:
  ```go
  res.WriteJSON(map[string]string{"message": "Success"})
  ```

#### Server-Sent Events (SSE)

- **`FlushEvent(event *Event) error`**  
  Sends a server-sent event and flushes it immediately.  
  Example:
  ```go
  event := &webserver.Event{
      Name: "update",
      Data: map[string]string{"message": "Hello, SSE!"},
  }
  res.FlushEvent(event)
  ```

- **`FlushText(text string) error`**  
  Writes text to the response and flushes it immediately.  
  Example:
  ```go
  res.FlushText("Real-time update")
  ```

#### Rendering Files

- **`Render(filePath string)`**  
  Reads a file from the file system and writes it to the response.  
  Example:
  ```go
  res.Render("templates/index.html")
  ```

#### Other

- **`NoBody()`**  
  Writes an empty response.  
  Example:
  ```go
  res.Status(204).NoBody()
  ```

## Author

Created and maintained by [ecromaneli-golang](https://github.com/ecromaneli-golang).

## License

This project is licensed under the MIT License. See the [`LICENSE`](LICENSE) file for details.

Feel free to contribute to this project by submitting issues or pull requests!
