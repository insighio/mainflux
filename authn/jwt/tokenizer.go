// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package jwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mainflux/mainflux/authn"
	"fmt"
)

type claims struct {
	jwt.StandardClaims
	Type *uint32 `json:"type,omitempty"`
}

func (c claims) Valid() error {
	if c.Type == nil || *c.Type > authn.APIKey {
		return authn.ErrMalformedEntity
	}

	return c.StandardClaims.Valid()
}

type tokenizer struct {
	secret string
}

// New returns new JWT Tokenizer.
func New(secret string) authn.Tokenizer {
	return tokenizer{secret: secret}
}

func (svc tokenizer) Issue(key authn.Key) (string, error) {
	claims := claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:   key.Issuer,
			Subject:  key.Secret,
			IssuedAt: key.IssuedAt.UTC().Unix(),
		},
		Type: &key.Type,
	}

	if !key.ExpiresAt.IsZero() {
		claims.ExpiresAt = key.ExpiresAt.UTC().Unix()
	}
	if key.ID != "" {
		claims.Id = key.ID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(svc.secret))
}

func (svc tokenizer) Parse(token string) (authn.Key, error) {
	c := claims{}
	fmt.Printf("tokenizer-parse 1: %s\n", token)
	_, err := jwt.ParseWithClaims(token, &c, func(token *jwt.Token) (interface{}, error) {
		fmt.Printf("tokenizer-parse 2.1\n")
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Printf("tokenizer-parse 2.2\n")
			return nil, authn.ErrUnauthorizedAccess
		}
		fmt.Printf("tokenizer-parse 2.3: %s\n", svc.secret)
		return []byte(svc.secret), nil
	})

	fmt.Printf("tokenizer-parse 3: %s\n", err)

	if err != nil {
		if e, ok := err.(*jwt.ValidationError); ok && e.Errors == jwt.ValidationErrorExpired {
			// Expired User key needs to be revoked.
			if c.Type != nil && *c.Type == authn.APIKey {
				return c.toKey(), nil
			}
			fmt.Printf("tokenizer-parse 4\n")
			return authn.Key{}, authn.ErrKeyExpired
		}
		fmt.Printf("tokenizer-parse 5\n")
		return authn.Key{}, authn.ErrUnauthorizedAccess
	}

	fmt.Printf("tokenizer-parse 6\n")
	return c.toKey(), nil
}

func (c claims) toKey() authn.Key {
	key := authn.Key{
		ID:       c.Id,
		Issuer:   c.Issuer,
		Secret:   c.Subject,
		IssuedAt: time.Unix(c.IssuedAt, 0).UTC(),
	}
	if c.ExpiresAt != 0 {
		key.ExpiresAt = time.Unix(c.ExpiresAt, 0).UTC()
	}

	// Default type is 0.
	if c.Type != nil {
		key.Type = *c.Type
	}

	return key
}
