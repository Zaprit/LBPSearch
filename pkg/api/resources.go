package api

import "net/http"

func ResourcesEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resource := r.PathValue("resource")

		http.Redirect(w, r, "https://lbparchive.zaprit.fish/"+resource[0:2]+"/"+resource[2:4]+"/"+resource, http.StatusFound)
	}
}
