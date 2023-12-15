package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/pkg/errors"
)

var (
	errSave     = errors.New("failed to save key in database")
	errRetrieve = errors.New("failed to retrieve key from database")
	errDelete   = errors.New("failed to delete key from database")
	errView     = errors.New("View entity failed")
)
var _ auth.KeyRepository = (*repo)(nil)

const (
	errDuplicate = "unique_violation"
	errInvalid   = "invalid_text_representation"
)

type repo struct {
	db Database
}

// New instantiates a PostgreSQL implementation of key repository.
func New(db Database) auth.KeyRepository {
	return &repo{
		db: db,
	}
}

func (kr repo) Save(ctx context.Context, key auth.Key) (string, error) {
	q := `INSERT INTO keys (id, type, issuer_id, subject, issued_at, expires_at)
	      VALUES (:id, :type, :issuer_id, :subject, :issued_at, :expires_at)`

	dbKey := toDBKey(key)
	if _, err := kr.db.NamedExecContext(ctx, q, dbKey); err != nil {

		pqErr, ok := err.(*pq.Error)
		if ok {
			if pqErr.Code.Name() == errDuplicate {
				return "", errors.Wrap(auth.ErrConflict, pqErr)
			}
		}

		return "", errors.Wrap(errSave, err)
	}

	return dbKey.ID, nil
}

func (kr repo) RetrieveByID(ctx context.Context, issuerID, id string) (auth.Key, error) {
	q := `SELECT id, name, type, issuer_id, subject, issued_at, expires_at FROM keys WHERE issuer_id = $1 AND id = $2`
	key := dbKey{}
	if err := kr.db.QueryRowxContext(ctx, q, issuerID, id).StructScan(&key); err != nil {
		pqErr, ok := err.(*pq.Error)
		if err == sql.ErrNoRows || ok && errInvalid == pqErr.Code.Name() {
			return auth.Key{}, errors.Wrap(auth.ErrNotFound, err)
		}

		return auth.Key{}, errors.Wrap(errRetrieve, err)
	}

	return toKey(key), nil
}

func (kr repo) RetrieveAll(ctx context.Context, issuerID string, pm auth.PageMetadata) (auth.KeyPage, error) {
	var query []string
	var emq string
	query = append(query, fmt.Sprintf("issuer_id = '%s'", issuerID))
	if pm.Type != 0 {
		query = append(query, fmt.Sprintf("type = '%d'", pm.Type))
	}
	if pm.Subject != "" {
		query = append(query, fmt.Sprintf("subject = '%s'", pm.Subject))
	}
	if len(query) > 0 {
		emq = fmt.Sprintf(" WHERE %s", strings.Join(query, " AND "))
	}

	q := fmt.Sprintf(`SELECT id, name, type, issuer_id, subject, issued_at, expires_at FROM keys %s ORDER BY issued_at LIMIT :limit OFFSET :offset;`, emq)
	params := map[string]interface{}{
		"limit":  pm.Limit,
		"offset": pm.Offset,
	}

	rows, err := kr.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return auth.KeyPage{}, errors.Wrap(errView, err)
	}
	defer rows.Close()

	var items []auth.Key
	for rows.Next() {
		dbkey := dbKey{}
		if err := rows.StructScan(&dbkey); err != nil {
			return auth.KeyPage{}, errors.Wrap(errView, err)
		}

		key := toKey(dbkey)
		items = append(items, key)
	}

	cq := fmt.Sprintf(`SELECT COUNT(*) FROM keys %s;`, emq)

	total, err := total(ctx, kr.db, cq, params)
	if err != nil {
		return auth.KeyPage{}, errors.Wrap(errView, err)
	}

	page := auth.KeyPage{
		Keys: items,
		PageMetadata: auth.PageMetadata{
			Total:  total,
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}

	return page, nil
}

func (kr repo) Remove(ctx context.Context, issuerID, id string) error {
	q := `DELETE FROM keys WHERE issuer_id = :issuer_id AND id = :id`
	key := dbKey{
		ID:       id,
		IssuerID: issuerID,
	}
	if _, err := kr.db.NamedExecContext(ctx, q, key); err != nil {
		return errors.Wrap(errDelete, err)
	}

	return nil
}

type dbKey struct {
	ID        string       `db:"id"`
	Name      string       `db:"name"`
	Type      uint32       `db:"type"`
	IssuerID  string       `db:"issuer_id"`
	Subject   string       `db:"subject"`
	Revoked   bool         `db:"revoked"`
	IssuedAt  time.Time    `db:"issued_at"`
	ExpiresAt sql.NullTime `db:"expires_at"`
}

func toDBKey(key auth.Key) dbKey {
	ret := dbKey{
		ID:       key.ID,
		Name:     key.Name,
		Type:     key.Type,
		IssuerID: key.IssuerID,
		Subject:  key.Subject,
		IssuedAt: key.IssuedAt,
	}
	if !key.ExpiresAt.IsZero() {
		ret.ExpiresAt = sql.NullTime{Time: key.ExpiresAt, Valid: true}
	}

	return ret
}

func toKey(key dbKey) auth.Key {
	ret := auth.Key{
		ID:       key.ID,
		Name:     key.Name,
		Type:     key.Type,
		IssuerID: key.IssuerID,
		Subject:  key.Subject,
		IssuedAt: key.IssuedAt,
	}
	if key.ExpiresAt.Valid {
		ret.ExpiresAt = key.ExpiresAt.Time
	}

	return ret
}
