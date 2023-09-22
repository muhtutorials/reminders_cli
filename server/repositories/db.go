package repositories

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/muhtutorials/reminders_cli/server/models"
	"io"
	"log"
	"os"
)

type dbConfig struct {
	ID       int    `json:"id"`
	Checksum string `json:"checksum"`
}

// DB represents the application server database (json file)
type DB struct {
	dbPath    string
	dbCfgPath string
	cfg       dbConfig
	db        []byte
}

func NewDB(dbPath, dbCfgPath string) *DB {
	db := &DB{
		dbPath:    dbPath,
		dbCfgPath: dbCfgPath,
	}
	return db
}

// Start starts and initializes the file database
func (db *DB) Start() error {
	bts, err := db.readDBFile(db.dbCfgPath)
	if err != nil {
		return models.WrapError("could not read db config contents", err)
	}
	var cfg dbConfig
	if len(bts) == 0 {
		bts = []byte("{}")
	}
	err = json.Unmarshal(bts, &cfg)
	if err != nil {
		return models.WrapError("could not unmarshal db config", err)
	}

	bts, err = db.readDBFile(db.dbPath)
	if err != nil {
		return models.WrapError("could not read db contents", err)
	}
	db.db = bts

	if db.cfg.Checksum == "" {
		checksum, err := genChecksum(bytes.NewReader(bts))
		if err != nil {
			return err
		}
		cfg.Checksum = checksum
	}
	db.cfg = cfg
	return nil
}

// readFile reads the contents of a db file
func (db *DB) readDBFile(path string) ([]byte, error) {
	dbFile, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if errors.Is(err, os.ErrNotExist) {
		dbFile, err = os.Create(path)
	}
	if err != nil {
		return nil, models.WrapError("could not open or create db file", err)
	}
	return io.ReadAll(dbFile)
}

// Read fetches a list of reminders by given ids
func (db *DB) Read(bts []byte) (int, error) {
	n, err := bytes.NewReader(db.db).Read(bts)
	if err != nil && err != io.EOF {
		return 0, models.WrapError("could not read db file bytes", err)
	}
	return n, nil
}

// Write writes a list of reminders to DB
func (db *DB) Write(bts []byte) (int, error) {
	bts = append(bts, '\n')

	checksum, err := genChecksum(bytes.NewReader(bts))
	if err != nil {
		return 0, err
	}
	if db.cfg.Checksum == checksum {
		return 0, nil
	}
	db.cfg.Checksum = checksum

	if err := db.writeDBCfg(); err != nil {
		return 0, err
	}
	n, err := db.writeFile(db.dbPath, bts)
	if err != nil {
		return 0, err
	}
	db.db = bts
	return n, nil
}

// Size retrieves the current size of the database
func (db *DB) Size() int {
	if len(db.db) == 0 {
		db.db = []byte("[]")
	}
	return len(db.db)
}

// GenerateID generates the next AUTOINCREMENT id for a reminder
func (db *DB) GenerateID() int {
	db.cfg.ID++
	return db.cfg.ID
}

// Stop shuts down properly the file database by saving metadata to config file
func (db *DB) Stop() error {
	log.Println("shutting down the database")
	_, errDB := os.Open(db.dbPath)
	_, errDBCfg := os.Open(db.dbCfgPath)
	if errors.Is(errDB, os.ErrNotExist) {
		_, err := db.writeFile(db.dbPath, db.db)
		if err != nil {
			return err
		}
	}
	if errors.Is(errDBCfg, os.ErrNotExist) {
		if err := db.writeDBCfg(); err != nil {
			return err
		}
	}
	log.Println("database was successfully shut down")
	return nil
}

// genCheckSum generates checksum for a reader
func genChecksum(r io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, r); err != nil {
		return "", models.WrapError("could not copy db contents", err)
	}
	sum := hash.Sum(nil)
	return fmt.Sprintf("%x", sum), nil
}

// writeDBCfg writes db config to file
func (db *DB) writeDBCfg() error {
	bts, err := json.Marshal(db.cfg)
	if err != nil {
		return models.WrapError("could not marshal db config", err)
	}
	bts = append(bts, '\n')
	_, err = db.writeFile(db.dbCfgPath, bts)
	if err != nil {
		return models.WrapError("could not write to db cfg file", err)
	}
	return nil
}

func (db *DB) writeFile(path string, bts []byte) (int, error) {
	dbFile, err := os.Create(path)
	if err != nil {
		return 0, models.WrapError("could not create file", err)
	}
	defer func(dbFile *os.File) {
		err := dbFile.Close()
		if err != nil {
			log.Printf("could not close file '%s': %v", dbFile.Name(), err)
		}
	}(dbFile)

	n, err := dbFile.Write(bts)
	if err == nil {
		log.Printf("successfully wrote %d byte(s) to %s file", n, dbFile.Name())
	}
	return n, err
}

// close closes an open db file
func (db *DB) close(f *os.File) {
	if err := f.Close(); err != nil {
		log.Printf("could not close file '%s': %v", f.Name(), err)
	}
}
