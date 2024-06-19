package website

import (
	"embed"
	"html/template"
	xmlTmpl "text/template"
)

//go:embed templates
var Templates embed.FS

//go:embed static
var Static embed.FS

var Changelog string

var IndexTemplate = template.Must(template.ParseFS(Templates, "templates/index.html"))
var SlotTemplate = template.Must(template.ParseFS(Templates, "templates/slot.html"))
var UserTemplate = template.Must(template.ParseFS(Templates, "templates/user.html"))
var ChangelogTemplate = template.Must(template.ParseFS(Templates, "templates/changelog.html"))
var SitemapTemplate = xmlTmpl.Must(xmlTmpl.ParseFS(Templates, "templates/sitemap_other.xml"))
var SitemapIndexTemplate = xmlTmpl.Must(xmlTmpl.ParseFS(Templates, "templates/sitemapindex.xml"))
var RobotsTXTTemplate = template.Must(template.ParseFS(Templates, "templates/robots.txt"))
