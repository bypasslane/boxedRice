package boxedRice

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Box abstracts a directory for resources/files.
// It can either load files from disk, or from appended files
type Box struct {
	name         string
	absolutePath string
	appendd      *appendedBox
}

var defaultLocateOrder = []LocateMethod{LocateAppended, LocateFS}

func findBox(name string, order []LocateMethod) (*Box, error) {
	b := &Box{name: name}

	// no support for absolute paths since gopath can be different on different machines.
	// therefore, required box must be located relative to package requiring it.
	if filepath.IsAbs(name) {
		return nil, errors.New("given name/path is absolute")
	}

	var err error
	for _, method := range order {
		switch method {

		case LocateAppended:
			appendedBoxName := strings.Replace(name, `/`, `-`, -1)
			if appendd := appendedBoxes[appendedBoxName]; appendd != nil {
				b.appendd = appendd
				return b, nil
			}

		case LocateFS:
			// resolve absolute directory path
			err := b.resolveAbsolutePathFromCaller()
			if err != nil {
				continue
			}
			// check if absolutePath exists on filesystem
			info, err := os.Stat(b.absolutePath)
			if err != nil {
				continue
			}
			// check if absolutePath is actually a directory
			if !info.IsDir() {
				err = errors.New("given name/path is not a directory")
				continue
			}
			return b, nil
		case LocateWorkingDirectory:
			// resolve absolute directory path
			err := b.resolveAbsolutePathFromWorkingDirectory()
			if err != nil {
				continue
			}
			// check if absolutePath exists on filesystem
			info, err := os.Stat(b.absolutePath)
			if err != nil {
				continue
			}
			// check if absolutePath is actually a directory
			if !info.IsDir() {
				err = errors.New("given name/path is not a directory")
				continue
			}
			return b, nil
		}
	}

	if err == nil {
		err = fmt.Errorf("could not locate box %q", name)
	}

	return nil, err
}

// FindBox returns a Box instance for given name.
// When the given name is a relative path, it's base path will be the calling pkg/cmd's source root.
// When the given name is absolute, it's absolute. derp.
// Make sure the path doesn't contain any sensitive information as it might be placed into generated go source (embedded).
func FindBox(name string) (*Box, error) {
	return findBox(name, defaultLocateOrder)
}

// MustFindBox returns a Box instance for given name, like FindBox does.
// It does not return an error, instead it panics when an error occurs.
func MustFindBox(name string) *Box {
	box, err := findBox(name, defaultLocateOrder)
	if err != nil {
		panic(err)
	}
	return box
}

// This is injected as a mutable function literal so that we can mock it out in
// tests and return a fixed test file.
var resolveAbsolutePathFromCaller = func(name string, nStackFrames int) (string, error) {
	_, callingGoFile, _, ok := runtime.Caller(nStackFrames)
	if !ok {
		return "", errors.New("couldn't find caller on stack")
	}

	// resolve to proper path
	pkgDir := filepath.Dir(callingGoFile)
	// fix for go cover
	const coverPath = "_test/_obj_test"
	if !filepath.IsAbs(pkgDir) {
		if i := strings.Index(pkgDir, coverPath); i >= 0 {
			pkgDir = pkgDir[:i] + pkgDir[i+len(coverPath):]            // remove coverPath
			pkgDir = filepath.Join(os.Getenv("GOPATH"), "src", pkgDir) // make absolute
		}
	}
	return filepath.Join(pkgDir, name), nil
}

func (b *Box) resolveAbsolutePathFromCaller() error {
	path, err := resolveAbsolutePathFromCaller(b.name, 4)
	if err != nil {
		return err
	}
	b.absolutePath = path
	return nil

}

func (b *Box) resolveAbsolutePathFromWorkingDirectory() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	b.absolutePath = filepath.Join(path, b.name)
	return nil
}

// IsAppended indicates weather this box was appended to the application
func (b *Box) IsAppended() bool {
	return b.appendd != nil
}

// Time returns how actual the box is.
// When the box is live, this methods returns time.Now()
func (b *Box) Time() time.Time {
	//++ TODO: return time for appended box

	return time.Now()
}

// Open opens a File from the box
// If there is an error, it will be of type *os.PathError.
func (b *Box) Open(name string) (*File, error) {
	if Debug {
		fmt.Printf("Open(%s)\n", name)
	}

	if b.IsAppended() {
		// trim prefix (paths are relative to box)
		name = strings.TrimLeft(name, "/")

		// search for file
		appendedFile := b.appendd.Files[name]
		if appendedFile == nil {
			return nil, &os.PathError{
				Op:   "open",
				Path: name,
				Err:  os.ErrNotExist,
			}
		}

		// create new file
		f := &File{
			appendedF: appendedFile,
		}

		// if this file is a directory, we want to be able to read and seek
		if !appendedFile.dir {
			// looks like malformed data in zip, error now
			if appendedFile.content == nil {
				return nil, &os.PathError{
					Op:   "open",
					Path: "name",
					Err:  errors.New("error reading data from zip file"),
				}
			}
			// create new bytes.Reader
			f.appendedFileReader = bytes.NewReader(appendedFile.content)
		}

		// all done
		return f, nil
	}

	// perform os open
	if Debug {
		fmt.Printf("Using os.Open(%s)", filepath.Join(b.absolutePath, name))
	}
	file, err := os.Open(filepath.Join(b.absolutePath, name))
	if err != nil {
		return nil, err
	}
	return &File{realF: file}, nil
}

// Bytes returns the content of the file with given name as []byte.
func (b *Box) Bytes(name string) ([]byte, error) {
	file, err := b.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// MustBytes returns the content of the file with given name as []byte.
// panic's on error.
func (b *Box) MustBytes(name string) []byte {
	bts, err := b.Bytes(name)
	if err != nil {
		panic(err)
	}
	return bts
}

// String returns the content of the file with given name as string.
func (b *Box) String(name string) (string, error) {
	bts, err := b.Bytes(name)
	if err != nil {
		return "", err
	}
	return string(bts), nil
}

// MustString returns the content of the file with given name as string.
// panic's on error.
func (b *Box) MustString(name string) string {
	str, err := b.String(name)
	if err != nil {
		panic(err)
	}
	return str
}

// Name returns the name of the box
func (b *Box) Name() string {
	return b.name
}
