// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"github.com/mainflux/mainflux/authn"
	"fmt"
)

type identityReq struct {
	token string
	kind  uint32
}

func (req identityReq) validate() error {
	if req.token == "" {
		return authn.ErrMalformedEntity
	}
	if req.kind != authn.UserKey &&
		req.kind != authn.APIKey &&
		req.kind != authn.RecoveryKey &&
		req.kind != authn.EmailVerificationKey {
		return authn.ErrMalformedEntity
	}

	return nil
}

type issueReq struct {
	issuer  string
	keyType uint32
}

func (req issueReq) validate() error {
	fmt.Printf("grpc-requests 1: %s\n", req.issuer)
	if req.issuer == "" {
		fmt.Printf("grpc-requests 2\n")
		return authn.ErrUnauthorizedAccess
	}
	fmt.Printf("grpc-requests 3\n")
	if req.keyType != authn.UserKey &&
		req.keyType != authn.APIKey &&
		req.keyType != authn.RecoveryKey &&
		req.keyType != authn.EmailVerificationKey {
			fmt.Printf("grpc-requests 4\n")
		return authn.ErrMalformedEntity
	}

	fmt.Printf("grpc-requests 4\n")
	return nil
}
