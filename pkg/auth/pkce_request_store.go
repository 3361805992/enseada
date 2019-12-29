package auth

import (
	"context"
	"errors"
	"github.com/enseadaio/enseada/internal/couch"
	"github.com/go-kivik/kivik"
	"github.com/labstack/echo"
	"github.com/ory/fosite"
)

type PKCERequestStore struct {
	data   *kivik.Client
	logger echo.Logger
}

func NewPKCERequestStore(data *kivik.Client, logger echo.Logger) *PKCERequestStore {
	return &PKCERequestStore{data: data, logger: logger}
}

func (r *PKCERequestStore) CreatePKCERequestSession(ctx context.Context, signature string, requester fosite.Requester) error {
	req := &OAuthRequest{}
	req.Merge(requester)
	db := r.data.DB(ctx, couch.OAuthDB)
	_, _, err := db.CreateDoc(ctx, &OAuthRequestWrapper{
		Kind: couch.KindPKCERequest,
		Sig:  signature,
		Req:  req,
	})
	return err
}

func (r *PKCERequestStore) GetPKCERequestSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	db := r.data.DB(ctx, couch.OAuthDB)
	rows, err := db.Find(ctx, couch.Query{
		"selector": couch.Query{
			"kind": couch.KindPKCERequest,
			"sig":  signature,
		},
	})
	if err != nil {
		return nil, err
	}

	var request OAuthRequestWrapper
	if rows.Next() {
		if err := rows.ScanDoc(&request); err != nil {
			return nil, err
		}
		session = request.Req.GetSession()
		return request.Req, nil
	}

	return nil, errors.New("pkce request not found")
}

func (r *PKCERequestStore) DeletePKCERequestSession(ctx context.Context, signature string) error {
	db := r.data.DB(ctx, couch.OAuthDB)
	rows, err := db.Find(ctx, couch.Query{
		"selector": couch.Query{
			"kind": couch.KindPKCERequest,
			"sig":  signature,
		},
	})
	if err != nil {
		return err
	}

	var request OAuthRequestWrapper
	if rows.Next() {
		if err := rows.ScanDoc(&request); err != nil {
			return err
		}
		_, err = db.Delete(ctx, request.ID, request.Rev)
		return err
	}
	return errors.New("pkce request not found")
}
