package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/gobwas/glob"
	"github.com/lolorenzo777/loadfavicon/getfavicon"
	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v3"
)

const (
	ZSDIR  = ".zazzy"
	DFTPUBDIR = ".pub"
)

var PUBDIR string = DFTPUBDIR

type Vars map[string]string

// renameExt renames extension (if any) from oldext to newext
// If oldext is an empty string - extension is extracted automatically.
// If path has no extension - new extension is appended
func renameExt(path, oldext, newext string) string {
	if oldext == "" {
		oldext = filepath.Ext(path)
	}
	if oldext == "" || strings.HasSuffix(path, oldext) {
		return strings.TrimSuffix(path, oldext) + newext
	} else {
		return path
	}
}

// globals returns list of global OS environment variables that start
// with ZS_ prefix as Vars, so the values can be used inside templates
func globals() Vars {
	vars := Vars{}
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if strings.HasPrefix(pair[0], "ZS_") {
			vars[strings.ToLower(pair[0][3:])] = pair[1]
		}
	}

	// special environment variable
	if len(vars["pubdir"]) != 0 {
		PUBDIR = vars["pubdir"]
	}
	if len(vars["favicondir"]) == 0 {
		vars["favicondir"] = "/img/favicons"
	}

	return vars
}

// load .zazzy/.ignore file with list of files and directories to be ignored during the process
// each entry must be formatted as a glob pattern https://github.com/gobwas/glob
// return an array of trimed pattern of files to ignore
func loadIgnore() (lst []string) {
    f, err := os.Open(filepath.Join(ZSDIR, ".ignore"))
    if err != nil {
		// .ignore file is not mandatory
        return nil
    }
    defer f.Close()

    // read the file line by line using scanner
    scanner := bufio.NewScanner(f)

    for scanner.Scan() {
		entry := strings.Trim(scanner.Text(), " ")
		if len(entry)>0 && entry[:1] != "#" {
			if _, err := glob.Compile(entry); err != nil {
				log.Println(err)
			} else {
				lst = append(lst, entry)
			}
		}
    }

	// ensure PUBDIR is always ignored
	if filepath.Base(PUBDIR)[0] != '.' && !strings.HasPrefix(PUBDIR, ".") {
		pubdir := strings.TrimRight(PUBDIR, "/") + "**"
		lst = append(lst, pubdir)
	}
	return lst
}

var gSitemapWarning bool

// appendSitemap generate an entry in the sitemap.txt file
// according to paramaters: ZS_SITEMAPTXT must be true, and 
// the "sitemap: true" is in the YAML file header
func appendSitemap(path string, vars Vars) {
	if strings.ToLower(vars["sitemaptype"]) != "txt" {
		return 
	}

	if strings.ToLower(vars["sitemap"]) != "true" {
		return 
	}

	if len(vars["hosturl"]) == 0 && !gSitemapWarning {
		gSitemapWarning = true
		fmt.Println("Warning: generating sitemap without hosturl.")
	}

	sitemapentry := filepath.Join(vars["hosturl"], vars["url"])

    file, err := os.OpenFile(filepath.Join(PUBDIR, "sitemap.txt"), os.O_RDWR|os.O_CREATE, 0755)
    if err != nil {
		log.Println(err)
		return
    }
    defer file.Close()
	scanner := bufio.NewScanner(file)
    for scanner.Scan() {
		// do not add twice the same URL
		if strings.ToLower(strings.Trim(scanner.Text(), " ")) == sitemapentry {
			return 
		}
    }
    if err := scanner.Err(); err != nil {
		log.Println(err)
		return
    }
	if _, err := file.WriteString( sitemapentry +"\n"); err != nil {
		log.Println(err)
		return
	}
}

// run executes a command or a script. Vars define the command environment,
// each zs var is converted into OS environemnt variable with ZS_ prefix
// prepended.  Additional variable $ZS contains path to the zs binary. Command
// stderr is printed to zs stderr, command output is returned as a string.
func run(vars Vars, cmd string, args ...string) (string, error) {
	// external commande (plugin)
	var errbuf, outbuf bytes.Buffer
	c := exec.Command(cmd, args...)
	env := []string{"ZS=" + os.Args[0], "ZS_OUTDIR=" + PUBDIR}
	env = append(env, os.Environ()...)
	for k, v := range vars {
		env = append(env, "ZS_"+strings.ToUpper(k)+"="+v)
	}
	c.Env = env
	c.Stdout = &outbuf
	c.Stderr = &errbuf

	err := c.Run()

	if errbuf.Len() > 0 {
		log.Println("ERROR:", errbuf.String())
	}
	if err != nil {
		return "", err
	}
	return outbuf.String(), nil
}

