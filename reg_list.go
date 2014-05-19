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
    "time"
)

var (
	registry_addr string
    style_path string
    script_path string
)

func main() {
	var (
		bind          = flag.String("b", "127.0.0.1", "address to bind on")
		port          = flag.String("p", "8080", "port to listen on")
		cpus          = flag.Int("c", 1, "CPUs to use")
		flAddr          = flag.String("registry", "localhost:5000", "address to prefix the `docker pull ...`")
		registry_path = "/tmp"
        style         = flag.String("s", "./style.css", "path to style.css file")
        script        = flag.String("j", "./script.js", "path to script.js file")
		err           error
	)
	flag.Parse()

  if len(*flAddr) > 0 && !strings.HasSuffix(*flAddr,"/") {
    registry_addr = *flAddr + "/"
  }
    style_path = *style
    script_path = *script
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

        style_content, err := ioutil.ReadFile(style_path)
        if err != nil { panic(err) }
        script_content, err2 := ioutil.ReadFile(script_path)
        if err2 != nil { panic(err2) }

		fmt.Fprintln(w, "<html>");
        fmt.Fprintf(w, "<head><style>%s</style><script src='//ajax.googleapis.com/ajax/libs/jquery/1.7.1/jquery.min.js'></script><script>%s</script></head>", style_content, script_content);
        fmt.Fprintln(w, "<body><a href='#' id='show_all'>Show all</a><table>")
        last_namespace := ""
        first := true
        const layout = "Jan 2, 2006"
		for _, repo := range repos {
            if last_namespace != repo.Namespace {
                if first {
                    first = false
                    fmt.Fprintf(w, "</tbody>")
                }
                namespace := ""
                if repo.Namespace == "" {namespace = "No namespace"} else {namespace = repo.Namespace}
                fmt.Fprintf(w, "<thead class='repository_name'><tr><th colspan='3'>%s</th></tr></thead><tbody>", namespace)
                last_namespace = repo.Namespace

            }
			name := filepath.Clean(filepath.Join(repo.Namespace, repo.Name))
			fmt.Fprintf(w, "<tr><td class='pull_cmd'><b>docker pull %s%s:%s</b></td><td> (hash %s))</td><td>%s</td></tr>", registry_addr, name, repo.Tags[0].Name, repo.Tags[0].HashID, repo.Time.Format(layout)) // XXX
		}
		fmt.Fprintln(w, "</tbody></table></body></html>")
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
    info, err := os.Stat(path)
	r := Repo{
		Namespace: chunks[len(chunks)-2],
		Name:      chunks[len(chunks)-1],
		Tags:      []Tag{t},
        Time:      info.ModTime(),
	}
	if r.Namespace == "library" {
		r.Namespace = ""
	}
	return r, nil
}

type Repo struct {
	Namespace, Name string
	Tags            []Tag
    Time            time.Time
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
