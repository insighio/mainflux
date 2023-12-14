// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package keys

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	kitot "github.com/go-kit/kit/tracing/opentracing"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/internal/httputil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/opentracing/opentracing-go"
)

const (
	contentType = "application/json"
	offsetKey   = "offset"
	limitKey    = "limit"
	subjectKey  = "subject"
	typeKey     = "type"
	defOffset   = 0
	defLimit    = 10
	defType     = 2
)

var errUnsupportedContentType = errors.New("unsupported content type")

func MakeHandler(svc auth.Service, mux *bone.Mux, tracer opentracing.Tracer) *bone.Mux {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}
	mux.Post("/keys", kithttp.NewServer(
		kitot.TraceServer(tracer, "issue")(issueEndpoint(svc)),
		decodeIssue,
		encodeResponse,
		opts...,
	))

	mux.Get("/keys", kithttp.NewServer(
		kitot.TraceServer(tracer, "issue")(retrieveKeysEndpoint(svc)),
		decodeListKeysRequest,
		encodeResponse,
		opts...,
	))

	mux.Get("/keys/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "retrieve")(retrieveEndpoint(svc)),
		decodeKeyReq,
		encodeResponse,
		opts...,
	))

	mux.Delete("/keys/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "revoke")(revokeEndpoint(svc)),
		decodeKeyReq,
		encodeResponse,
		opts...,
	))

	return mux
}

func decodeIssue(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errUnsupportedContentType
	}
	req := issueKeyReq{
		Token: r.Header.Get("Authorization"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(auth.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeKeyReq(_ context.Context, r *http.Request) (interface{}, error) {
	req := keyReq{
		token: r.Header.Get("Authorization"),
		id:    bone.GetValue(r, "id"),
	}
	return req, nil
}

func decodeListKeysRequest(_ context.Context, r *http.Request) (interface{}, error) {
	s, err := httputil.ReadStringQuery(r, subjectKey, "")
	if err != nil {
		return nil, err
	}

	t, err := httputil.ReadUintQuery(r, typeKey, defType)
	if err != nil {
		return nil, err
	}

	o, err := httputil.ReadUintQuery(r, offsetKey, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := httputil.ReadUintQuery(r, limitKey, defLimit)
	if err != nil {
		return nil, err
	}

	req := listKeysReq{
		token:   httputil.ExtractBearerToken(r),
		subject: s,
		keyType: uint32(t),
		offset:  o,
		limit:   l,
	}
	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", contentType)

	if ar, ok := response.(mainflux.Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}

		w.WriteHeader(ar.Code())

		if ar.Empty() {
			return nil
		}
	}

	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	switch {
	case errors.Contains(err, auth.ErrMalformedEntity):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Contains(err, auth.ErrUnauthorizedAccess):
		w.WriteHeader(http.StatusForbidden)
	case errors.Contains(err, auth.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
	case errors.Contains(err, errors.ErrInvalidQueryParams),
		errors.Contains(err, errors.ErrMalformedEntity),
		err == auth.ErrMissingID,
		err == auth.ErrBearerKey,
		err == auth.ErrLimitSize,
		err == auth.ErrOffsetSize,
		err == auth.ErrInvalidIDFormat:
		w.WriteHeader(http.StatusBadRequest)
	case errors.Contains(err, auth.ErrConflict):
		w.WriteHeader(http.StatusConflict)
	case errors.Contains(err, io.EOF):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Contains(err, io.ErrUnexpectedEOF):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Contains(err, errUnsupportedContentType):
		w.WriteHeader(http.StatusUnsupportedMediaType)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	errorVal, ok := err.(errors.Error)
	if ok {
		if err := json.NewEncoder(w).Encode(errorRes{Err: errorVal.Msg()}); err != nil {
			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