// getDownloadedFavicon get favicon URL of the downloaded Favicon, and download it 
// if it doesn't exist on the local directory.
func getDownloadedFavicon(website string) (url string, err error) {

	vars := globals()
	faviconCachePath := filepath.Join(PUBDIR, vars["favicondir"])
	faviconSlugifiedWebsite := getfavicon.SlugHost(website)

	// look if favicon(s) has already been downloaded
	cache, err := filepath.Glob(filepath.Join(faviconCachePath, faviconSlugifiedWebsite) + "+*.*")
	if len(cache) == 0 && err == nil{
		// Connect to the website and download the best favicon
		favicons, err := getfavicon.Download(website, faviconCachePath, true)
		if len(favicons) == 0 {
			log.Println(err)
			return "", err
		}
		url = filepath.Join("/", vars["favicondir"], favicons[0].DiskFileName)
	} else {
		url = cache[0]
		if url[:len(PUBDIR)] != PUBDIR {
			panic("getDownloadedFavicon")
		}
		url = url[len(PUBDIR):]
	}

	return url, err
}

// renderFavicon donwload th favicon of a website given in paramaters 
// and generate html to render thefavicon image. 
func renderFavicon(vars Vars, args ...string) (string, error){
	if len(args) != 1 {
		log.Println("favicon placeholder requires a website in parameter. nothing rendered")
		return "", nil
	}

	faviconURL, err := getDownloadedFavicon(args[0])
	if len(faviconURL) > 0 && err == nil {
		return "<img src=\"" + faviconURL +"\" alt=\"icon\" class=\"favicon\" role=\"img\">", nil
	}
	return "", err
}

// renderlist generate an HTML string for every files in the pattern 
// passed in arg[0]. The string if rendered according to the itemlayout.html file.
// Than all strings are concatenated and ordered accordng to filenames in the pattern
func renderlist(vars Vars, args ...string) (string, error){
	// get the pattern of files to scan and list
	if len(args) != 1 {
		log.Println("renderlist placeholder requires pattern in parameter. nothing rendered.")
		return "", nil
	}
	filelistpattern := args[0]

	// check the pattern and get lisy of corresponding files
	matchingfiles, err := filepath.Glob(filelistpattern)
	if err != nil {
		return "", errors.New("bad pattern")
	}
	if len(matchingfiles) == 0 {
		fmt.Println("renderlist: no files corresponds to this pattern. The list is empty.", err)
		return "", errors.New("bad pattern")
	}
	sort.Sort(sort.Reverse(sort.StringSlice(matchingfiles)))
	// get list of files to ignore
	ignorelist := loadIgnore()

	// get the layout for items
	if _, ok := vars["itemlayout"]; !ok {
		vars["itemlayout"] = filepath.Join(ZSDIR, "itemlayout.html")
	}
	_, itemlayout, err := getVars(vars["itemlayout"], vars)
	if err != nil {
		fmt.Println("unable to proceed item layout file:", err)
	}

	// scan all existing files, and process as a list item
	result := ""
	for _, path := range matchingfiles {
		// ignore hidden files and directories
		if filepath.Base(path)[0] == '.' || strings.HasPrefix(path, ".") {
			continue
		}
		
		// ignore files and directory listed in the .zazzy/.ignore file
		for _, ignoreentry := range ignorelist {
			g, _ := glob.Compile(ignoreentry)
			if g.Match(path) {
				continue
			}
		}

		// inform user about fs errors, but continue iteration
		info, err := os.Stat(path)
		if err != nil {
			fmt.Println("renderlist item error:", err)
			continue
		}

		if info.IsDir() {
			continue
		} else {
			log.Println("renderlist item:", path)
			// load file's vars
			vitem, _, err := getVars(path, vars)
			if err != nil {
				fmt.Println("renderlist item error:", err)
				return "", err
			}
			vitem["file"] = path
			vitem["url"] = path[:len(path)-len(filepath.Ext(path))] + ".html"
			vitem["output"] = filepath.Join(PUBDIR, vitem["url"])
			item, err := render(itemlayout, vitem, 1)
			if err != nil {
				return "", err
			}
			result += item
		}
	}
	return result, nil
}


