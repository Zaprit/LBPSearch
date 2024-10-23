package website

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
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
var BackupFailTemplate = template.Must(template.ParseFS(Templates, "templates/backupfail.html"))
var NotFoundTemplate = template.Must(template.ParseFS(Templates, "templates/notfound.html"))

func MissingRootLevelTemplate(w http.ResponseWriter, slotName, levelID, headerInjection, requestID string) {
	err := BackupFailTemplate.Execute(w, map[string]any{
		"LevelName":       slotName,
		"LevelID":         levelID,
		"HeaderInjection": template.HTML(headerInjection),
		"RequestID":       requestID,
		"Message":         template.HTML("The specified level's rootLevel (the primary resource that contains the required information for an LBP level) could not be found in the archive. Unfortunately this means that the level cannot be downloaded.<br>Sorry for the inconvenience."),
		"ShortMessage":    "Missing Root Level",
		"FailType":        "missingRootLevel",
	})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

	if err != nil {
		slog.Error("failed to execute template", slog.Any("error", err))
	}
}
