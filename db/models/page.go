package models

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/utils"
	"strings"
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

	getPagesMeta, err = db.Prepare("SELECT id, sortValue, title, customrootdir FROM pages;")
	if err != nil {
		return err
	}

	getPageByID, err = db.Prepare("SELECT id, sortValue, title, customrootdir FROM pages WHERE id = ?;")
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

func NewPages(db *sql.DB) *Pages {
	return &Pages{
		db: db,
	}
}

type Pages struct {
	db *sql.DB
}

func (p *Pages) All() ([]Page, error) {
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
			return nil, err
		}
		pages = append(pages, page)
	}

	return pages, nil
}

func (p *Pages) Get(id int64) (*Page, error) {
	row := getPageByID.QueryRow(id)
	var page Page
	if err := readPage(row, &page); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &page, nil
}

func (p *Pages) Upsert(pages ...*Page) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		if err = tx.Rollback(); err != nil {
			log.Warn("failed to rollback transaction", "err", err)
		}
	}(tx)

	for _, page := range pages {
		if err = upsertPage(tx, page); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (p *Pages) Delete(pageID int64) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		if err = tx.Rollback(); err != nil {
			log.Warn("failed to rollback transaction", "err", err)
		}
	}(tx)

	_, err = tx.Exec(`DELETE FROM modifier_values WHERE modifier_id IN (
        SELECT id FROM modifiers WHERE page_id = ?
    )`, pageID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM modifiers WHERE page_id = ?`, pageID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM providers WHERE page_id = ?`, pageID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM dirs WHERE page_id = ?`, pageID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM pages WHERE id = ?`, pageID)
	if err != nil {
		return err
	}

	return tx.Commit()
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
	ID            int64      `json:"id"`
	Title         string     `json:"title" validate:"required,min=3,max=25"`
	SortValue     int        `json:"sort_value"`
	Providers     []Provider `json:"providers" validate:"required,min=1"`
	Modifiers     []Modifier `json:"modifiers"`
	Dirs          []string   `json:"dirs" validate:"required,min=1"`
	CustomRootDir string     `json:"custom_root_dir"`
}

func (p *Page) read(s scanner) error {
	return s.Scan(&p.ID, &p.SortValue, &p.Title, &p.CustomRootDir)
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

	p.Modifiers = make([]Modifier, 0)
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
		p.Modifiers = append(p.Modifiers, modifier)
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
	DYNASTY
)

type ModifierType int

const (
	DROPDOWN ModifierType = iota + 1
	MULTI
)

type Modifier struct {
	ID     int64             `json:"id"`
	Title  string            `json:"title"`
	Type   ModifierType      `json:"type"`
	Key    string            `json:"key"`
	Values map[string]string `json:"values"`
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

func upsertPage(tx *sql.Tx, page *Page) error {
	pageId := func() any {
		if page.ID == 0 {
			return nil
		}
		return page.ID
	}()

	var sortValue any
	if page.SortValue != 0 {
		sortValue = page.SortValue
	} else {
		row := tx.QueryRow("SELECT MAX(sortValue) FROM pages;")
		var val int
		if err := row.Scan(&val); err != nil {
			return err
		}

		sortValue = val + 1
	}

	result, err := tx.Exec(`INSERT INTO pages (id, title, customRootDir, sortValue) VALUES (?, ?, ?, ?) 
		ON CONFLICT(id) DO UPDATE SET title = excluded.title, customRootDir = excluded.customRootDir, sortValue = excluded.sortValue`,
		pageId, page.Title, page.CustomRootDir, sortValue)
	if err != nil {
		return err
	}

	if page.ID == 0 {
		pageID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		page.ID = pageID
	}

	_, err = tx.Exec("DELETE FROM providers WHERE page_id = ?", page.ID)
	if err != nil {
		return err
	}

	for _, provider := range page.Providers {
		_, err = tx.Exec(`INSERT INTO providers (page_id, provider) VALUES (?, ?) ON CONFLICT DO NOTHING;`, page.ID, provider)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec("DELETE FROM dirs WHERE page_id = ?", page.ID)
	if err != nil {
		return err
	}

	for _, dir := range page.Dirs {
		_, err = tx.Exec(`INSERT INTO dirs (page_id, dir) VALUES (?, ?) ON CONFLICT DO NOTHING;`, page.ID, dir)
		if err != nil {
			return err
		}
	}

	for _, modifier := range page.Modifiers {
		err = upsertModifier(tx, page.ID, &modifier)
		if err != nil {
			return err
		}
	}

	modifierIDs := utils.Map(page.Modifiers, func(t Modifier) any {
		return t.ID
	})
	placeholders := make([]string, len(modifierIDs))
	for i := range modifierIDs {
		placeholders[i] = "?"
	}
	query := fmt.Sprintf("DELETE FROM modifiers WHERE id NOT IN (%s) AND page_id = %d", strings.Join(placeholders, ","), page.ID)
	_, err = tx.Exec(query, modifierIDs...)
	if err != nil {
		return err
	}

	return nil
}

func upsertModifier(tx *sql.Tx, pageID int64, modifier *Modifier) error {
	result, err := tx.Exec(`INSERT INTO modifiers (id, page_id, title, type, key) 
		VALUES (?, ?, ?, ?, ?) 
		ON CONFLICT(id) DO UPDATE SET title = excluded.title, type = excluded.type, key = excluded.key`,
		modifier.ID, pageID, modifier.Title, modifier.Type, modifier.Key)
	if err != nil {
		return err
	}

	if modifier.ID < 0 {
		modifierID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		modifier.ID = modifierID
	}

	_, err = tx.Exec(`DELETE FROM modifier_values WHERE modifier_id = ?`, modifier.ID)
	if err != nil {
		return err
	}

	for k, v := range modifier.Values {
		_, err = tx.Exec(`INSERT INTO modifier_values (modifier_id, key, value) VALUES (?, ?, ?)`, modifier.ID, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
