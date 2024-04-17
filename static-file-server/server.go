package main

import (
	"fmt"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

/// MIME
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types

const usage = `
static-file-server <port> <path>
	Start a server on :<port> that caches and serves files from <path> and its subdirectories. 
	Only files with mime type recognized by https://pkg.go.dev/mime are served. Other file extensions will be ignored.`

const timeout = 5000
const maxHeaderLength = 4096
const indexString = "index.html"

// TODO: extension filter
// TODO: - consider files might be mis"mimed"
// TODO: - consider you might want to transfer a file as a different mime type
// TODO: - consider removing mime package ( or keep as default) and allow loading a json map

func main() {
	if len(os.Args) < 3 {
		usageAndExit(fmt.Sprintf("error: invalid number of arguments: (%v)", os.Args[1:]))
	}
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		usageAndExit("error: port is not an integer")
	}

	abs_path, err := filepath.Abs(os.Args[2])
	if err != nil {
		log.Printf("%v", err)
		usageAndExit("error: could not create abolute path")
	}

	filter := fileFilter{make([]cacheableFile, 0), 0, 0, 0}

	dir := os.DirFS(abs_path)
	err = fs.WalkDir(dir, ".", filter.filterAndCache) // use fs.WalkDir to avoid symbolic links
	if err != nil {
		log.Printf("%v", err)
		usageAndExit("error: failed to walk <path>")
	}
	log.Printf("FilesCacheable: %v, FilesUnknownMime: %v, Directories: %v", len(filter.files), filter.filesUnknownMime, filter.directoriesFound)

	cache := make(map[string]cacheableFile)
	for _, v := range filter.files {
		data, err := os.ReadFile(abs_path + v.path)
		if err != nil {
			log.Printf("error: could not cache %v", v.path)
		} else {
			v.data = data
			cache[v.path] = v
			log.Printf("cached %v", v.path)

			// Special case for file indexString
			base := filepath.Base(v.path)
			if base == indexString {
				dir := filepath.Dir(v.path)
				// indexString available on /<path>/   (not for path="/", root //)
				if dir != "/" {
					cache[dir+"/"] = v
					log.Printf("cached %v/ for %v", dir, indexString)
				}
				// indexString available on /<path>
				cache[dir] = v
				log.Printf("cached %v for %v", dir, indexString)
			}
		}
	}

	log.Println("Starting static file server...")

	server := http.Server{
		Addr:           fmt.Sprintf(":%v", port),
		Handler:        http.MaxBytesHandler(cacheHandler{cache}, maxHeaderLength),
		ReadTimeout:    timeout * time.Millisecond,
		WriteTimeout:   timeout * time.Millisecond,
		MaxHeaderBytes: maxHeaderLength,
	}
	log.Printf("Listening on :%v", port)
	err = server.ListenAndServe()
	log.Printf("Server return: %v", err)
}

type cacheHandler struct {
	cache map[string]cacheableFile
}

func (c cacheHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	url := req.URL.Path
	if url[0] != '/' {
		url = "/" + url
	}
	file, ok := c.cache[url]
	if !ok {
		log.Printf("refused: %v requested %v", req.RemoteAddr, req.URL)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	if req.Method != http.MethodGet {
		log.Printf("refused: %v %v %v", req.RemoteAddr, req.Method, req.URL)
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	log.Printf("accepted: %v requested %v", req.RemoteAddr, url)
	res.Header().Set("Content-Type", file.mime)
	res.Write(file.data)
}

type cacheableFile struct {
	path string
	mime string
	data []byte
}
type fileFilter struct {
	files            []cacheableFile
	directoriesFound int
	filesFound       int
	filesUnknownMime int
}

func (f *fileFilter) filterAndCache(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if d.Type().IsRegular() {
		f.filesFound++
		f.filter(path)
	} else {
		f.directoriesFound++
	}
	return nil
}

func (f *fileFilter) filter(path string) {
	mimeType := mime.TypeByExtension(filepath.Ext(path))
	if mimeType != "" {
		f.files = append(f.files, cacheableFile{"/" + path, mimeType, nil})
	} else {
		f.filesUnknownMime++
	}
}

func usageAndExit(errorMessage string) {
	log.Println(errorMessage)
	log.Println(usage)
	os.Exit(1)
}
