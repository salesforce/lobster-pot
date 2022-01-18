// Copyright (c) 2022, salesforce.com, inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
// Package handlers - auth
// Contains the logic to provide an authentication middleware
package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// AuthCheck acts as a middle-ware to check that the correct github headers have been supplied
// by the incoming githook
func AuthCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// check if valid github Post
		xsig := r.Header.Get("X-Hub-Signature")
		xguid := r.Header.Get("X-GitHub-Delivery")
		xevent := r.Header.Get("X-GitHub-Event")

		if xsig == "" || xguid == "" || xevent == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte("Nope!"))
			if err != nil {
				log.Error(err)
			}
			return
		}

		// all passes, pass it on to the handler

		next.ServeHTTP(w, r)

	})
}
