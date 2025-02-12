package implementations

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/GoCEP/api/cep/structs"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteCepRepository struct {
	db *sql.DB
}

func NewSqliteCepRepo() *SqliteCepRepository {
	db, err := sql.Open("sqlite3", os.Getenv("SQLITE_PATH"))

	if err != nil {
		error := fmt.Errorf("failed to open database: %w", err)
		panic(error)
	}

	if err := db.Ping(); err != nil {
		error := fmt.Errorf("failed to ping database: %w", err)
		panic(error)
	}

	query := `
    CREATE TABLE IF NOT EXISTS ceps (
      CEP TEXT PRIMARY KEY,
      LOGRADOURO TEXT,
      COMPLEMENTO TEXT,
      BAIRRO TEXT,
      LOCALIDADE TEXT,
      UF TEXT,
      IBGE TEXT
    )
  `

	_, err = db.Exec(
		query,
	)

	if err != nil {
		error := fmt.Errorf("failed to create table cep: %w", err)
		panic(error)
	}

	return &SqliteCepRepository{
		db: db,
	}
}

func (sr *SqliteCepRepository) Read(ctx context.Context, cep string) (*structs.Cep, error) {
	query := `
    SELECT CEP, LOGRADOURO, COMPLEMENTO, BAIRRO, LOCALIDADE, UF, IBGE
    FROM ceps
    WHERE CEP = ?
  `

	row := sr.db.QueryRowContext(
		ctx,
		query,
		cep,
	)

	var result structs.Cep
	err := row.Scan(
		&result.ZipCode,
		&result.PublicPlace,
		&result.Complement,
		&result.District,
		&result.Place,
		&result.Uf,
		&result.IbgeCode,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("faied to read cep: %w", err)
	}

	return &result, nil
}

func (sr *SqliteCepRepository) Create(ctx context.Context, cep structs.Cep) error {
	query := `
		INSERT INTO ceps (CEP, LOGRADOURO, COMPLEMENTO, BAIRRO, LOCALIDADE, UF, IBGE) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
  `
	_, err := sr.db.ExecContext(
		ctx,
		query,
		cep.ZipCode,
		cep.PublicPlace,
		cep.Complement,
		cep.District,
		cep.Place,
		cep.Uf,
		cep.IbgeCode,
	)

	if err != nil {
		return fmt.Errorf("failed to create cep: %w", err)
	}

	return nil
}

func (sr *SqliteCepRepository) CreateAndUpdateMany(ctx context.Context, ceps []structs.Cep) error {
	tx, err := sr.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("couldn't begin transaction")
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				log.Printf("rollback failed: %s", rollbackErr)
			}
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `
    INSERT INTO ceps (CEP, LOGRADOURO, COMPLEMENTO, BAIRRO, LOCALIDADE, UF, IBGE) 
    VALUES (?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(CEP) DO UPDATE SET 
        LOGRADOURO = excluded.LOGRADOURO,
        COMPLEMENTO = excluded.COMPLEMENTO,
        BAIRRO = excluded.BAIRRO,
        LOCALIDADE = excluded.LOCALIDADE,
        UF = excluded.UF,
        IBGE = excluded.IBGE
  `)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %s", err)
	}
	defer stmt.Close()

	for _, cep := range ceps {
		_, err := stmt.ExecContext(ctx,
			cep.ZipCode,
			cep.PublicPlace,
			cep.Complement,
			cep.District,
			cep.Place,
			cep.Uf,
			cep.IbgeCode,
		)
		if err != nil {
			return fmt.Errorf("could not execute statement, error: %s", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction, error: %s", err)
	}

	return nil
}

func (sr *SqliteCepRepository) Update(ctx context.Context, cep structs.Cep) error {
	return nil
}

func (sr *SqliteCepRepository) Delete(ctx context.Context, cep string) error {
	return nil
}
