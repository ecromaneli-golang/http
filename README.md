# http/WebServer

## Basic Usage
```go
    import "github.com/ecromaneli-golang/http/webserver"
    
    server := webserver.NewServer(":80")

    server.Get("/example/{id}/**", func(req *webserver.Request, res *webserver.Response) {
        
        strId := req.Param("id")
        uintId := req.UIntParam("id")

        res.Status(200).WriteText("example")
    })

    err := server.ListenAndServe()
```