package scss

/*
#cgo LDFLAGS: ./libsass/lib/libsass.a -lc++
#cgo CPPFLAGS: -I./libsass

#include <sass.h>
#include <sass_context.h>
#include <stdlib.h>

// Defined in import_cb.c
struct Sass_Data_Context* new_context(char* input_path, char* source, void* cookie);

*/
import "C"
import (
	"errors"
	"path"
	"unsafe"
)

type Import struct {
	Path   string
	Source string // leave blank if import should be passed to css
	//Map    string
}

type Loader interface {
	Load(path string) []Import
}

// This is our internal context
type internalContext struct {
	loader   Loader
	cContext *C.struct_Sass_Context
}

//export go_import_cb
func go_import_cb(url_s *C.char, cookie unsafe.Pointer) unsafe.Pointer {
	// For some reason the url_s comes in quoted.
	unquoted_url := C.sass_string_unquote(url_s)
	defer C.free(unsafe.Pointer(unquoted_url))
	url := C.GoString(unquoted_url)

	iContext := (*internalContext)(cookie)

	imports := iContext.loader.Load(url)

	// Copy The golang []Import object into something sass understands.
	c_imports := C.sass_make_import_list(C.size_t(len(imports)))
	for i, _ := range imports {
		path_s := C.CString(imports[i].Path)
		var source_s *C.char

		// This is so source_s will be NULL. which triggers direct imports???
		if len(imports[i].Source) > 0 {
			source_s = C.CString(imports[i].Source)
		}

		entry := C.sass_make_import_entry(url_s, source_s, nil)
		C.sass_import_set_list_entry(c_imports, C.size_t(i), entry)

		// Who owns what? sass has a shitty API
		C.free(unsafe.Pointer(path_s))
		//C.free(unsafe.Pointer(source_s))
	}

	return unsafe.Pointer(c_imports)
}

// Returns scss files that this could refer to.
// this issues no syscalls.
func PossiblePaths(p string) (out []string) {
	ext := path.Ext(p)

	if ext == ".css" {
		return nil
	}

	if ext == "" {
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

func Compile(inputPath string, source string, loader Loader) (string, error) {
	input_path_s := C.CString(inputPath)
	defer C.free(unsafe.Pointer(input_path_s))

	source_s := C.CString(source)
	defer C.free(unsafe.Pointer(source_s))

	iContext := &internalContext{
		loader: loader,
	}

	cookie := unsafe.Pointer(iContext)

	data_context := C.new_context(input_path_s, source_s, cookie)
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