// getVars returns list of variables defined in a text file and actual file
// content following the variables declaration. Header is separated from
// content by an empty line. Header can be either YAML or JSON.
// If no empty newline is found - file is treated as content-only.
func getVars(path string, globals Vars) (Vars, string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	s := string(b)

	// Pick some default values for content-dependent variables
	v := Vars{}
	title := strings.Replace(strings.Replace(path, "_", " ", -1), "-", " ", -1)
	v["title"] = strings.ToTitle(title)
	v["description"] = ""
	v["file"] = path
	v["url"] = path[:len(path)-len(filepath.Ext(path))] + ".html"
	v["output"] = filepath.Join(PUBDIR, v["url"])

	// Override default values with globals
	for name, value := range globals {
		v[name] = value
	}

	// Add layout if none is specified
	if _, ok := v["layout"]; !ok {
		v["layout"] = "layout.html"
	}

	delim := "\n---\n"
	if sep := strings.Index(s, delim); sep == -1 {
		return v, s, nil
	} else {
		header := s[:sep]
		body := s[sep+len(delim):]

		vars := Vars{}
		if err := yaml.Unmarshal([]byte(header), &vars); err != nil {
			fmt.Println("ERROR: failed to parse header", err)
			return nil, "", err
		} else {
			// Override default values + globals with the ones defines in the file
			for key, value := range vars {
				v[key] = value
			}
		}
		v["url"] = strings.TrimLeft(v["url"], "./")
		//if strings.HasPrefix(v["url"], "./") {
		//	v["url"] = v["url"][2:]
		//}
		return v, body, nil
	}
}

// Render expanding zs plugins and variables, and process special command
func render(s string, vars Vars, deep int) (string, error) {
	delim_open := "{{"
	delim_close := "}}"

	out := &bytes.Buffer{}
	for {
		if from := strings.Index(s, delim_open); from == -1 {
			out.WriteString(s)
			return out.String(), nil
		} else {
			if to := strings.Index(s, delim_close); to == -1 {
				return "", fmt.Errorf("close delim not found")
			} else {
				out.WriteString(s[:from])
				cmd := s[from+len(delim_open) : to]
				s = s[to+len(delim_close):]
				m := strings.Fields(cmd)
				// proceed with special commands
				switch {
				case m[0] == "renderlist": 
					if res, err := renderlist(vars, m[1:]...); err == nil {
						out.WriteString(res)
					} else {
						fmt.Println(err)
					}
					continue
				case m[0] == "favicon" :
					if res, err := renderFavicon(vars, m[1:]...); err == nil {
						out.WriteString(res)
					} else {
						fmt.Println(err)
					}
					continue
				case filepath.Ext(m[0]) == ".html" || filepath.Ext(m[0]) == ".md":
					// proceed partials (.html or md) 
					if b, err := ioutil.ReadFile(filepath.Join(ZSDIR, m[0])); err == nil {
						// make it recursive
						if deep > 10 {
							return string(b), nil
						}
						if res, err := render(string(b), vars, deep+1); err == nil {
							out.WriteString(res)
						} else {
							fmt.Println(err)
						}
						continue
					}
					fallthrough
				case len(m) == 1 :
					// variable
					if v, ok := vars[m[0]]; ok {
						out.WriteString(v)
						continue
					}
				}

				// sz pluggins 
				if res, err := run(vars, m[0], m[1:]...); err == nil {
					out.WriteString(res)
				} else {
					fmt.Println(err)
				}
			}
		}
	}
}

// Renders markdown with the given layout into html expanding all the macros
func buildMarkdown(path string, w io.Writer, vars Vars) error {
	v, body, err := getVars(path, vars)
	if err != nil {
		return err
	}
	content, err := render(body, v, 1)
	if err != nil {
		return err
	}
	v["content"] = string(blackfriday.Run([]byte(content)))
	if w == nil {
		out, err := os.Create(filepath.Join(PUBDIR, renameExt(path, "", ".html")))
		if err != nil {
			return err
		}
		defer out.Close()
		w = out
	}
	appendSitemap(path, v)

	// process layout only if it exists
	layoutfile := filepath.Join(ZSDIR, v["layout"])
	_, errlayout := os.Stat(layoutfile)
	if errors.Is(errlayout, os.ErrNotExist) {
		_, err = io.WriteString(w, v["content"])
		return err
	} else if errlayout != nil {
		return errlayout
	}

	return buildHTML(filepath.Join(ZSDIR, v["layout"]), w, v)
}

