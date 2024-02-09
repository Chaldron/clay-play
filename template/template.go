package template

import (
	"html/template"
	"os"
	"path/filepath"
	"time"
)

type TemplateMap map[string]*template.Template

func Generate() (TemplateMap, error) {
	templates := TemplateMap{}

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
		return TemplateMap{}, err
	}

	for _, pagePath := range pages {
		name := pagePath[len(pagesPath)+1:]
		t := template.New(name)

		t.Funcs(template.FuncMap{
			"jsTime": jsTime,
		})

		t, err = t.ParseFiles(
			filepath.Join(rootPath, "base.html"),
			filepath.Join(rootPath, "header.html"),
			pagePath,
		)
		if err != nil {
			return TemplateMap{}, err
		}

		templates[name] = t
	}

	errorNotifFile := filepath.Join(rootPath, "error-notif.html")
	errorNotifTemplate, err := template.ParseFiles(errorNotifFile)
	if err != nil {
		return TemplateMap{}, err
	}

	templates[filepath.Base(errorNotifFile)] = errorNotifTemplate

	return templates, nil
}

func jsTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05Z")
}
