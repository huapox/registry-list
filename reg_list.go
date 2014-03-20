package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	registry_addr string
)

func main() {
	var (
		bind          = flag.String("b", "127.0.0.1", "address to bind on")
		port          = flag.String("p", "8080", "port to listen on")
		cpus          = flag.Int("c", 1, "CPUs to use")
		flAddr          = flag.String("registry", "localhost:5000", "address to prefix the `docker pull ...`")
		registry_path = "/tmp"
		err           error
	)
	flag.Parse()

  if len(*flAddr) > 0 && !strings.HasSuffix(*flAddr,"/") {
    registry_addr = *flAddr + "/"
  }
	runtime.GOMAXPROCS(*cpus)

	if flag.NArg() > 0 {
		if registry_path, err = filepath.Abs(flag.Args()[0]); err != nil {
			log.Fatal(err)
		}
	}
	addr := fmt.Sprintf("%s:%s", *bind, *port)
	http.Handle("/", ImageListMux{registry_path})
	log.Printf("serving image list from %q listening on %s ...", registry_path, addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

type ImageListMux struct {
	BaseDir string
}

func (ilm ImageListMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		repos, err := ilm.Repos()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Fprintln(w, "<html><body><ul>")
		for _, repo := range repos {
			name := filepath.Clean(filepath.Join(repo.Namespace, repo.Name))
			fmt.Fprintf(w, "<li><b>docker pull %s%s:%s</b> (hash %s))</li>", registry_addr, name, repo.Tags[0].Name, repo.Tags[0].HashID) // XXX
		}
		fmt.Fprintln(w, "</ul></body></html>")
	} else {
		msg := fmt.Sprintf("TODO: handle %s", r.URL.String())
		fmt.Fprintln(w, msg)
		log.Println(msg)
	}
	r.Body.Close()
}

func (ilm ImageListMux) Repos() ([]Repo, error) {
	repos := []Repo{}
	err := filepath.Walk(filepath.Join(ilm.BaseDir, "repositories"), func(path string, fi os.FileInfo, err error) error {
		if fi.Mode().IsRegular() && strings.HasPrefix(filepath.Base(path), "tag_") {
			r, err := NewRepoFromTagFile(path)
			if err != nil {
				return err
			}
			if !HasRepo(r, repos) {
				repos = append(repos, r)
			}
		}
		return nil
	})
	if err != nil {
		return []Repo{}, err
	}
	return repos, nil
}

func HasRepo(r Repo, repos []Repo) bool {
	return false // XXX
}

func NewRepoFromTagFile(path string) (Repo, error) {
	t, err := NewTag(path)
	if err != nil {
		return Repo{}, nil
	}
	chunks := strings.Split(filepath.Dir(path), "/")
	r := Repo{
		Namespace: chunks[len(chunks)-2],
		Name:      chunks[len(chunks)-1],
		Tags:      []Tag{t},
	}
	if r.Namespace == "library" {
		r.Namespace = ""
	}
	return r, nil
}

type Repo struct {
	Namespace, Name string
	Tags            []Tag
}

func NewTag(path string) (Tag, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return Tag{}, err
	}
	return Tag{
		Name:   strings.TrimPrefix(filepath.Base(path), "tag_"),
		HashID: string(buf),
	}, nil
}

type Tag struct {
	Name   string
	HashID string
}
