# GoWebServer
This is an HTTP* server, which does basic serving of webpages and has a couple default error codes.

(*[#1](https://github.com/krestenlaust/GoWebServer/issues/1) SSL support)

## How to use it?
Following is a short description of how to get the server running as a user.

### Normal executable

1. Get recent release / compile a folder
2. Place executable in your WWW-folder
3. Run executable

The default handler is [file-handler](https://github.com/krestenlaust/GoWebServer/blob/67e3063ce05c471fd61a069358499c0679aadcb5/request_handers.go#L35) (currently, handler is defined at compile-time)
- It searches it's own directory for a file by URL-path
- If the URL doesn't contain path, it searches for the file "index.html"
- It does Not support displaying folders

### Docker

1. Place WWW-files inside source directory (as of writing, it's the root directory)
2. Compile docker image

## Project structure
This is one of my first projects written in Go, hence I'm not used to the way that files don't simply contain a single class-level type, but instead are grouped by topic with multiple top-level types inside. Following is a short description of the most relevant files.

### http_handling.go
This file contains the main logic surrounding the implementation of http.

Development: One of the main considerations I've been having, is whether to bubble any issues encountered, like a bad request or a server side issue, to another file. Currently, when an error is encountered, it is simply logged to the console, with default log method.

### request_handlers.go
This file contains implementations of the request handler type. The default request handler, "file handler", is a mediator between the networking request being received, and what is returned from the file system.

Development: The only considerations I have for this file, is to make the file handler more generic, right now, it has hardcoded values. There's also some concerns regarding bubbling of errors. Since some errors are only recognized later and are dependent on handler type.

### networking.go
The file contains logic that mediates between the established socket and the handler/parser.

Development: There's a todo comment.

### main.go
Entrypoint of the project, this is where basic configuration is made and the selected handler is instantiated before being passed.

## Contributing
Contributions in any shape are more than welcome. Whether isn't a typo or an entire new feature.