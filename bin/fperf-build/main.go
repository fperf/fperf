package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

const filename = "/tmp/fperf_main.go"

type option struct {
	output  string
	verbose bool
}

func gobuild(o *option, imports []string) error {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("package main\n")
	buf.WriteString(`import "github.com/fperf/fperf"` + "\n")
	for _, imp := range imports {
		buf.WriteString(`import _ "` + imp + `"` + "\n")
	}
	buf.WriteString(`
	func main() {
		fperf.Main()
	}
	`)

	if err := ioutil.WriteFile(filename, buf.Bytes(), 0655); err != nil {
		log.Fatalln(err)
	}
	defer os.Remove(filename)

	args := []string{"build", "-o", o.output, filename}
	if o.verbose {
		args = append(args, "-v")
	}
	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func main() {
	o := &option{}
	flag.StringVar(&o.output, "o", "fperf", "build output")
	flag.BoolVar(&o.verbose, "v", false, "print the names of packages as they are compiled.")
	flag.Parse()

	paths := flag.Args()
	if len(paths) == 0 {
		flag.Usage()
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	if len(paths) == 0 {
		paths = append(paths, cwd)
	}

	imports := make([]string, len(paths))
	for i, p := range paths {
		cfg := packages.Config{}
		if strings.HasPrefix(p, "./") {
			cfg.Dir, _ = filepath.Abs(p)
			pkgs, err := packages.Load(&cfg, "./")
			if err != nil {
				log.Fatalln(err)
			}
			imports[i] = pkgs[i].String()
		} else {
			imports[i] = p
		}
	}
	fmt.Println(paths, imports)
	if err := gobuild(o, imports); err != nil {
		log.Fatalln(err)
	}
}
