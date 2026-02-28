// Copyright (c) 2026 Reiner Pröls
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// SPDX-License-Identifier: MIT
//
// Author: Reiner Pröls

package database

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"bytemystery-com/passwordsafe/crypt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	_ "modernc.org/sqlite"
)

const (
	CURRENT_VERSION              = 1
	DEF_NUMBER_OF_CATEGORIES     = 15
	DEF_NUMBER_OF_ENTRIES        = 35
	DEF_NUMBER_OF_SEARCH_RESULTS = 25
)

type Db struct {
	sql *sql.DB
}

type CategoryDataType struct {
	Id        int64     `json:"id"`
	Name      string    `json:"n"`
	Timestamp time.Time `json:"t"`
}

type EntryDataType struct {
	Id         int64     `json:"id"`
	Name       string    `json:"n"`
	Icon       string    `json:"i"`
	CategoryId int64     `json:"c"`
	Timestamp  time.Time `json:"t"`
}

type EntryFieldData struct {
	Name       string `json:"l"`
	Value      string `json:"v"`
	IsPassword bool   `json:"p"`
}

type EntryFullDataType struct {
	EntryDataType `json:"e"`
	Fields        []*EntryFieldData `json:"f"`
}

func GetDBFile(name string) (string, error) {
	app := fyne.CurrentApp()
	root := app.Storage().RootURI()
	dbUri, err := storage.Child(root, name+".db")
	if err != nil {
		return "", err
	}
	return filepath.FromSlash(dbUri.Path()), nil
}

func NewDb() *Db {
	return &Db{}
}

