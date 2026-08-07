# webarchive-extract

Unpacks Safari `.webarchive` files into plain HTML.

For each archive it writes `<archive name>.html` into the `-out` directory, or
straight to `-out` itself when that path ends in `.html`. By default every
embedded stylesheet, script, image and font is folded into that file as a
`data:` URI, so the result is a single self-contained page. Use `-inline-max` to
send the larger assets to an `embed/` folder next to the HTML file instead and
reference them from there.

Both binary and XML (plutil-converted) archives are handled by a built-in
parser, with `plutil` as a fallback. Whatever the page already carried inline —
`<style>` and `<script>` blocks, existing `data:` URIs — is left exactly as it
was. Subframe archives are flattened in, so iframe sources get extracted and
rewritten like anything else.

## Install

```sh
go install github.com/msoap/etc/webarchive-extract@latest
```

## Usage

```sh
webarchive-extract [flags] archive.webarchive...
```

Given several archives, each one gets its own subfolder under the embed
directory so same-named assets from different sites cannot collide.

```sh
webarchive-extract page.webarchive                    # one self-contained page.html
webarchive-extract -inline-max 200k page.webarchive   # big assets go to ./embed/
webarchive-extract -list page.webarchive              # just show what's inside
webarchive-extract -out build *.webarchive            # batch into ./build/
webarchive-extract -out docs/intro.html page.webarchive   # exact output file
```

An `-out` path ending in `.html` (or `.htm`) names the output file itself
rather than a directory. Missing directories are created, and the embed folder
goes next to that file — `-out docs/intro.html` writes `docs/intro.html` plus
`docs/embed/`. Since it names one file, it takes exactly one archive.

## Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-out <path>` | `.` | Directory to write the HTML file into, or an exact path ending in `.html`. |
| `-inline-max <size>` | `all` | Embed assets smaller than this as `data:` URIs, write bigger ones to files. `all` inlines everything, `0` writes everything to files. Accepts `200k`, `1.5M`, `2gb`, … |
| `-data-uris` | `false` | Extract `data:` URIs the page already had into files. |
| `-rewrite` | `true` | Rewrite asset references to the embed directory. |
| `-manifest` | `true` | Write `embed/manifest.json` mapping files to original URLs. |
| `-list` | `false` | List archive contents and exit without writing anything. |
| `-plutil` | `false` | Always decode via `plutil` instead of the built-in parser. |
| `-v` | `false` | Print every file written. |

## Environment

| Variable | Default | Description |
| --- | --- | --- |
| `WEBARCHIVE_EMBED_DIR` | `embed` | Assets directory, relative to the HTML file. Must be a relative path. |
