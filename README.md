
## Fork
This is a heavily modified fork of the [go.rice](https://github.com/GeertJohan/go.rice) tool by GeertJohan. We found the rice tool's use of go source parsing excessive for our needs so we developed a smaller and simpler version that would be easy to maintain and still use the useful features go.rice provided. This version simply creates a tool that will import the paths listed as 'boxes' and append them to the executable. 

## boxedRice


[![Build Status](https://travis-ci.org/bypasslane/boxedRice.png)](https://travis-ci.org/bypasslane/boxedRice)
[![Godoc](https://img.shields.io/badge/godoc-boxedRice-blue.svg?style=flat-square)](https://godoc.org/github.com/bypasslane/boxedRice)

boxedRice is a [Go](http://golang.org) package that makes working with resources such as html,js,css,images and templates very easy. During development `boxedRice` will load required files directly from disk. Upon deployment it is easy to add all resource files to a executable using the `boxedRice` tool, without changing the source code for your package. boxedRice provides several methods to add resources to a binary.

### What does it do?
The first thing boxedRice does is finding the correct absolute path for your resource files. Say you are executing go binary in your home directory, but your `html-files` are located in `$GOPATH/src/yourApplication/html-files`. `boxedRice` will lookup the correct path for that directory (relative to the location of yourApplication). The only thing you have to do is include the resources using `boxedRice.FindBox("html-files")`.

This only works when the source is available to the machine executing the binary. Which is always the case when the binary was installed with `go get` or `go install`. It might occur that you wish to simply provide a binary, without source. The `boxedRice` tool provides the `append` command to append items to the binary as a zip. `boxedRice` will detect the appended resources and load those, instead of looking up files from disk.

### Installation

Use `go get` to install the package the `boxedRice` tool.
```
go get github.com/bypasslane/boxedRice
go get github.com/bypasslane/boxedRice/boxedRice
```

### Package usage

Import the package: `import "github.com/bypasslane/boxedRice"`

**Serving a static content folder over HTTP with a boxedRice Box**
```go
http.Handle("/", http.FileServer(boxedRice.MustFindBox("http-files").HTTPBox()))
http.ListenAndServe(":8080", nil)
```

**Service a static content folder over HTTP at a non-root location**
```go
box := boxedRice.MustFindBox("cssfiles")
cssFileServer := http.StripPrefix("/css/", http.FileServer(box.HTTPBox()))
http.Handle("/css/", cssFileServer)
http.ListenAndServe(":8080", nil)
```

Note the *trailing slash* in `/css/` in both the call to
`http.StripPrefix` and `http.Handle`.

**Loading a template**
```go
// find a boxedRice.Box
templateBox, err := boxedRice.FindBox("example-templates")
if err != nil {
	log.Fatal(err)
}
// get file contents as string
templateString, err := templateBox.String("message.tmpl")
if err != nil {
	log.Fatal(err)
}
// parse and execute the template
tmplMessage, err := template.New("message").Parse(templateString)
if err != nil {
	log.Fatal(err)
}
tmplMessage.Execute(os.Stdout, map[string]string{"Message": "Hello, world!"})

```

Never call `FindBox()` or `MustFindBox()` from an `init()` function, as the boxes might have not been loaded at that time.

### Tool usage
The `boxedRice` tool lets you add the resources to a binary executable so the files are not loaded from the filesystem anymore. This creates a 'standalone' executable. 

#### append
**Append resources to executable as zip file**

This method changes an already built executable. It appends the resources as zip file to the binary. It makes compilation a lot faster and can be used with large resource files.

Downsides for appending are that it requires `zip` to be installed and does not provide a working Seek method.

Run the following commands to create a standalone executable.
```
go build -o example
boxedRice append -b example/example-files --exec example
```

**Note: requires zip command to be installed**

On windows, install zip from http://gnuwin32.sourceforge.net/packages/zip.htm or cygwin/msys toolsets.

#### Help information
Run `boxedRice -h` for information about all options.

You can run the -h option for each sub-command, e.g. `boxedRice append -h`.

### Order of precedence
When opening a new box, the boxedRice package tries to locate the resources in the following order:

 - appended as zip
 - 'live' from filesystem


### License
This project is licensed under a Simplified BSD license. Please read the [LICENSE file][license].

### TODO & Development
This package is not completed yet. Though it already provides working appending, there are some refinements to be made. These are noted in code as `TODO` comments.

### Package documentation

You will find package documentation at [godoc.org/github.com/bypasslane/boxedRice][godoc].

 [license]: https://github.com/bypasslane/boxedRice/blob/master/LICENSE
 [godoc]: http://godoc.org/github.com/bypasslane/boxedRice
