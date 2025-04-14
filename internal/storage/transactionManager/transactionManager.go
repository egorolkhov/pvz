package transactionManager

import (
	"avito/internal/config"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"strings"
	"time"
)

type Transactor struct {
	DB *sql.DB
}

func NewTransactionManager(cfg config.Config) *Transactor {
	DatabaseDSN := GetDatabaseDSN(cfg)
	log.Println("STORAGE", DatabaseDSN)
	db, err := sql.Open("pgx", DatabaseDSN)
	if err != nil {
		log.Fatalf(err.Error())
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)

	for i := 0; i < 50; i++ {
		go func() {
			_ = db.Ping()
		}()
	}
	time.Sleep(1 * time.Second)

	return &Transactor{db}
}

func GetDatabaseDSN(config config.Config) string {
	if config.PostgresConn != "" {
		return config.PostgresConn
	} else if config.PostgresJDBCUrl != "" {
		result, _ := jdbcToGoConnectionString(config.PostgresJDBCUrl, config.PostgresUser, config.PostgresPass)
		return result
	} else {
		var sb strings.Builder
		sb.WriteString("postgres://")
		sb.WriteString(config.PostgresUser)
		sb.WriteString(":")
		sb.WriteString(config.PostgresPass)
		sb.WriteString("@")
		sb.WriteString(config.PostgresHost)
		sb.WriteString(":")
		sb.WriteString(config.PostgresPort)
		sb.WriteString("/")
		sb.WriteString(config.PostgresDB)
		sb.WriteString("?sslmode=disable")
		result := sb.String()
		return result
	}
}

func jdbcToGoConnectionString(jdbc string, username, password string) (string, error) {
	if !strings.HasPrefix(jdbc, "jdbc:postgresql://") {
		return "", fmt.Errorf("invalid JDBC string: must start with jdbc:postgresql://")
	}

	jdbc = strings.TrimPrefix(jdbc, "jdbc:postgresql://")

	goConnStr := fmt.Sprintf("postgres://%s:%s@%s", username, password, jdbc)

	if !strings.Contains(goConnStr, "?") {
		goConnStr += "?sslmode=disable"
	} else if !strings.Contains(goConnStr, "sslmode=") {
		goConnStr += "&sslmode=disable"
	}

	return goConnStr, nil
}
