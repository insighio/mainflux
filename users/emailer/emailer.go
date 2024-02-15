// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package emailer

import (
	"fmt"

	"github.com/absmach/magistrala/internal/email"
	"github.com/absmach/magistrala/users"
)

var _ users.Emailer = (*emailer)(nil)

type emailer struct {
	resetURL   string
	verifyURL  string
	agent      *email.Agent
	resetTmpl  string
	verifyTmpl string
}

// New creates new emailer utility.
func New(reset_url string, verify_url string, configReset *email.Config, configVerify *email.Config) (users.Emailer, error) {
	e, err := email.New(configReset)
	return &emailer{resetURL: reset_url, verifyURL: verify_url, agent: e, resetTmpl: configReset.Template, verifyTmpl: configVerify.Template}, err
}

func (e *emailer) SendPasswordReset(to []string, host, user, token string) error {
	url := fmt.Sprintf("%s%s?token=%s", host, e.resetURL, token)
	content := fmt.Sprintf("%s", url)
	return e.agent.Send(to, "", "Password reset", "", user, content, "", e.resetTmpl)
}

func (e *emailer) SendEmailVerification(To []string, host string, user, token string) error {
	url := fmt.Sprintf("%s%s?token=%s", host, e.verifyURL, token)
	content := fmt.Sprintf("%s", url)
	return e.agent.Send(To, "", "Email Verification", "", user, content, "", e.verifyTmpl)
}
