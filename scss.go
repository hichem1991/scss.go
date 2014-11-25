package scss

/*
#cgo LDFLAGS: ./libsass/lib/libsass.a -lstdc++
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
	"unsafe"
)

// This is our internal context
type ctx struct {
	cb ImportCallback
}

type Import struct {
	Path   string
	Source string // leave blank if import should be passed to css
	//Map    string
}

type ImportCallback func(url string) []Import

//export go_import_cb
func go_import_cb(url_s *C.char, cookie unsafe.Pointer) unsafe.Pointer {
	// For some reason the url_s comes in quoted.
	unquoted_url := C.sass_string_unquote(url_s)
	defer C.free(unsafe.Pointer(unquoted_url))
	url := C.GoString(unquoted_url)

	context := (*ctx)(cookie)
	imports := context.cb(url)

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

func Compile(inputPath string, source string, cb ImportCallback) (string, error) {
	input_path_s := C.CString(inputPath)
	defer C.free(unsafe.Pointer(input_path_s))

	source_s := C.CString(source)
	defer C.free(unsafe.Pointer(source_s))

	cookie := unsafe.Pointer(&ctx{
		cb: cb,
	})

	data_context := C.new_context(input_path_s, source_s, cookie)
	defer C.sass_delete_data_context(data_context)

	context := C.sass_data_context_get_context(data_context)

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
