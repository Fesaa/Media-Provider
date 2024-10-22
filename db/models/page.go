package models

import (
	"database/sql"
	"errors"
	"github.com/Fesaa/Media-Provider/log"
)

var (
	getPagesMeta            *sql.Stmt
	getPageByID             *sql.Stmt
	getPageProviders        *sql.Stmt
	getPageDirs             *sql.Stmt
	getModifiersMetaForPage *sql.Stmt
	getModifiersValues      *sql.Stmt
)

func initPages(db *sql.DB) error {
	var err error

	getPagesMeta, err = db.Prepare("SELECT id, title, customrootdir FROM pages;")
	if err != nil {
		return err
	}

	getPageByID, err = db.Prepare("SELECT id, title, customrootdir FROM pages WHERE id = ?;")
	if err != nil {
		return err
	}

	getPageProviders, err = db.Prepare("SELECT provider FROM providers WHERE page_id = ?;")
	if err != nil {
		return err
	}

	getPageDirs, err = db.Prepare("SELECT dir FROM dirs WHERE page_id = ?;")
	if err != nil {
		return err
	}

	getModifiersMetaForPage, err = db.Prepare("SELECT id,title,type,key FROM modifiers WHERE page_id = ?;")
	if err != nil {
		return err
	}

	getModifiersValues, err = db.Prepare("SELECT key,value FROM modifier_values WHERE modifier_id = ?;")
	if err != nil {
		return err
	}

	return nil
}

func GetPages() ([]Page, error) {
	rows, err := getPagesMeta.Query()
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		if err = rows.Close(); err != nil {
			log.Warn("failed to close rows", "err", err)
		}
	}(rows)

	pages := make([]Page, 0)
	for rows.Next() {
		var page Page
		if err = readPage(rows, &page); err != nil {
			log.Error("failed to read page", "err", err)
			return nil, err
		}
		pages = append(pages, page)
	}

	return pages, nil
}

func GetPage(id int64) (*Page, error) {
	row := getPageByID.QueryRow(id)
	var page Page
	if err := readPage(row, &page); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		log.Error("failed to read page", "id", id, "err", err)
		return nil, err
	}

	return &page, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func readPage(s scanner, p *Page) error {
	if err := p.read(s); err != nil {
		return err
	}

	rows, err := getPageProviders.Query(p.ID)
	if err != nil {
		return err
	}
	if err = p.readProviders(rows); err != nil {
		return err
	}

	rows, err = getPageDirs.Query(p.ID)
	if err != nil {
		return err
	}
	if err = p.readDirs(rows); err != nil {
		return err
	}

	rows, err = getModifiersMetaForPage.Query(p.ID)
	if err != nil {
		return err
	}
	if err = p.readModifiers(rows); err != nil {
		return err
	}

	return nil
}

type Page struct {
	ID            int64               `json:"id"`
	Title         string              `json:"title" validate:"required,min=3,max=25"`
	Providers     []Provider          `json:"providers" validate:"required,min=1"`
	Modifiers     map[string]Modifier `json:"modifiers"`
	Dirs          []string            `json:"dirs" validate:"required,min=1"`
	CustomRootDir string              `json:"custom_root_dir"`
}

func (p *Page) read(s scanner) error {
	return s.Scan(&p.ID, &p.Title, &p.CustomRootDir)
}

func (p *Page) readProviders(rows *sql.Rows) error {
	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			log.Warn("failed to close rows", "err", err)
		}
	}(rows)

	p.Providers = make([]Provider, 0)
	for rows.Next() {
		var provider Provider
		err := rows.Scan(&provider)
		if err != nil {
			return err
		}
		p.Providers = append(p.Providers, provider)
	}
	return nil
}

func (p *Page) readDirs(rows *sql.Rows) error {
	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			log.Warn("failed to close rows", "err", err)
		}
	}(rows)

	p.Dirs = make([]string, 0)
	for rows.Next() {
		var dir string
		if err := rows.Scan(&dir); err != nil {
			return err
		}
		p.Dirs = append(p.Dirs, dir)
	}
	return nil
}

func (p *Page) readModifiers(rows *sql.Rows) error {
	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			log.Warn("failed to close rows", "err", err)
		}
	}(rows)

	p.Modifiers = make(map[string]Modifier)
	for rows.Next() {
		var modifier Modifier
		err := rows.Scan(&modifier.ID, &modifier.Title, &modifier.Type, &modifier.Key)
		if err != nil {
			return err
		}

		valueRows, err := getModifiersValues.Query(modifier.ID)
		if err != nil {
			return err
		}

		if err = modifier.readValues(valueRows); err != nil {
			return err
		}
		p.Modifiers[modifier.Key] = modifier
	}

	return nil
}

type Provider int

const (
	SUKEBEI Provider = iota + 1
	NYAA
	YTS
	LIME
	SUBSPLEASE
	MANGADEX
	WEBTOON
)

type ModifierType int

const (
	DROPDOWN ModifierType = iota + 1
	MULTI
)

type Modifier struct {
	ID     int64             `json:"id"`
	Title  string            `yaml:"title" json:"title"`
	Type   ModifierType      `yaml:"type" json:"type"`
	Key    string            `yaml:"key" json:"key"`
	Values map[string]string `yaml:"values" json:"values"`
}

func (m *Modifier) readValues(rows *sql.Rows) error {
	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			log.Warn("failed to close rows", "err", err)
		}
	}(rows)

	m.Values = make(map[string]string)

	for rows.Next() {
		var key string
		var value string

		if err := rows.Scan(&key, &value); err != nil {
			return err
		}

		m.Values[key] = value
	}

	return nil
}
