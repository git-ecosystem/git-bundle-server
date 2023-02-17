# Man pages

This directory contains Asciidoc-formatted files that are built and installed as
manual pages for the tools in the Git Bundle Server repository.

## File naming
Files with the extension `.adoc` will generate matching `man` page entries
through the `make doc` target. Supplemental content for these files (utilized
via [`include` directives][include]) must _not_ use the `.adoc` extension; the
recommended extension for these files is `.asc`.

[include]: https://docs.asciidoctor.org/asciidoc/latest/directives/include/

## Updating

When major user-facing features of the repository's CLI tools are added or
changed (e.g. new options or subcommands), the corresponding `.adoc` should be
updated to reflect those changes.
