// Copyright 2019 Enseada authors
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package maven

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/enseadaio/enseada/internal/couch"
	"github.com/go-kivik/kivik"
)

const baseMetadataFile = `
<?xml version="1.0" encoding="UTF-8"?>
<metadata>
	<groupId>{{ .GroupId }}</groupId>
	<artifactId>{{ .ArtifactId }}</artifactId>
	<versioning>
		<versions></versions>
		<lastUpdated>{{ .TimeStamp }}</lastUpdated>
	</versioning>
</metadata>
`

type Repo struct {
	ID          string     `json:"_id,omitempty"`
	Rev         string     `json:"_rev,omitempty"`
	GroupID     string     `json:"group_id"`
	ArtifactID  string     `json:"artifact_id"`
	StoragePath string     `json:"storage_path"`
	Files       []string   `json:"files"`
	Kind        couch.Kind `json:"kind"`
}

func NewRepo(groupID string, artifactID string) Repo {
	group := strings.ReplaceAll(groupID, ".", "/")
	return Repo{
		ID:          repoID(groupID, artifactID),
		GroupID:     groupID,
		ArtifactID:  artifactID,
		StoragePath: strings.Join([]string{group, artifactID}, "/"),
		Kind:        couch.KindRepository,
	}
}

func (m *Maven) ListRepos(ctx context.Context, selector couch.Query) ([]*Repo, error) {
	db := m.data.DB(ctx, couch.MavenDB)
	s := couch.Query{
		"kind": couch.KindRepository,
	}
	if len(selector) > 0 {
		delete(selector, "kind")
		for k, v := range selector {
			s[k] = v
		}

	}

	rows, err := db.Find(ctx, couch.Query{
		"selector": s,
	})

	if err != nil {
		return nil, err
	}

	repos := make([]*Repo, 0)
	for rows.Next() {
		var repo Repo
		if err := rows.ScanDoc(&repo); err != nil {
			return nil, err
		}
		repos = append(repos, &repo)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return repos, nil
}

func (m *Maven) GetRepo(ctx context.Context, id string) (*Repo, error) {
	db := m.data.DB(ctx, couch.MavenDB)
	row := db.Get(ctx, id)
	repo := &Repo{}
	if err := row.ScanDoc(repo); err != nil {
		if kivik.StatusCode(err) == kivik.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	return repo, nil
}

func (m *Maven) GetRepoByCoordinates(ctx context.Context, groupID string, artifactID string) (*Repo, error) {
	return m.GetRepo(ctx, repoID(groupID, artifactID))
}

func (m *Maven) GetRepoByFile(ctx context.Context, path string) (*Repo, error) {
	return m.FindRepo(ctx, couch.Query{
		"files": couch.Query{
			"$elemMatch": couch.Query{
				"$eq": path,
			},
		},
	})
}

func (m *Maven) FindRepo(ctx context.Context, selector couch.Query) (*Repo, error) {
	s := couch.Query{
		"kind": couch.KindRepository,
	}

	if len(selector) > 0 {
		delete(selector, "kind")
		for k, v := range selector {
			s[k] = v
		}

	}

	db := m.data.DB(ctx, couch.MavenDB)
	rows, err := db.Find(ctx, couch.Query{
		"selector": s,
	})
	if err != nil {
		return nil, err
	}

	if rows.Next() {
		repo := new(Repo)
		if err := rows.ScanDoc(repo); err != nil {
			return nil, err
		}
		return repo, nil
	}

	return nil, nil
}

func (m *Maven) SaveRepo(ctx context.Context, repo *Repo) error {
	db := m.data.DB(ctx, couch.MavenDB)
	if repo.Rev == "" {
		_, rev, err := db.GetMeta(ctx, repo.ID)
		if err != nil {
			return err
		}

		repo.Rev = rev
	}

	rev, err := db.Put(ctx, repo.ID, repo)
	if err != nil {
		return err
	}
	repo.Rev = rev
	return err
}

func (m *Maven) DeleteRepo(ctx context.Context, id string) (*Repo, error) {
	db := m.data.DB(ctx, couch.MavenDB)
	repo, err := m.GetRepo(ctx, id)
	if err != nil || repo == nil {
		return nil, err
	}

	if err := m.ClearRepoStorage(ctx, repo); err != nil {
		return nil, err
	}

	rev, err := db.Delete(ctx, repo.ID, repo.Rev)
	if err != nil {
		return nil, err
	}

	repo.Rev = rev
	return repo, nil
}

func repoID(groupID string, artifactID string) string {
	return strings.Join([]string{groupID, artifactID}, ":")
}

func (m *Maven) InitRepo(ctx context.Context, repo *Repo) error {
	db := m.data.DB(ctx, couch.MavenDB)

	m.Logger.Infof("Initializing repo %s", repo.ID)
	err := save(ctx, db, repo)
	if err != nil {
		return err
	}

	m.Logger.Infof("Created repo %s", repo.ID)
	t, err := template.New("maven-metadata.xml").Parse(baseMetadataFile)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]interface{}{
		"GroupId":    repo.GroupID,
		"ArtifactId": repo.ArtifactID,
		"TimeStamp":  time.Now().Unix(),
	})
	if err != nil {
		return err
	}

	m.Logger.Infof("Creating file %s", t.ParseName)
	file := &RepoFile{
		Repo:     repo,
		Filename: t.ParseName,
		content:  buf.Bytes(),
	}

	md5sum := &RepoFile{
		Repo:     repo,
		Filename: fmt.Sprintf("%s.md5", t.ParseName),
		content:  []byte(fmt.Sprintf("%x", md5.Sum(file.content))),
	}

	sha1sum := &RepoFile{
		Repo:     repo,
		Filename: fmt.Sprintf("%s.sha1", t.ParseName),
		content:  []byte(fmt.Sprintf("%x", sha1.Sum(file.content))),
	}

	path := filePath(file)
	repo.Files = append(repo.Files, path)
	err = m.PutFile(ctx, path, file.content)
	if err != nil {
		return err
	}

	path = filePath(md5sum)
	repo.Files = append(repo.Files, path)
	err = m.PutFile(ctx, path, md5sum.content)
	if err != nil {
		return err
	}

	path = filePath(sha1sum)
	repo.Files = append(repo.Files, path)
	err = m.PutFile(ctx, path, sha1sum.content)
	if err != nil {
		return err
	}

	return save(ctx, db, repo)
}

func save(ctx context.Context, db *kivik.DB, repo *Repo) error {
	rev, err := db.Put(ctx, repo.ID, repo)
	if err != nil {
		switch kivik.StatusCode(err) {
		case kivik.StatusConflict:
			return ErrRepoAlreadyPresent
		default:
			return err
		}
	}
	repo.Rev = rev
	return nil
}
