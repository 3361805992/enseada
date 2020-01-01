// Copyright 2019-2020 Enseada authors
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package middleware

import (
	"context"

	"github.com/ory/fosite"
)

type contextKey int

const (
	currentUserID contextKey = 1 + iota
	authStrategy
	scopes
)

func CurrentUserID(ctx context.Context) (string, bool) {
	g, ok := ctx.Value(currentUserID).(string)
	return g, ok
}

func WithCurrentUserID(ctx context.Context, g string) context.Context {
	return context.WithValue(ctx, currentUserID, g)
}

func Scopes(ctx context.Context) (fosite.Arguments, bool) {
	s, ok := ctx.Value(scopes).(fosite.Arguments)
	return s, ok
}

func WithScopes(ctx context.Context, s fosite.Arguments) context.Context {
	return context.WithValue(ctx, scopes, s)
}
