package main

import (
	"github.com/howardedidin/bcda-app/bcda/responseutils"
	"net/http"
)

func ValidateBulkRequestHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header

		acceptHeader := h.Get("Accept")
		preferHeader := h.Get("Prefer")

		if acceptHeader == "" {
			oo := responseutils.CreateOpOutcome(responseutils.Error, responseutils.Structure, "", responseutils.FormatErr)
			oo.Issue[0].Diagnostics = "Accept header is required"
			responseutils.WriteError(oo, w, http.StatusBadRequest)
			return
		} else if acceptHeader != "application/fhir+json" {
			oo := responseutils.CreateOpOutcome(responseutils.Error, responseutils.Structure, "", responseutils.FormatErr)
			oo.Issue[0].Diagnostics = "application/fhir+json is the only supported response format"
			responseutils.WriteError(oo, w, http.StatusBadRequest)
			return
		}

		if preferHeader == "" {
			oo := responseutils.CreateOpOutcome(responseutils.Error, responseutils.Structure, "", responseutils.FormatErr)
			oo.Issue[0].Diagnostics = "Prefer header is required"
			responseutils.WriteError(oo, w, http.StatusBadRequest)
			return
		} else if preferHeader != "respond-async" {
			oo := responseutils.CreateOpOutcome(responseutils.Error, responseutils.Structure, "", responseutils.FormatErr)
			oo.Issue[0].Diagnostics = "Only asynchronous responses are supported"
			responseutils.WriteError(oo, w, http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
