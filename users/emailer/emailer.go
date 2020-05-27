// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package emailer

import (
	"fmt"

	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/users"
)

var _ users.Emailer = (*emailer)(nil)

type emailer struct {
	resetURL  string
	verifyURL string
	agent     *email.Agent
	resetTmpl string
	verifyTmpl string 
}

// New creates new emailer utility
func New(reset_url, verify_url string, configReset *email.Config, configVerify *email.Config) (users.Emailer, error) {
	e, err := email.New(configReset)
	if err != nil {
		return nil, err
	}
	return &emailer{resetURL: reset_url, verifyURL: verify_url, agent: e, resetTmpl: configReset.Template, verifyTmpl: configVerify.Template}, nil
}

func (e *emailer) SendPasswordReset(To []string, host string, token string) errors.Error {
	url := fmt.Sprintf("%s%s?token=%s", host, e.resetURL, token)
	content := fmt.Sprintf("%s", url)
	return e.agent.Send(To, "", "Password reset", "", content, "", e.resetTmpl)
}

func (e *emailer) SendEmailVerification(To []string, host string, token string) errors.Error {
	url := fmt.Sprintf("%s%s?token=%s", host, e.verifyURL, token) // TO-DO
	content := fmt.Sprintf("%s", url)
	return e.agent.Send(To, "", "Email Verification", "", content, "", e.verifyTmpl)
}
