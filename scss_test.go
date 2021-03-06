package scss

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestBootstrap(t *testing.T) {
	inputPath := "bootstrap-sass/stylesheets/_bootstrap.scss"
	source, err := readAll(inputPath)
	if err != nil {
		t.Fatal(err)
	}

	loader := TestLoader{
		Dir: path.Dir(inputPath),
	}
	output, err := Compile(inputPath, string(source), false, loader)

	if err != nil {
		t.Fatal(err)
	}

	println(output)
}

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
		if strings.HasPrefix(inputPath, "spec/spec/libsass-todo") {
			continue
		}

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
		println(p)
		if fileExists(p) {
			return p
		}
	}
	panic("couldn't find any!")
	return ""
}

type TestLoader struct {
	Dir string
}

func (l TestLoader) Load(parentPath string, importedPath string) (out Import) {
	println("parentPath", parentPath)
	println("importedPath", importedPath)

	var absImportedPath string
	if path.IsAbs(importedPath) {
		absImportedPath = importedPath
	} else {
		parentDir := path.Dir(parentPath)
		absImportedPath = path.Join(parentDir, importedPath)
	}

	paths := PossiblePaths(absImportedPath)

	if paths == nil {
		// return error?
		out.Path = importedPath
		return out
	}

	p := firstPathExists(paths)
	out.Path = p

	source_bytes, err := readAll(p)
	if err != nil {
		panic(err)
	}

	out.Source = string(source_bytes)
	return out
}

// The ruby spec that we're using to test uses the following function to
// "clean" the output. So we want to reproduce its behavior exactly.
//
//  def _clean_output(css)
//    css.gsub(/\s+/, " ")
//       .gsub(/ *\{/, " {\n")
//       .gsub(/([;,]) */, "\\1\n")
//       .gsub(/ *\} */, " }\n")
//       .strip
//  end
func cleanOutput(css string) string {
	r1, _ := regexp.Compile(`\s+`)
	css = r1.ReplaceAllString(css, " ")

	r2, _ := regexp.Compile(` *\{`)
	css = r2.ReplaceAllString(css, " {\n")

	/* uh.. how do i do pattern match replace? too lazy to figure out right now.
	r3, _ := regexp.Compile(`([;,]) *`)
	css = r3.ReplaceAllString(css, "\\1\n")
	*/

	r4, _ := regexp.Compile(` *\} *`)
	css = r4.ReplaceAllString(css, " }\n")

	return strings.TrimSpace(css)
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

	loader := TestLoader{
		Dir: path.Dir(inputPath),
	}
	output, err := Compile(inputPath, string(source), false, loader)

	if err != nil {
		t.Fatal(err)
	}

	expectedOutputBytes, err := readAll(expectedOutputPath)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := string(expectedOutputBytes)

	expectedOutput = cleanOutput(expectedOutput)
	output = cleanOutput(output)

	if output != expectedOutput {
		println("loader.Dir", loader.Dir)
		println("--------------------output------")
		println(output)
		println("--------------------expected----")
		println(expectedOutput)
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