// Renders text file expanding all variable macros inside it
func buildHTML(path string, w io.Writer, vars Vars) error {
	v, body, err := getVars(path, vars)
	if err != nil {
		return err
	}
	if body, err = render(body, v, 1); err != nil {
		return err
	}
	tmpl, err := template.New("").Delims("<%", "%>").Parse(body)
	if err != nil {
		return err
	}
	if w == nil {
		f, err := os.Create(filepath.Join(PUBDIR, path))
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	appendSitemap(path, v)

	return tmpl.Execute(w, vars)
}

// Copies file as is from path to writer
func buildRaw(path string, w io.Writer) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	if w == nil {
		if out, err := os.Create(filepath.Join(PUBDIR, path)); err != nil {
			return err
		} else {
			defer out.Close()
			w = out
		}
	}
	_, err = io.Copy(w, in)
	return err
}

func build(path string, w io.Writer, vars Vars) error {
	ext := filepath.Ext(path)
	var err error
	if ext == ".md" || ext == ".mkd" {
		err = buildMarkdown(path, w, vars)
	} else if ext == ".html" || ext == ".xml" {
		err = buildHTML(path, w, vars)
	} else {
		err = buildRaw(path, w)
	}
	if err != nil {
		log.Println(err)
	}
	return err
}

func buildAll(watch bool) {
	lastModified := time.Unix(0, 0)
	modified := false

	vars := globals()
	ignorelist := loadIgnore()
	// clear sitemap if any
	os.Remove(filepath.Join(PUBDIR, "sitemap.txt"))

	for {
		os.Mkdir(PUBDIR, 0755)
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			// ignore hidden files and directories
			if filepath.Base(path)[0] == '.' || strings.HasPrefix(path, ".") {
				return nil
			}
			
			// ignore files and directory listed in the .zazzy/.ignore file
			for _, ignoreentry := range ignorelist {
				g, _ := glob.Compile(ignoreentry)
				if g.Match(path) {
					return nil
				}
			}

			// inform user about fs walk errors, but continue iteration
			if err != nil {
				fmt.Println("error:", err)
				return nil
			}

			if info.IsDir() {
				os.Mkdir(filepath.Join(PUBDIR, path), 0755)
				return nil
			} else if info.ModTime().After(lastModified) {
				if !modified {
					// First file in this build cycle is about to be modified
					run(vars, "prehook")
					modified = true
				}
				log.Println("build:", path)
				return build(path, nil, vars)
			}
			return nil
		})
		if modified {
			// At least one file in this build cycle has been modified
			run(vars, "posthook")
			modified = false
		}
		if !watch {
			break
		}
		lastModified = time.Now()
		time.Sleep(1 * time.Second)
	}
}

func init() {
	// prepend .zazzy to $PATH, so plugins will be found before OS commands
	p := os.Getenv("PATH")
	p = ZSDIR + ":" + p
	os.Setenv("PATH", p)
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println(os.Args[0], "<command> [args]")
		return
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case "build":
		if len(args) == 0 {
			buildAll(false)
		} else if len(args) == 1 {
			if err := build(args[0], os.Stdout, globals()); err != nil {
				fmt.Println("ERROR: " + err.Error())
			}
		} else {
			fmt.Println("ERROR: too many arguments")
		}
	case "watch":
		buildAll(true)
	case "var":
		if len(args) == 0 {
			fmt.Println("var: filename expected")
		} else {
			s := ""
			if vars, _, err := getVars(args[0], Vars{}); err != nil {
				fmt.Println("var: " + err.Error())
			} else {
				if len(args) > 1 {
					for _, a := range args[1:] {
						s = s + vars[a] + "\n"
					}
				} else {
					for k, v := range vars {
						s = s + k + ":" + v + "\n"
					}
				}
			}
			fmt.Println(strings.TrimSpace(s))
		}
	default:
		if s, err := run(globals(), cmd, args...); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(s)
		}
	}
}
