package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

func errCheck(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func main() {
	o := flag.String("o", "", "out dir path")
	i := flag.String("i", "", "build import path")
	f := flag.String("f", "", "out file name")
	v := flag.String("v", "", "out file version")
	d := flag.Bool("d", false, "delete and not build")
	flag.Parse()

	if *i == "" {
		fmt.Println("bulid import path is nil")
		os.Exit(1)
	}
	if *o == "" {
		*o = path.Dir(*i)
	}
	if *f == "" {
		ext := path.Ext(*i)
		*f = strings.TrimSuffix(path.Base(*i), ext)
	}
	if *v != "" {
		*v = "_" + *v
	}
	run(path.Clean(*o), path.Clean(*i), *f, *v, *d)
}

func run(o, i, f, v string, d bool) {
	for goos := range goosMap {
		for goarch := range goarchMap {
			if !disBuild(goos, goarch) {
				continue
			}
			if d {
				err := os.Remove(getOutPath(o, f, goos, goarch, v))
				errCheck(err)
			} else {
				toBuild(goos, goarch, i, o, f, v)
			}
		}
	}
}

func toBuild(goos, goarch, ipath, opath, fName, version string) {
	toSetEnv(goos, goarch)
	out := getOutPath(opath, fName, goos, goarch, version)
	cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", out, ipath)
	b, err := cmd.CombinedOutput()
	fmt.Println("build:", goos, goarch, string(b), err)
	err = os.Chmod(out, 0777)
	errCheck(err)
}

func toSetEnv(goos, goarch string) {
	err := os.Setenv("CGO_ENABLED", "0")
	errCheck(err)
	err = os.Setenv("GOOS", goos)
	errCheck(err)
	err = os.Setenv("GOARCH", goarch)
	errCheck(err)
}

func getOutPath(outPath, fName, goos, goarch, version string) string {
	str := fmt.Sprintf("%s_%s_%s%s", fName, goos, goarch, version)
	return path.Join(outPath, str+outMap[goos])
}

var outMap = map[string]string{
	"windows": ".exe",
}

var goosMap = map[string]string{
	"darwin":  "darwin",
	"linux":   "linux",
	"windows": "windows",
}

var goarchMap = map[string]string{
	"386":   "386",
	"amd64": "amd64",
	"arm":   "arm",
	"arm64": "arm64",
}

func disBuild(goos, goarch string) bool {
	if goos == "darwin" {
		if goarch == "386" || goarch == "arm" {
			return false
		}
	}

	return true
}
