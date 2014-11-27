package scss

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestBasicImports(t *testing.T) {
	inputPath := "spec/spec/basic/14_imports/input.scss"
	check(t, inputPath)
}

func TestAll(t *testing.T) {
	matches, err := filepath.Glob("spec/spec/*/*/input.scss")
	if err != nil {
		t.Fatal("glob fail")
	}

	if len(matches) == 0 {
		t.Fatal("no spec matches")
	}

	for _, inputPath := range matches {
		println(inputPath)
		check(t, inputPath)
	}
}

func findPath(p string) string {
	if !fileExists(p) {
		p = p + ".scss"
		if !fileExists(p) {
			p = path.Join(path.Dir(p), "_"+path.Base(p))
			if !fileExists(p) {
				panic("can't find input path: " + p)
			}
		}
	}
	return p
}

func firstPathExists(paths []string) string {
	for _, p := range paths {
		//println(i, p)
		if fileExists(p) {
			return p
		}
	}
	panic("couldn't find any!")
	return ""
}

func check(t *testing.T, inputPath string) {
	expectedOutputPath := expectedOutputPath(inputPath)
	if !fileExists(expectedOutputPath) {
		t.Fatalf("output file doesn't exist: %s", expectedOutputPath)
	}
	source, err := readAll(inputPath)

	if err != nil {
		t.Fatal(err)
	}

	inputPathDir := path.Dir(inputPath)

	output, err := Compile(inputPath, string(source), func(url string) []Import {
		imports := make([]Import, 1)

		paths := PossiblePaths(path.Join(inputPathDir, url))

		if paths == nil {
			imports[0].Path = url
			return imports
		}

		p := firstPathExists(paths)
		imports[0].Path = p

		source_bytes, err := readAll(p)
		if err != nil {
			panic(err)
		}

		imports[0].Source = string(source_bytes)
		return imports
	})

	if err != nil {
		t.Fatal(err)
	}

	expectedOutputBytes, err := readAll(expectedOutputPath)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := string(expectedOutputBytes)

	if strings.TrimSpace(expectedOutput) != strings.TrimSpace(output) {
		println("inputPathDir", inputPathDir)
		println("--------------------output------")
		print(output)
		println("--------------------expected----")
		print(expectedOutput)
		println("--------------------------------")

		t.Fatal("expected output does not match output for " + inputPath)
	}
}

func fileExists(filename string) bool {
	fi, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	if err != nil {
		panic(err)
	}

	if fi.IsDir() {
		return false
	}

	return true
}

func readAll(fn string) ([]byte, error) {
	file, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}

func expectedOutputPath(inputPath string) string {
	return path.Join(path.Dir(inputPath), "expected_output.css")
}
