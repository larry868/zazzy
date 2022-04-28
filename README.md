zazzy
==

zazzy is an extended version of the `sz` extremely minimal static site generator written in Go.

`sz` itself was inspired by `zas` generator, but is even more minimal.

## Features

* Zero configuration (no configuration file needed)
* Cross-platform
* Highly extensible
* Works well for blogs and generic static websites (landing pages etc)
* Easy to learn
* Fast

## Installation

Download the binaries from Github or build it manually:

	$ go get github.com/lolorenzo777/zazzy

then install it 

	$ go install github.com/lolorenzo777/zazzy

## Ideology

Keep your texts in markdown, [amber] or HTML format right in the main directory
of your blog/site.

Keep all service files (extensions, layout pages, deployment scripts etc)
in the `.zs` subdirectory.

Define variables in the header of the content files using [YAML]:

	title: My web site
	keywords: best website, hello, world
	---

	Markdown text goes after a header *separator*

Use placeholders for variables and plugins in your markdown or html
files, e.g. `{{ title }}` or `{{ command arg1 arg2 }}.

Write extensions in any language you like and put them into the `.zs`
subdiretory.

Everything the extensions prints to stdout becomes the value of the
placeholder.

Every variable from the content header will be passed via environment variables like `title` becomes `$ZS_TITLE` and so on. There are some special variables:

* `$ZS` - a path to the `zs` executable
* `$ZS_OUTDIR` - a path to the directory with generated files
* `$ZS_FILE` - a path to the currently processed markdown file
* `$ZS_URL` - a URL for the currently generated page

## Default variables in file's header

- `layout:` defines the `.html` or `.amber` file to be used as the layout. By default, the `.sz/layout.html` or `.sz/layout.amber` will be used. If none of these files are founded, the file is processed without any layout.
- `title:` define the title of the page. By default this is the name of the file
- `description` define the description of the page. Nothing by default.
- `url` define the URL for the generated page. By default this is the markdown or amber filename with the ``.html`` extension.

## Ignored files

Hidden directories and files, and the one starting with a ``.`` are ignored. They're not processed and will not be generated to the PUBDIR.

You can also list files to ignore into the `.zazzy/.ignore` file. Each ligne must represent the glob pattern of filenames to ignore.

```shell
# ignore the readme.md file in the working directory
readme.md
# ignore the test directory and all its content
test**
# ignote all txt files
*.txt
```

By default ``.pub`` published directory is ignored. If you change it, it's also systematically ignored even if not starting with a ``.``. For example if your website aims to be published to Github Pages, the ``docs `` directory will be ignored from the generation.

## Publishing to Github Pages

To publish your website to Github Pages you've limited choice for the published directory, it's either the root of your repository or the `docs` directory of your repository.

So with zazzy it's easy, you just have to set the $SZ_PUBDIR environnement variable to `docs` and that's it.

The command to run the build looks like that:
```
$SZ_PUBDIR=docs zazzy build
```

## Working with partials (aka. embedded html)

Create your partial file into the `.zazzy` directory. Partials must be either `.html` or `.amber` files.

Then insert it's placeholder whenever you want into your other files (could be the layout too). 

For example if you've created the file `.zazzy/footer.html` and you want to use it into the `index.md`, then add the following code into this file:

```html
{{ footer }}
```

It's important not to add any quotes or file extension, only the file name, working as a variable.

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

## Syntax sugar

By default, `zs` converts each `.amber` file into `.html`, so you can use lightweight Jade-like syntax instead of bloated HTML.

Also, `zs` converts `.gcss` into `.css`, so you don't really need LESS or SASS. More about GCSS can be found [here][gcss].

## Command line usage

`zs build` re-builds your site.

`zs build <file>` re-builds one file and prints resulting content to stdout.

`zs watch` rebuilds your site every time you modify any file.

`zs var <filename> [var1 var2...]` prints a list of variables defined in the
header of a given markdown file, or the values of certain variables (even if
it's an empty string).

## Versions

Starting from `zs` version +

### tag V0.1.0
- upgraded to go 1.16
- enhancement to allow processing of markdonw and amber file without layout file
- ``.zs/.ignore `` file allows to list files and directories to be igniorred from the processor

## License

The software is distributed under the MIT license.

[amber]: https://github.com/eknkc/amber/
[YAML]: https://github.com/go-yaml/yaml
[gcss]: https://github.com/yosssi/gcss
