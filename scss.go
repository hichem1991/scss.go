package scss

/*
#cgo pkg-config: scss.pc

#include <sass.h>
#include <sass_context.h>
#include <stdlib.h>

// Defined in import_cb.c
struct Sass_Data_Context* new_context(char* input_path, char* source, int compress, void* cookie);
*/
import "C"
import (
	"errors"
	"math/rand"
	"path"
	"unsafe"
)

type Import struct {
	Path   string
	Source string // leave blank if import should be passed to css
	//Map    string
}

type Loader interface {
	Load(parentPath string, importPath string) Import
}

// This is our internal context
type internalContext struct {
	name     string
	loader   Loader
	cContext *C.struct_Sass_Context
}

//export go_import_cb
func go_import_cb(parentPath_s *C.char, importPath_s *C.char,
	cookie unsafe.Pointer) **C.struct_Sass_Import {
	iContext := (*internalContext)(cookie)

	// For some reason the importPath_s comes in quoted.
	unquoted_importPath := C.sass_string_unquote(importPath_s)
	defer C.free(unsafe.Pointer(unquoted_importPath))
	importPath := C.GoString(unquoted_importPath)

	unquoted_parentPath := C.sass_string_unquote(parentPath_s)
	defer C.free(unsafe.Pointer(unquoted_parentPath))
	parentPath := C.GoString(unquoted_parentPath)

	/*
		options := C.sass_context_get_options(iContext.cContext)
		outputPath_s := C.sass_option_get_output_path(options)
		outputPath := C.GoString(outputPath_s)
		println("outputPath", outputPath)
	*/

	//println(">>>>>>> before", parentPath, importPath)
	//println(">>>>>>> loader:", string(iContext.name))
	import_ := iContext.loader.Load(parentPath, importPath)
	//println("<<<<<<< after", parentPath, importPath)

	// Copy The golang []Import object into something sass understands.
	c_imports := C.sass_make_import_list(C.size_t(1))

	path_s := C.CString(import_.Path)

	// This is so source_s will be NULL. which triggers direct imports???
	var source_s *C.char
	if len(import_.Source) > 0 {
		source_s = C.CString(import_.Source)
	}

	entry := C.sass_make_import_entry(path_s, source_s, nil)
	C.sass_import_set_list_entry(c_imports, 0, entry)

	// Who owns what? sass has a shitty API
	C.free(unsafe.Pointer(path_s))
	//C.free(unsafe.Pointer(source_s))

	return c_imports
}

// Returns scss files that this could refer to.
// this issues no syscalls.
func PossiblePaths(p string) (out []string) {
	ext := path.Ext(p)

	if ext == ".css" {
		return nil
	} else if ext == "" {
		out = make([]string, 2)
		out[0] = path.Join(path.Dir(p), "_"+path.Base(p)+".scss")
		out[1] = p + ".scss"
	} else if ext == ".scss" {
		out = make([]string, 1)
		out[0] = p
	} else {
		panic("uhh")
	}

	return out
}

var sassContextMap = make(map[int32]*internalContext)

func Compile(inputPath string, source string, compress bool, loader Loader) (string, error) {
	input_path_s := C.CString(inputPath)
	defer C.free(unsafe.Pointer(input_path_s))

	source_s := C.CString(source)
	defer C.free(unsafe.Pointer(source_s))

	// This is to work around a crash I was hitting where
	// the internalContext would be GCed (I think) while we are
	// inside of sass_compile_data_context, despite the object
	// being rooted in this scope.
	// It seems attaching it to this global map fixes the bug?
	sassContextId := rand.Int31()
	sassContextMap[sassContextId] = &internalContext{
		name:   inputPath,
		loader: loader,
	}
	defer delete(sassContextMap, sassContextId)

	iContext := sassContextMap[sassContextId]

	cookie := unsafe.Pointer(iContext)

	var compress_c C.int
	if compress {
		compress_c = 1
	} else {
		compress_c = 0
	}

	data_context := C.new_context(input_path_s, source_s, compress_c, cookie)
	defer C.sass_delete_data_context(data_context)

	context := C.sass_data_context_get_context(data_context)

	iContext.cContext = context

	C.sass_compile_data_context(data_context)

	status := C.sass_context_get_error_status(context)

	if status == 0 {
		output_s := C.sass_context_get_output_string(context)
		output := C.GoString(output_s)
		return output, nil
	} else {
		// error
		/*
			error_json_s := C.sass_context_get_error_json(context)
			error_json := C.GoString(error_json_s)
			//println("error json", error_json)
		*/

		error_message_s := C.sass_context_get_error_message(context)
		error_message := C.GoString(error_message_s)
		//println("error message", error_message)

		return "", errors.New(error_message)
	}

}
