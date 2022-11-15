# zazzy

Zazzy is a simple Static Site Generator, written in Go.

``zazzy`` is an extended version of the `sz` extremely minimal static site generator written in Go. `sz` itself was inspired by `zas` generator.

## Values

* Easy to learn
* Fast
* Highly extensible
* Works well for blogs and generic static websites (landing pages etc)

## Features

* Zero configuration (no configuration file needed)
* Publishing to Github Pages
* Customized layout by page
* Customized layout for embedded list (usefull to list blog posts)
* Ignore files if required
* Accept Standard Go Templating syntax
* Working with partials (aka. embedded html)
* pre and post build external command call
* `watch` feature for live generation during dev.
* Can run with ``LESS`` or any other CSS prprocessor
* sitemap generation

## Installation

Download the binaries from Github or build it manually:

```bash
go get github.com/lolorenzo777/zazzy
```

Install it:

```bash
go install github.com/lolorenzo777/zazzy
```

Then run zazzy in the directory of your website:

```bash
cd {mywebsite}
export ZS_PUBDIR=docs && zazzy build
```

## Ideology

Keep your texts in markdown or HTML format right in the main directory
of your blog/site.

Keep all service files (extensions, layout pages, deployment scripts etc)
in the `.zazzy` subdirectory.

Define variables in the header of the content files using [YAML][YAML]:

	title: My web site
	keywords: best website, hello, world
	---

	Markdown text goes after a header *separator*

Use placeholders for variables and plugins in your markdown or html
files, e.g. `{{ title }}` or `{{ command arg1 arg2 }}`.

Write extensions in any language you like and put them into the `.zazzy`
subdiretory.

Everything the extensions prints to stdout becomes the value of the
placeholder.

Every variable from the content header will be passed via environment variables like `title` becomes `$ZS_TITLE` and so on. There are some special variables:

* `$ZS` - a path to the `zs` executable
* `$ZS_OUTDIR` - a path to the directory with generated files
* `$ZS_FILE` - a path to the currently processed markdown file
* `$ZS_URL` - a URL for the currently generated page

## Default variables in file's header

- `layout:` defines the `.html` file to be used as the layout. By default, the `.zazzy/layout.html` will be used. If none of these files are founded, the file is processed without any layout.
- `title:` define the title of the page. By default this is the name of the file
- `description` define the description of the page. Nothing by default.
- `url` define the URL for the generated page. By default this is the markdown filename with the ``.html`` extension.

## Ignored files

Hidden directories and files, and the one starting with a ``.`` are ignored. They're not processed and will not be generated to the published directory.

You can also list files to ignore into the `.zazzy/.ignore` file. Each line must represents the glob pattern of filenames to ignore.

```shell
# example of .zazzy/.ignore file
# ignore the readme.md file in the working directory
readme.md
# ignore the test directory and all its content
test**
# ignote all txt files
*.txt
```

The published directory is always ignored. You don't need to include it in the .ignore file. For example if your website aims to be published to Github Pages, the ``docs `` directory will be ignored from the generation.

## Publishing to Github Pages

To publish your website to Github Pages you've limited choice for the published directory, it's either the root of your repository or the `docs` directory of your repository.

So with zazzy it's easy, you just have to set the $SZ_PUBDIR environnement variable to `docs` and that's it.

The command to run the build looks like that:
```
$SZ_PUBDIR=docs zazzy build
```

## Working with partials (aka. embedded html)

Create your partial file into the `.zazzy` directory. Partials can be `.html` or `.md` files.

Then insert it's placeholder whenever you want into your other files (could be the layout too). 

For example if you've created the file `.zazzy/footer.html` and you want to use it into the `index.md`, then add the following code into this file:

```html
{{ footer.html }}
```

It's important not to add any quotes or file extension, only the file name, working as a variable.

The rendering of partials is recursive, that means in our example footer.htlm can include a placholder to embedd another partial.

## Standard Go Templating syntax 

If you're familiar with the [go html template package](https://pkg.go.dev/html/template) then you can use it's syntax directly within your html layout.

To avoid conflict with zazzy placeholder symbols `{{` and `}}` then the html templating syntax must be call with `<%` and `%>` symbols.

## Advanced placeholder `renderlist`

placeholder `{{ renderlist {pattern} }}` run special command to parse itemlayout html for every file in the list correspondinf to the pattern. Usefull to generate a list of blog entries.

Default layout for items is `itemlayout.html` but can be specified in YAML environment variable of the page where the placeholder is found.

## Advanced placeholder `favicon` 

placeholder `{{ favicon {website} }}` run special command to download and generate html to render thefavicon image.

## sitemap

To generate a sitemap then declare `ZS_SITEMAPTYPE=txt` in your environmetnt variable, then add the YAML entry `sitemap: true` in every `markdown` and `html` files you want to seein the sitemap.

## Example of RSS generation

Extensions can be written in any language you know (Bash, Python, Lua, JavaScript, Go, even Assembler). Here's an example of how to scan all markdown blog posts and create RSS items:

``` bash
for f in ./blog/*.md ; doc
	d=$($ZS var $f date)
	if [ ! -z $d ] ; then
		timestamp=`date --date "$d" +%s`
		url=`$ZS var $f url`
		title=`$ZS var $f title | tr A-Z a-z`
		descr=`$ZS var $f description`
		echo $timestamp \
			"<item>" \
			"<title>$title</title>" \
			"<link>http://zserge.com/$url</link>" \
			"<description>$descr</description>" \
			"<pubDate>$(date --date @$timestamp -R)</pubDate>" \
			"<guid>http://zserge.com/$url</guid>" \
		"</item>"
	fi
done | sort -r -n | cut -d' ' -f2-
```

## Hooks

There are two special plugin names that are executed every time the build
happens - `prehook` and `posthook`. You can define some global actions here like
content generation, or additional commands, like LESS to CSS conversion:

	# .zs/post

	#!/bin/sh
	lessc < $ZS_OUTDIR/styles.less > $ZS_OUTDIR/styles.css
	rm -f $ZS_OUTDIR/styles.css

## Command line usage

`zazzy init <title> <hosturl> [--vscode] [--ghpages] [--sitemap]` create a fresh new website with default files and path

`zazzy build` re-builds your site.

`zazzy build <file>` re-builds one file and prints resulting content to stdout.

`zazzy watch` rebuilds your site every time you modify any file.

`zazzy var <filename> [var1 var2...]` prints a list of variables defined in the
header of a given markdown file, or the values of certain variables (even if
it's an empty string).

## Versions

Fork of `zs` version commit [`4900afa45db4d9254110f2eabcac6cfd606423b6`](https://github.com/zserge/zs/commit/4900afa45db4d9254110f2eabcac6cfd606423b6)

### v0.4

- Feature command 'init' to generate a fresh new site
- migrate to go v1.19.3

### v0.3

- Feature generate a sitemap.txt
- Feature favicon placeholder, with cache and download of favicon image
- Feature recursive rendering throu partials
- remove amber and gcss features

### V0.2

- placeholder `{{ renderlist {pattern} }}`

### V0.1

- upgraded to go 1.16
- enhancement to allow processing of markdonw and amber file without layout file
- ``.zazzy/.ignore`` file allows to list files and directories to be igniorred from the processor

### Zazzy go module

Zazzy is published and available in the go public package manager https://pkg.go.dev/github.com/lolorenzo777/zazzy

`GOPROXY=proxy.golang.org go list -m github.com/lolorenzo777/zazzy@vx.y.z`

## License

The software is distributed under the [MIT license](LICENSE).

[YAML]: https://github.com/go-yaml/yaml