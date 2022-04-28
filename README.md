# http/WebServer
An easy-to-use Router. The main objective here is to have what you want right when you want it. Just that.

No benchmarks done yet.

## Basic Usage
```go
    import "github.com/ecromaneli-golang/http/webserver"
    
    server := webserver.NewServer()

    server.Get("/example/{id}/**", func(req *webserver.Request, res *webserver.Response) {
        
        strId := req.Param("id")
        uintId := req.UIntParam("id")

        res.Status(200).WriteText("example")
    })

    err := server.ListenAndServe(":80")
```

## Features

- Filter path dynamically (wildcard, accept-all, variables, optional variables, ...);
- Filter host dynamically (same rules as path);
- Parameter collector (host, path, query and body parameters all in the same place);
- Easy-to-use response writer;

# Creating a Server and Routing

How can I create a Server?
```golang
server := webserver.NewServer()
```

How to route? The Server can be used to route directly like this:
```golang
server.Handle(method, pattern, handler)
server.MultiHandle([]methods, pattern, handler)

// For any method
server.All(pattern, handler)

// The four famous
server.Get(pattern, handler)
server.Post(pattern, handler)
server.Put(pattern, handler)
server.Delete(pattern, handler)
```

How to [listen and] serve?
```golang
    server.ListenAndServe(addr)
```

Can I render a file or create a file server?
```golang
// It may change a lot, still working on making this easy

// Create a server with a file-system pointing to FS root path
server := webserver.NewServerWithFS(fileSystem)

// then
server.Get("/", func(req *webserver.Request, res *webserver.Response) {
    res.render("path/to/file")
})

// or to serve a file server pointing to the root path of the file system passed, do:

server.FileServer("/")

// Note that the '/' here is not the file system path, is the URL path.
```

Can I listen UDP? Not yet. But we have plans to.

# Routing URLs

The WebServer implements a set of special patterns to be able to handle paths dynamically:

- `*` any;
- `**` accepts everything ahead;
- `{name}` variable;
- `{name?}` optional variable;

Note that the WebServer also matches the host (without port), so everything before the first slash will be recognized as host pattern. The host pattern allows the same set of special patterns then path. The only difference is that the host is compared from RTL with the path is from LTR.

Also, slash as the final character of the path has no real effect.

Example:

```golang

    server.Get("{subdomain}.github.com/example/{id}/**", ...

    // will match with

    www.github.com/example/1/a/b/c/d
    subdomain.github.com/example/2/a/b
    hash.github.com/example/value
    hash.github.com/example/value/

    // will NOT match with

    github.com/example/1 // subdomain is required
    www.github.com/example/ // id is required

```

# Handler

Handler is a function that provides our modified Request and Response to make things easy. Just like this:

```golang
    func(req *webserver.Request, res *webserver.Response) {}
```

Next question...

# Request

The `Request` was made to make my projects easier, and I hope that yours too.

To get a Header, just use `Header` functions, we have a lot, no news here.

All parameters be host, path, query, body (formencoded) is provided by a single function called `.Param(name)`. You can also perform a automated conversion using `.UIntParam()`, `.FloatParam()` and ... The body is accessible by using the `.Body()` that reads the body Reader. 

All these functions just read the original request buffers when called to avoid some unecessary performance problems. But, of course, the project have a long way to be called "performance friendly".

You allways can access the original request by using the `Raw` attribute:
```golang
    req.Raw *http.Request
```

You can always call `req.IsDone()` to know if the request is still alive. The method does NOT return a channel.

# Response

Another ADT made to put a smile on my face when providing a response.

To set a Header, just use `Header` functions, we have a lot here too.

Here, the name of the functions talk for yourselves (I'm lazy, I want to go back to program).

I will document this better later.

```golang
    .Status(statusCode)
    .Write([]byte)
    .WriteText(string)
    .WriteJSON(any)
    .FlushEvent(*webserver.Event) // yes! SSE just don't die.
    .Render("path/to/file")
```

You can alsos access the original writer by using `res.RawWriter` and the file server (if passed) using `res.RawFS`.

# License

MIT License

Copyright (c) 2022 ecromaneli-golang

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
