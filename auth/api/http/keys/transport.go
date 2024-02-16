// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package keys

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/internal/apiutil"
	"github.com/absmach/magistrala/pkg/errors"
	"github.com/go-chi/chi/v5"
	kithttp "github.com/go-kit/kit/transport/http"
)

const (
	contentType = "application/json"
	offsetKey   = "offset"
	limitKey    = "limit"
	subjectKey  = "subject"
	typeKey     = "type"
	defOffset   = 0
	defLimit    = 10
	defType     = 3
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc auth.Service, mux *chi.Mux, logger *slog.Logger) *chi.Mux {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}
	mux.Route("/keys", func(r chi.Router) {
		r.Get("/", kithttp.NewServer(
			retrieveKeysEndpoint(svc),
			decodeListKeysRequest,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Post("/", kithttp.NewServer(
			issueEndpoint(svc),
			decodeIssue,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Get("/{id}", kithttp.NewServer(
			(retrieveEndpoint(svc)),
			decodeKeyReq,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Delete("/{id}", kithttp.NewServer(
			(revokeEndpoint(svc)),
			decodeKeyReq,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)
	})
	return mux
}

func decodeIssue(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := issueKeyReq{token: apiutil.ExtractBearerToken(r)}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeKeyReq(_ context.Context, r *http.Request) (interface{}, error) {
	req := keyReq{
		token: apiutil.ExtractBearerToken(r),
		id:    chi.URLParam(r, "id"),
	}
	return req, nil
}

func decodeListKeysRequest(_ context.Context, r *http.Request) (interface{}, error) {
	s, err := apiutil.ReadStringQuery(r, subjectKey, "")
	if err != nil {
		return nil, err
	}

	t, err := apiutil.ReadNumQuery[uint64](r, typeKey, defType)
	if err != nil {
		return nil, err
	}

	o, err := apiutil.ReadNumQuery[uint64](r, offsetKey, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := apiutil.ReadNumQuery[uint64](r, limitKey, defLimit)
	if err != nil {
		return nil, err
	}

	req := listKeysReq{
		token:   apiutil.ExtractBearerToken(r),
		subject: s,
		keyType: uint32(t),
		offset:  o,
		limit:   l,
	}
	return req, nil
}
