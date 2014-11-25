#include "_cgo_export.h"
#include <sass_context.h>
#include <stdlib.h>
#include <stdio.h>

struct Sass_Import** import_cb(const char* url, void* cookie) {
  void* import_entries_ptr = go_import_cb((char*)url, cookie);
  struct Sass_Import** imports = (struct Sass_Import**)(import_entries_ptr);
  return imports;
}

struct Sass_Data_Context* new_context(char* input_path, char* source, void* cookie) {
  struct Sass_Data_Context* data_context = sass_make_data_context(source);

  struct Sass_Context* context = sass_data_context_get_context(data_context);

  struct Sass_Options* options = sass_context_get_options(context);

  sass_option_set_input_path(options, input_path);

  Sass_C_Import_Callback importer = sass_make_importer(import_cb, cookie);
  sass_option_set_importer(options, importer);

  return data_context;
}
