#include "_cgo_export.h"
#include <sass_context.h>
#include <stdlib.h>
#include <stdio.h>

struct Sass_Import** import_cb(const char* parentPath,
                               const char* importPath,
                               void* cookie) {
  void* import_entries_ptr = go_import_cb((char*)parentPath,
                                          (char*)importPath,
                                          cookie);
  struct Sass_Import** imports = (struct Sass_Import**)(import_entries_ptr);
  return imports;
}

struct Sass_Data_Context* new_context(char* input_path, char* source, int compress, void* cookie) {
  struct Sass_Data_Context* data_context = sass_make_data_context(source);

  struct Sass_Context* context = sass_data_context_get_context(data_context);

  struct Sass_Options* options = sass_context_get_options(context);

  // SASS_STYLE_NESTED
  // SASS_STYLE_EXPANDED
  // SASS_STYLE_COMPACT
  // SASS_STYLE_COMPRESSED
  if (compress != 0) {
    sass_option_set_output_style(options, SASS_STYLE_COMPRESSED);
  }

  sass_option_set_input_path(options, input_path);

  Sass_C_Import_Callback importer = sass_make_importer(import_cb, cookie);
  sass_option_set_importer(options, importer);

  return data_context;
}
