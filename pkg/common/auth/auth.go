/*
# Auth

This package initializes methods for functions which need special authentication to interact with APIs:

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/common/auth/auth.go
package auth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"
)

type JWTConfig *jwt.Config

type OAuthConfig *oauth2.Config
