package storage

import (
	"database/sql"
	"fmt"

	"github.com/alvarofc/mode/types"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(host, port, user, password, dbname string) (*Postgres, error) {
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &Postgres{db: db}, nil
}

func (p *Postgres) GetUserById(id int) (types.User, error) {
	row := p.db.QueryRow("SELECT * FROM users WHERE id = $1", id)

	var user types.User
	err := row.Scan(&user.ID, &user.Name)
	return user, err
}

func (p *Postgres) CreateUser(email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = p.db.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", email, string(hashedPassword))
	return err
}

func (p *Postgres) GetUserByEmail(email string) (types.User, error) {
	var user types.User
	err := p.db.QueryRow("SELECT id, email, password FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email, &user.Password)
	return user, err
}
