package template

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type Map map[string]*template.Template

func Generate() (map[string]*template.Template, error) {
	templates := map[string]*template.Template{}

	rootPath := "ui/views"
	pagesPath := filepath.Join(rootPath, "pages")

	pages := []string{}
	err := filepath.WalkDir(pagesPath, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			pages = append(pages, path)
		}
		return nil
	})
	if err != nil {
		return map[string]*template.Template{}, err
	}

	for _, pagePath := range pages {
		name := strings.TrimPrefix(pagePath, pagesPath+"/")
		t := template.New(name)

		t, err = t.ParseFiles(
			filepath.Join(rootPath, "base.html"),
			filepath.Join(rootPath, "header.html"),
			pagePath,
		)
		if err != nil {
			return map[string]*template.Template{}, err
		}

		templates[name] = t
	}

	errorNotifFile := filepath.Join(rootPath, "error-notif.html")
	errorNotifTemplate, err := template.ParseFiles(errorNotifFile)
	if err != nil {
		return map[string]*template.Template{}, err
	}

	templates[filepath.Base(errorNotifFile)] = errorNotifTemplate

	return templates, nil
}
