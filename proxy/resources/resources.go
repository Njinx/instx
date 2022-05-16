package resources

import (
	"embed"
	"text/template"
)

//go:embed root
var fs embed.FS

var pages = map[string]string{
	"/":               "root/root.gohtml",
	"/opensearch.xml": "root/opensearch.xml",
	"/getstarted":     "root/getstarted.html",
	"/favicon.ico":    "root/favicon.ico",
}

type VFS struct {
	pages map[string]string
	fs    embed.FS
}

func New() VFS {
	return VFS{
		pages: pages,
		fs:    fs,
	}
}

func (vfs *VFS) GetFile(path string) (template.Template, error) {
	tmpl, err := template.ParseFS(vfs.fs, vfs.pages[path])
	if err != nil {
		return template.Template{}, err
	}

	return *tmpl, nil
}
