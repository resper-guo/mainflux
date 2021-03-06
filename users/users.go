// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package users

import (
	"context"
	"regexp"
	"strings"

	"github.com/mainflux/mainflux/errors"
	"golang.org/x/net/idna"
)

const (
	minPassLen  = 8
	maxLocalLen = 64
	maxDomainLen  = 255
	maxTLDLen = 24 // longest TLD currently in existence

	atSeparator = "@"
	dotSeparator = "."
)

var (
	userRegexp    = regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+/=?^_`{|}~.-]+$")
	hostRegexp    = regexp.MustCompile("^[^\\s]+\\.[^\\s]+$")
	userDotRegexp = regexp.MustCompile("(^[.]{1})|([.]{1}$)|([.]{2,})")
)

// User represents a Mainflux user account. Each user is identified given its
// email and password.
type User struct {
	Email    string
	Password string
	Metadata map[string]interface{}
}

// Validate returns an error if user representation is invalid.
func (u User) Validate() errors.Error {
	if !isEmail(u.Email) {
		return ErrMalformedEntity
	}

	if len(u.Password) < minPassLen {
		return ErrMalformedEntity
	}

	return nil
}

// UserRepository specifies an account persistence API.
type UserRepository interface {
	// Save persists the user account. A non-nil error is returned to indicate
	// operation failure.
	Save(context.Context, User) errors.Error

	// Update updates the user metadata.
	UpdateUser(context.Context, User) errors.Error

	// RetrieveByID retrieves user by its unique identifier (i.e. email).
	RetrieveByID(context.Context, string) (User, errors.Error)

	// UpdatePassword updates password for user with given email
	UpdatePassword(_ context.Context, email, password string) errors.Error
}

func isEmail(email string) bool {
	if email == "" {
		return false
	}

	es := strings.Split(email, atSeparator)
	if len(es) != 2 {
		return false
	}
	local, host := es[0], es[1]

	if local == "" || len(local) > maxLocalLen {
		return false
	}

	hs := strings.Split(host, dotSeparator)
	if len(hs) != 2 {
		return false
	}
	domain, ext := hs[0], hs[1]

	if domain == "" || len(domain) > maxDomainLen {
		return false
	}
	if ext == "" || len(ext) > maxTLDLen {
		return false
	}

	punyLocal, err := idna.ToASCII(local)
	if err != nil {
		return false
	}
	punyHost, err := idna.ToASCII(host)
	if err != nil {
		return false
	}

	if userDotRegexp.MatchString(punyLocal) || !userRegexp.MatchString(punyLocal) || !hostRegexp.MatchString(punyHost) {
		return false
	}

	return true
}
