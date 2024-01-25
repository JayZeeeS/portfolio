package main

import (
	"bytes"
	"fmt"

	//"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	//"golang.org/x/text/cases"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

type Page struct {
	Title       string
	Description string
	Links       []DirLink
	Body        []Code
	Content     interface{}
}

type DirLink struct {
	URL  string
	Text string
}

type Code struct {
	Filename string
	Content  string
}

func main() {
    http.Handle("/styles/", http.StripPrefix("/styles/",
        http.FileServer(http.Dir("styles"))))

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/projects/", projHandler)
	http.ListenAndServe(":8080", nil)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
    templ, err := templates()
    if err != nil {
        log.Println(err)
    }

	projectDir, err := os.ReadDir("projects")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	index := projectIndex(projectDir, "/projects")
	data := Page{
		Title:       "Welcome!",
		Description: "",
		Links:       index,
	}

	err = templ.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Println(err)
	}
}

func projectIndex(files []os.DirEntry, path string) []DirLink {
	var links []DirLink
	for _, v := range files {
		if !v.IsDir() {
			continue
		}
		newProject := DirLink{
            URL: filepath.Join("/", path, v.Name()),
            Text: v.Name(),
        }
		links = append(links, newProject)
	}
	return links
}

func projHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/"):]
	dir, err := os.ReadDir(path)
	if err != nil {
		log.Printf("error reading directory: %v\n", err)
	}
	//log.Println(dir)
	links := projectIndex(dir, path)
	code := []Code{}
	for _, v := range dir {
		if !v.IsDir() {
			content, err := codeContent(filepath.Join(path, v.Name()))
			if err != nil {
				//      log.Println(v.Name())
				log.Printf("error at codeContent(): %v", err)
			}
			code = append(code, content)
		}
	}

	var highlightedCode []Code
	for _, v := range code {
		newBlock, err := highlightCode(v)
		if err != nil {
			log.Println(err)
		}
		// log.Printf("Highlited the folllowing code:\n%v\n", newBlock)
		highlightedCode = append(highlightedCode, newBlock)
	}
	// log.Printf("This is the highlited code:\n%v\n", highlightedCode)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	desc, err := os.ReadFile(filepath.Join(path, "description.txt"))
	if err != nil {
		// log.Println("No desc file")
	}
    titlePath := strings.Split(path, "/")
    title := titlePath[len(titlePath)-1]
	data := Page{
		Title:       strings.Title(title),
		Description: string(desc),
		Links:       links,
		Body:        highlightedCode,
	}
    tmpl, err := templates()
    if err != nil {
        log.Println(err)
    }
	err = tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Println(err)
	}
}

func codeContent(filename string) (content Code, err error) {
	fileContents, err := os.ReadFile(filename)
	if err != nil {
		return Code{}, err
	}
	name := strings.Split(filename, "/")
	content = Code{Filename: name[len(name)-1], Content: string(fileContents)}
	return
}

func highlightCode(codeBlock Code) (highlitedCodeBlock Code, err error) {
	lexer := lexers.Match(codeBlock.Filename)
	if lexer == nil {
		log.Printf("Could not get lexer for %v", codeBlock.Filename)
		lexer = lexers.Fallback
	}

	formatter := html.New(html.WithLineNumbers(true), html.TabWidth(2))
	style := styles.Get("base16-snazzy")
	if style == nil {
		log.Printf("Could not get style for %v", codeBlock.Filename)
		style = styles.Fallback
	}

	var cssBuilder strings.Builder
	formatter.WriteCSS(&cssBuilder, style)

	iterator, err := lexer.Tokenise(nil, codeBlock.Content)
	if err != nil {
		return Code{}, err
	}

	var highlitedCode bytes.Buffer
	err = formatter.Format(&highlitedCode, style, iterator)
	if err != nil {
		return Code{}, err
	}

	highlitedCodeBlock = Code{
        Content: highlitedCode.String(),
        Filename: codeBlock.Filename,
    }
	return
}

func templates() (templ *template.Template, err error) {
	funcMap := template.FuncMap{
		"safeHTML": func(html string) template.HTML {
			return template.HTML(html)
		}}
    
    templates, err := filepaths("./layouts")
    if err != nil {
        return nil, err
    }

	templ, err = template.New("").Funcs(funcMap).ParseFiles(templates...)
	if err != nil {
		log.Printf("error reading templates: %v\n", err)
		return nil, err
	}

	return templ, nil
}

func filepaths(dirname string) (files []string, err error) {
    dir, err := os.ReadDir(dirname)
    if err != nil {
        return files, fmt.Errorf("error reading directory:= %v\n", err)
    }
    for _, v := range dir {
        if v.IsDir() {
            newPaths, err := filepaths(filepath.Join(dirname, v.Name()))
            if err != nil {
                log.Printf("error recursively reading files: %v\n", err)
            }
            files = append(files, newPaths...)
            continue
        }
        files = append(files, filepath.Join(dirname, v.Name()))
    }
    return
}
