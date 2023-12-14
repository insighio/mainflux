// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package httputil

import (
	"net/http"
)

// ExtractBearerToken returns value of the token.
func ExtractBearerToken(r *http.Request) string {
	token := r.Header.Get("Authorization")

	return token
}
