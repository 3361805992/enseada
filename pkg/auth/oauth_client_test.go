package auth

import (
	"github.com/enseadaio/enseada/internal/couch"
	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

const secret = "test"

func TestNewPrivateOAuthClient(t *testing.T) {
	client, err := NewOAuthClient("test", secret)
	assert.NoError(t, err)

	assert.Equal(t, couch.KindOAuthClient, client.Kind)
	assert.Equal(t, client.ID, client.GetID())
	assert.Equal(t, client.HashedSecret, client.GetHashedSecret())
	assert.Equal(t, client.RedirectURIs, client.GetRedirectURIs())
	assert.Equal(t, fosite.Arguments(client.GrantTypes), client.GetGrantTypes())
	assert.Equal(t, fosite.Arguments(client.GrantTypes), fosite.Arguments{"authorization_code"})
	assert.Equal(t, fosite.Arguments(client.ResponseTypes), client.GetResponseTypes())
	assert.Equal(t, fosite.Arguments(client.ResponseTypes), fosite.Arguments{"code"})
	assert.Equal(t, fosite.Arguments(client.Scopes), client.GetScopes())
	assert.Equal(t, client.Public, client.IsPublic())
	assert.Equal(t, fosite.Arguments(client.Audiences), client.GetAudience())

	err = bcrypt.CompareHashAndPassword(client.GetHashedSecret(), []byte(secret))
	assert.NoError(t, err)
}

func TestNewPrivateOAuthClientNoID(t *testing.T) {
	client, err := NewOAuthClient("", "")
	assert.Nil(t, client)
	assert.EqualError(t, err, "client ID cannot be empty")
}

func TestNewPrivateOAuthClientNoSecret(t *testing.T) {
	client, err := NewOAuthClient("test", "")
	assert.Nil(t, client)
	assert.EqualError(t, err, "client secret cannot be empty for non-public clients")
}

func TestNewPublicOAuthClient(t *testing.T) {
	client, err := NewOAuthClient("test", secret, OIDCPublic(true))
	assert.NoError(t, err)
	assert.Nil(t, client.HashedSecret)
}