func (d *Db) create(c *crypt.Crypt, pass []byte) error {
	_, err := d.sql.Exec(`CREATE TABLE crypt (id INTEGER PRIMARY KEY AUTOINCREMENT, key BLOB NOT NULL, saltsize INTEGER NOT NULL, noncesize INTEGER NOT NULL, keysize INTEGER NOT NULL, argon2time INTEGER NOT NULL, argon2memory INTEGER NOT NULL, argon2threads INTEGER NOT NULL, ts DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)`)
	if err != nil {
		return err
	}
	_, err = d.sql.Exec(`CREATE TRIGGER crypt_set_ts_update AFTER UPDATE ON crypt FOR EACH ROW BEGIN UPDATE crypt SET ts = CURRENT_TIMESTAMP WHERE id=OLD.id; END`)
	if err != nil {
		return err
	}

	k, err := c.CreateKey(pass)
	if err != nil {
		return err
	}

	p, err := d.sql.Prepare(`INSERT INTO crypt (key, saltsize, noncesize, keysize, argon2time, argon2memory, argon2threads) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer p.Close()
	_, err = p.Exec(k, c.Cfg.SaltSize, c.Cfg.NonceSize, c.Cfg.KeySize, c.Cfg.Argon2Time, c.Cfg.Argon2Memory, c.Cfg.Argon2Threads)
	if err != nil {
		return err
	}

	// Config
	_, err = d.sql.Exec(`CREATE TABLE config (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT DEFAULT "", value TEXT DEFAULT "", ts DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)`)
	if err != nil {
		return err
	}
	_, err = d.sql.Exec(`CREATE TRIGGER config_set_ts_update AFTER UPDATE ON config FOR EACH ROW BEGIN UPDATE config SET ts = CURRENT_TIMESTAMP WHERE id=OLD.id; END`)
	if err != nil {
		return err
	}
	_, err = d.sql.Exec(fmt.Sprintf(`INSERT INTO config (name, value) VALUES ('version', '%d')`, CURRENT_VERSION))
	if err != nil {
		return err
	}

	// Category
	_, err = d.sql.Exec(`CREATE TABLE category (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT DEFAULT '' NOT NULL, ts DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)`)
	if err != nil {
		return err
	}
	_, err = d.sql.Exec(`CREATE UNIQUE INDEX idx_category_name ON category(name ASC)`)
	if err != nil {
		return err
	}
	_, err = d.sql.Exec(`CREATE TRIGGER category_set_ts_update AFTER UPDATE ON category FOR EACH ROW BEGIN UPDATE category SET ts = CURRENT_TIMESTAMP WHERE id=OLD.id; END`)
	if err != nil {
		return err
	}
	_, err = d.sql.Exec(`INSERT INTO category (name) VALUES ('test')`)
	if err != nil {
		return err
	}

	// Entry
	_, err = d.sql.Exec(`CREATE TABLE entry (id INTEGER PRIMARY KEY AUTOINCREMENT, categoryId INTEGER NOT NULL, name TEXT DEFAULT '' NOT NULL, icon TEXT DEFAULT '' NOT NULL, data BLOB, ts DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (categoryId) REFERENCES category(id) ON DELETE CASCADE)`)
	if err != nil {
		return err
	}
	_, err = d.sql.Exec(`CREATE UNIQUE INDEX idx_entry_id_name ON entry(categoryId, name ASC)`)
	if err != nil {
		return err
	}
	_, err = d.sql.Exec(`CREATE TRIGGER entry_set_ts_update AFTER UPDATE ON entry FOR EACH ROW BEGIN UPDATE entry SET ts = CURRENT_TIMESTAMP WHERE id=OLD.id; END`)
	if err != nil {
		return err
	}

	/*
		// Binary
		_, err = d.sql.Exec(`CREATE TABLE binary (id INTEGER PRIMARY KEY AUTOINCREMENT, entryId INTEGER NOT NULL, data BLOB NOT NULL, ts DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)`)
		if err != nil {
			return err
		}
		_, err = d.sql.Exec(`CREATE UNIQUE INDEX idx_binary_id ON binary(entryId)`)
		if err != nil {
			return err
		}
		_, err = d.sql.Exec(`CREATE TRIGGER binary_set_ts_update AFTER UPDATE ON binary FOR EACH ROW BEGIN UPDATE binary SET ts = CURRENT_TIMESTAMP WHERE id=OLD.id; END`)
		if err != nil {
			return err
		}
	*/

	return nil
}

func (d *Db) update(oldVer int) error {
	return nil
}

func (d *Db) Open(file string, pass []byte, c *crypt.Crypt) error {
	d.Close()
	b, err := sql.Open("sqlite", fmt.Sprintf("file:%s?pragma=foreign_keys(1)", file))
	if err != nil {
		return err
	}
	d.sql = b
	v := ""
	q, err := d.sql.Query(`SELECT value from config WHERE name = 'version'`)
	if err != nil {
		if pass != nil && c != nil {
			err = d.create(c, pass)
		}
		return err
	}
	defer q.Close()
	if !q.Next() {
		return errors.New("no version information")
	}
	err = q.Scan(&v)
	if err != nil {
		return err
	}
	ver, err := strconv.Atoi(v)
	if err != nil {
		return err
	}
	if ver == CURRENT_VERSION {
		return nil
	} else {
		return d.update(ver)
	}
}

func (d *Db) IsOpen() bool {
	return d.sql != nil
}

func (d *Db) Close() {
	if d.sql != nil {
		d.sql.Close()
		d.sql = nil
	}
}

func (d *Db) ChangePassword(passOld []byte, passNew []byte, c *crypt.Crypt) error {
	key, id, err := d.getMasterKeyInternal(passOld, c)
	if err != nil {
		return err
	}
	p, err := d.sql.Prepare(`UPDATE crypt SET key=? WHERE id=?`)
	if err != nil {
		return err
	}
	defer p.Close()
	keyNew, err := c.EncryptBinary(passNew, key)
	if err != nil {
		crypt.ErasePassword(key)
		return err
	}
	_, err = p.Exec(keyNew, id)
	crypt.ErasePassword(key)
	return err
}

func (d *Db) GetMasterKey(pass []byte, c *crypt.Crypt) ([]byte, error) {
	b, _, err := d.getMasterKeyInternal(pass, c)
	return b, err
}

func (d *Db) getMasterKeyInternal(pass []byte, c *crypt.Crypt) ([]byte, int64, error) {
	q := d.sql.QueryRow(`SELECT id, key FROM crypt`)
	var k []byte
	var id int64
	err := q.Scan(&id, &k)
	if err != nil {
		return nil, -1, err
	}
	x, err := c.Decrypt(pass, k)
	return x, id, err
}

func (d *Db) GetCryptCfg() (*crypt.CryptCfg, error) {
	cfg := crypt.CryptCfg{}
	q := d.sql.QueryRow(`SELECT saltsize, noncesize, keysize, argon2time, argon2memory, argon2threads FROM crypt`)
	err := q.Scan(&cfg.SaltSize, &cfg.NonceSize, &cfg.KeySize, &cfg.Argon2Time, &cfg.Argon2Memory, &cfg.Argon2Threads)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (d *Db) AddCategory(data *CategoryDataType) error {
	p, err := d.sql.Prepare(`INSERT INTO category (name) VALUES (?)`)
	if err != nil {
		return err
	}
	defer p.Close()
	r, err := p.Exec(data.Name)
	if err != nil {
		return err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return err
	}
	data.Id = id
	return nil
}

func (d *Db) UpdateCategory(data *CategoryDataType) error {
	p, err := d.sql.Prepare(`UPDATE category SET name=? WHERE id=?`)
	if err != nil {
		return err
	}
	defer p.Close()
	_, err = p.Exec(data.Name, data.Id)
	return err
}

func (d *Db) DeleteCategory(id int64) error {
	p, err := d.sql.Prepare(`DELETE FROM category WHERE id=?`)
	if err != nil {
		return err
	}
	defer p.Close()
	_, err = p.Exec(id)
	if err != nil {
		return err
	}
	p1, err := d.sql.Prepare(`DELETE FROM entry WHERE categoryId=?`)
	if err != nil {
		return err
	}
	defer p1.Close()
	_, err = p1.Exec(id)

	return err
}

func (d *Db) GetEntriesInCategory(id int64) (int, error) {
	p, err := d.sql.Prepare(`SELECT COUNT(*) FROM entry WHERE categoryId=?`)
	if err != nil {
		return -1, err
	}
	defer p.Close()
	r := p.QueryRow(id)
	count := 0
	err = r.Scan(&count)
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (d *Db) GetNumberOfCategories() (int, error) {
	if !d.IsOpen() {
		return 0, errors.New("database is not open")
	}
	r := d.sql.QueryRow(`SELECT COUNT(*) FROM category`)
	count := 0
	err := r.Scan(&count)
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (d *Db) GetNumberOfEntries() (int, error) {
	if !d.IsOpen() {
		return 0, errors.New("database is not open")
	}
	r := d.sql.QueryRow(`SELECT COUNT(*) FROM entry`)
	count := 0
	err := r.Scan(&count)
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (d *Db) GetCategories() ([]*CategoryDataType, error) {
	if d.sql == nil {
		return nil, errors.New("database not open")
	}
	q, err := d.sql.Query(`SELECT id, name, ts FROM category ORDER BY name COLLATE NOCASE ASC`)
	if err != nil {
		return nil, err
	}
	defer q.Close()
	list := make([]*CategoryDataType, 0, DEF_NUMBER_OF_CATEGORIES)
	for q.Next() {
		data := CategoryDataType{}
		err := q.Scan(&data.Id, &data.Name, &data.Timestamp)
		if err == nil {
			list = append(list, &data)
		}
	}
	return list, nil
}

func (d *Db) GetCategoryName(categoryId int64) (string, error) {
	if d.sql == nil {
		return "", errors.New("database not open")
	}
	p, err := d.sql.Prepare(`SELECT name FROM category WHERE id=?`)
	if err != nil {
		return "", err
	}
	defer p.Close()
	r := p.QueryRow(categoryId)
	name := ""
	err = r.Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}

func (d *Db) GetEntries(categoryId int64) ([]*EntryDataType, error) {
	if d.sql == nil {
		return nil, errors.New("database not open")
	}
	p, err := d.sql.Prepare(`SELECT id, name, categoryId, icon, ts FROM entry WHERE categoryId = ? ORDER BY name COLLATE NOCASE ASC`)
	if err != nil {
		return nil, err
	}
	defer p.Close()
	r, err := p.Query(categoryId)
	if err != nil {
		return nil, err
	}
	list := make([]*EntryDataType, 0, DEF_NUMBER_OF_ENTRIES)

	for r.Next() {
		data := EntryDataType{}
		err := r.Scan(&data.Id, &data.Name, &data.CategoryId, &data.Icon, &data.Timestamp)
		if err == nil {
			list = append(list, &data)
		}
	}
	return list, nil
}

func (d *Db) fieldsToBin(data []*EntryFieldData, pass []byte, cry *crypt.Crypt, useJson bool) ([]byte, error) {
	var bData []byte
	var err error
	if useJson {
		bData, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	} else {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err = enc.Encode(data)
		if err != nil {
			return nil, err
		}
		bData = buf.Bytes()
	}
	// Crypt
	k, err := d.GetMasterKey(pass, cry)
	if err != nil {
		return nil, err
	}
	return cry.EncryptBinary(k, bData)
}

func (d *Db) binToFields(data []byte, pass []byte, cry *crypt.Crypt, useJson bool) ([]*EntryFieldData, error) {
	// Crypt
	k, err := d.GetMasterKey(pass, cry)
	if err != nil {
		return nil, err
	}
	b, err := cry.Decrypt(k, data)
	if err != nil {
		return nil, err
	}

	list := []*EntryFieldData{}
	if useJson {
		err := json.Unmarshal(b, &list)
		if err != nil {
			return nil, err
		}
	} else {
		dec := gob.NewDecoder(bytes.NewReader(b))
		err = dec.Decode(&list)
		if err != nil {
			return nil, err
		}
	}
	return list, err
}

func (d *Db) AddEntry(data *EntryFullDataType, pass []byte, cry *crypt.Crypt) (int64, error) {
	if d.sql == nil {
		return 0, errors.New("database not open")
	}
	p, err := d.sql.Prepare(`INSERT INTO entry (categoryId, name, icon, data) VALUES(?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer p.Close()
	b, err := d.fieldsToBin(data.Fields, pass, cry, true)
	if err != nil {
		return 0, err
	}
	r, err := p.Exec(data.CategoryId, data.Name, data.Icon, b)
	if err != nil {
		return 0, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}
	data.Id = id
	return data.Id, nil
}

func (d *Db) UpdateEntry(data *EntryFullDataType, pass []byte, cry *crypt.Crypt) error {
	if d.sql == nil {
		return errors.New("database not open")
	}
	p, err := d.sql.Prepare(`UPDATE entry SET categoryId=?, name=?, icon=?, data=? WHERE id=?`)
	if err != nil {
		return err
	}
	defer p.Close()
	b, err := d.fieldsToBin(data.Fields, pass, cry, true)
	if err != nil {
		return err
	}
	_, err = p.Exec(data.CategoryId, data.Name, data.Icon, b, data.Id)
	if err != nil {
		return err
	}
	return nil
}

func (d *Db) GetEntry(id int64, pass []byte, cry *crypt.Crypt) (*EntryFullDataType, error) {
	if d.sql == nil {
		return nil, errors.New("database not open")
	}
	p, err := d.sql.Prepare(`SELECT name, categoryId, icon, data, ts FROM entry WHERE id = ?`)
	if err != nil {
		return nil, err
	}
	defer p.Close()
	r := p.QueryRow(id)
	data := EntryFullDataType{}
	var b []byte
	err = r.Scan(&data.Name, &data.CategoryId, &data.Icon, &b, &data.Timestamp)
	if err != nil {
		return nil, err
	}
	l, err := d.binToFields(b, pass, cry, true)
	if err != nil {
		return nil, err
	}
	data.Fields = l
	data.Id = id
	return &data, nil
}

func (d *Db) GetEntryBase(id int64) (*EntryDataType, error) {
	if d.sql == nil {
		return nil, errors.New("database not open")
	}
	p, err := d.sql.Prepare(`SELECT name, categoryId, icon, ts FROM entry WHERE id = ?`)
	if err != nil {
		return nil, err
	}
	defer p.Close()
	r := p.QueryRow(id)
	data := EntryDataType{}
	err = r.Scan(&data.Name, &data.CategoryId, &data.Icon, &data.Timestamp)
	if err != nil {
		return nil, err
	}
	data.Id = id
	return &data, nil
}

func (d *Db) Search(text string) ([]*EntryDataType, error) {
	if !d.IsOpen() {
		return nil, errors.New("database not open")
	}
	p, err := d.sql.Prepare(`SELECT id, name, categoryId, icon, ts FROM entry WHERE name LIKE ? COLLATE NOCASE ORDER BY name COLLATE NOCASE ASC`)
	if err != nil {
		return nil, err
	}
	defer p.Close()
	r, err := p.Query(text)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	list := make([]*EntryDataType, 0, DEF_NUMBER_OF_SEARCH_RESULTS)
	for r.Next() {
		e := EntryDataType{}
		err := r.Scan(&e.Id, &e.Name, &e.CategoryId, &e.Icon, &e.Timestamp)
		if err != nil {
			return nil, err
		}
		list = append(list, &e)
	}
	return list, nil
}

func (d *Db) DeleteEntry(id int64) error {
	p, err := d.sql.Prepare(`DELETE FROM entry WHERE id=?`)
	if err != nil {
		return err
	}
	defer p.Close()
	_, err = p.Exec(id)
	return err
}

/*
func (d *Db) ConverToJson(pass []byte, crypt *crypt.Crypt) error {
	tx, _ := d.sql.Begin()
	defer tx.Commit()

	if !d.IsOpen() {
		return errors.New("database not open")
	}
	pu, err := tx.Prepare(`UPDATE entry SET data = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer pu.Close()

	r, err := tx.Query(`SELECT id, data FROM entry`)
	if err != nil {
		return err
	}
	defer r.Close()
	var errResult error
	for r.Next() {
		var data []byte
		id := 0
		err := r.Scan(&id, &data)
		if err == nil {
			entryFields, err := d.binToFields(data, pass, crypt, false)
			if err == nil {
				data, err = d.fieldsToBin(entryFields, pass, crypt, true)
				if err == nil {
					_, err := pu.Exec(data, id)
					if err != nil {
						errResult = err
					}
				} else {
					errResult = err
				}
			} else {
				errResult = err
			}
		} else {
			errResult = err
		}
	}
	return errResult
}
*/

func (d *Db) ExportToJson(pass []byte, crypt *crypt.Crypt) (string, error) {
	if !d.IsOpen() {
		return "", errors.New("database not open")
	}
	type AllData struct {
		Categories []CategoryDataType
		Entries    []EntryFullDataType
	}
	var allData AllData

	q, err := d.sql.Query(`SELECT id, name, ts FROM category ORDER BY id ASC`)
	if err != nil {
		return "", err
	}
	defer q.Close()
	allData.Categories = make([]CategoryDataType, 0, DEF_NUMBER_OF_CATEGORIES)
	for q.Next() {
		data := CategoryDataType{}
		err := q.Scan(&data.Id, &data.Name, &data.Timestamp)
		if err != nil {
			return "", err
		}
		allData.Categories = append(allData.Categories, data)
	}

	r, err := d.sql.Query(`SELECT id, categoryId, name, icon, data, ts data FROM entry ORDER BY id ASC`)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for r.Next() {
		entry := EntryFullDataType{}
		var data []byte
		err := r.Scan(&entry.Id, &entry.CategoryId, &entry.Name, &entry.Icon, &data, &entry.Timestamp)
		if err != nil {
			return "", err
		}
		entry.Fields, err = d.binToFields(data, pass, crypt, true)
		if err != nil {
			return "", err
		}
		allData.Entries = append(allData.Entries, entry)
	}
	data, err := json.MarshalIndent(allData, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (d *Db) GetLastWrite() (time.Time, error) {
	if !d.IsOpen() {
		return time.Time{}, errors.New("database not open")
	}
	r := d.sql.QueryRow(`SELECT MAX(ts) FROM (SELECT ts FROM category UNION ALL SELECT ts FROM entry UNION ALL SELECT ts from crypt)`)
	var str string
	err := r.Scan(&str)
	if err != nil {
		return time.Time{}, err
	}
	t, err := time.Parse("2006-01-02 15:04:05", str)
	return t, err
}
