// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package keys

import (
	"time"

	"github.com/mainflux/mainflux/auth"
)

type issueKeyReq struct {
	Token    string        `json:"token,omitempty"`
	Type     uint32        `json:"type,omitempty"`
	Name     string        `json:"name,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

// It is not possible to issue Reset key using HTTP API.
func (req issueKeyReq) validate() error {
	if req.Type == auth.UserKey {
		return nil
	}
	if (req.Token == "") || (req.Type != auth.APIKey) || (req.Name == "") {
		return auth.ErrMalformedEntity
	}
	return nil
}

type keyReq struct {
	token string
	id    string
}

func (req keyReq) validate() error {
	if req.token == "" || req.id == "" {
		return auth.ErrMalformedEntity
	}
	return nil
}

type listKeysReq struct {
	token   string
	subject string
	keyType uint32
	offset  uint64
	limit   uint64
}

func (req listKeysReq) validate() error {
	if req.token == "" {
		return auth.ErrBearerToken
	}

	if req.limit < 1 {
		return auth.ErrLimitSize
	}

	return nil
}
