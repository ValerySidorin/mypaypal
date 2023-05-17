package pg

import (
	"context"
	"fmt"

	"github.com/ValerySidorin/mypaypal/internal/dto"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

type Config struct {
	Conn string `mapstructure:"conn"`
}

type Storage struct {
	conn *pgx.Conn
}

func NewStorage(ctx context.Context, cfg Config) (*Storage, error) {
	conn, err := pgx.Connect(ctx, cfg.Conn)
	if err != nil {
		return nil, errors.Wrap(err, "init pg storage:")
	}

	init := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		balance INTEGER NOT NULL DEFAULT 0
	);
	
	CREATE TABLE IF NOT EXISTS balance_records (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users (id),
		amount INTEGER NOT NULL,
		external_id TEXT NOT NULL
	);
	
	INSERT INTO users (id, balance)
	SELECT 1, 100
	WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = 1);

	INSERT INTO users (id, balance)
	SELECT 2, 200
	WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = 2);`

	_, err = conn.Exec(ctx, init)
	if err != nil {
		return nil, errors.Wrap(err, "init db:")
	}

	return &Storage{
		conn: conn,
	}, nil
}

func (s *Storage) ApplyTransaction(ctx context.Context, r dto.BalanceRequest) error {
	// Начало транзакции
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Проверка дубликатов запроса
	var cnt int
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM balance_records WHERE external_id = $1 AND user_id = $2", r.ID, r.UserID).Scan(&cnt)
	if err != nil {
		return err
	}
	if cnt > 0 {
		return fmt.Errorf("duplicate request detected")
	}

	// Проверка баланса пользователя
	var balance int
	err = tx.QueryRow(ctx, "SELECT balance FROM users WHERE id = $1", r.UserID).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user with balance not found")
		}
		return err
	}

	if r.Amount < 0 && balance < -r.Amount {
		// Баланс недостаточен
		return fmt.Errorf("insufficient balance")
	}

	// Создание записи в базе данных
	_, err = tx.Exec(ctx, "INSERT INTO balance_records (user_id, amount, external_id) VALUES ($1, $2, $3)", r.UserID, r.Amount, r.ID)
	if err != nil {
		return err
	}

	// Обновление баланса пользователя
	_, err = tx.Exec(ctx, "UPDATE users SET balance = balance + $1 WHERE id = $2", r.Amount, r.UserID)
	if err != nil {
		return err
	}

	// Фиксация транзакции
	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}
