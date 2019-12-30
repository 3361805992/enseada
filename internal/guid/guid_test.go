// Copyright 2019 Enseada authors
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package guid

import (
	"testing"

	"github.com/enseadaio/enseada/internal/couch"
	"github.com/stretchr/testify/assert"
)

var testKind = couch.Kind("test")

func TestNew(t *testing.T) {
	guid := New("test", "test", testKind)
	assert.Equal(t, "test", guid.db)
	assert.Equal(t, "test", guid.id)
}

func TestNewWithRev(t *testing.T) {
	guid := NewWithRev("test", "test", testKind, "1")
	assert.Equal(t, "test", guid.db)
	assert.Equal(t, "test", guid.id)
	assert.Equal(t, "1", guid.rev)
}

func TestParseWithRev(t *testing.T) {
	guid, err := Parse("test://test?rev=1&kind=test")
	assert.NoError(t, err)
	assert.Equal(t, "test", guid.db)
	assert.Equal(t, "test", guid.id)
	assert.Equal(t, "1", guid.rev)
}

func TestParse(t *testing.T) {
	guid, err := Parse("test://test?kind=test")
	assert.NoError(t, err)
	assert.Equal(t, "test", guid.db)
	assert.Equal(t, "test", guid.id)
}

func TestParseInvalid(t *testing.T) {
	_, err := Parse("test")
	assert.Error(t, err)
	assert.Equal(t, "is missing database", err.Error())

	_, err = Parse("test://")
	assert.Error(t, err)
	assert.Equal(t, "is missing ID", err.Error())
}

func TestParseEmpty(t *testing.T) {
	_, err := Parse("")
	assert.Error(t, err)
	assert.Equal(t, "GUID can't be blank", err.Error())
}

func TestGUID_String(t *testing.T) {
	guid := New("test", "test", testKind)
	assert.Equal(t, "test://test?kind=test", guid.String())
}

func TestGUID_StringWithRev(t *testing.T) {
	guid := NewWithRev("test", "test", testKind, "1")
	assert.Equal(t, "test://test?kind=test&rev=1", guid.String())
}
