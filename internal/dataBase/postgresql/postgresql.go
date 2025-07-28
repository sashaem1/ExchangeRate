package postgresql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sashaem1/ExchangeRate/internal/api"
	database "github.com/sashaem1/ExchangeRate/internal/dataBase"
)

type PostgreSqlDB struct {
	db *sql.DB
}

func NewPostgreSqlDB() *PostgreSqlDB {
	return &PostgreSqlDB{}
}

func (pdb *PostgreSqlDB) InitDB(api api.ExchangeRates) {
	// Получаем переменные окружения
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Формируем строку подключения
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	pdb.db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}

	for i := 0; i < 10; i++ {
		err = pdb.db.Ping()
		if err == nil {
			log.Println("Успешное подключение к базе данных")
			break
		}
		log.Printf("Попытка %d: не удалось подключиться к базе данных: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	pdb.fillDB(api)
}

func (pdb *PostgreSqlDB) fillDB(api api.ExchangeRates) {
	pdb.fillDefaultData(api)
	pdb.fillApiKey()
}

func (pdb *PostgreSqlDB) fillDefaultData(api api.ExchangeRates) {
	defaultDate := []string{
		"2025-07-21",
		"2025-07-22",
		"2025-07-23",
		"2025-07-24",
		"2025-07-25",
	}

	for _, date := range defaultDate {
		_, missingHistoricalRates, err := pdb.GetRatesByDate(api.GetDefaultBase(), date)

		if err != nil {
			log.Println(err.Error())
			continue
		}

		if len(missingHistoricalRates) != 0 {
			ratesApi, err := api.GetLatestExchangeRatesByDate(date)
			if err != nil {
				log.Println(err.Error())
				return
			}

			for _, rate := range missingHistoricalRates {
				for _, rateApi := range ratesApi {
					if rateApi.Base == rate.Base {
						pdb.InsertRate(rateApi.Base, rate.Currency, rateApi.Rates[rate.Currency], date)
					}
				}
			}

		}
	}
}

func (pdb *PostgreSqlDB) fillApiKey() {
	defaultApiFey := os.Getenv("DEFAULT_API_KEY")
	pdb.InsertApiKey(defaultApiFey)
}

func (pdb *PostgreSqlDB) CloseConnect() {
	if pdb.db != nil {
		pdb.db.Close()
		log.Println("Подключение к Postgresql закрыто")
	}
}

func (pdb *PostgreSqlDB) InsertRate(base, currency string, rate float64, updatedAt string) error {
	parsedDate, err := time.Parse("2006-01-02", updatedAt)
	if err != nil {
		return fmt.Errorf("Ошибка парсинга даты: %w", err)
	}

	_, err = pdb.db.Exec(
		`INSERT INTO exchange_rates (base, currency, rate, updated_at) VALUES ($1, $2, $3, $4)`,
		base, currency, rate, parsedDate,
	)

	if err != nil {
		return fmt.Errorf("Ошибка внесения данных валют: %w", err)
	}
	return nil
}

func (pdb *PostgreSqlDB) InsertApiKey(key string) error {

	exist, err := pdb.VerifyApiKey(key)
	if err != nil {
		return fmt.Errorf("Ошибка внесения данных ключа: %w", err)
	}

	if exist {
		return nil
	}

	_, err = pdb.db.Exec(
		`INSERT INTO api_keys (key) VALUES ($1)`,
		key,
	)

	if err != nil {
		return fmt.Errorf("Ошибка внесения данных ключа: %w", err)
	}
	return nil
}

func (pdb *PostgreSqlDB) InsertLog(request_type string) error {
	_, err := pdb.db.Exec(
		`INSERT INTO exchange_rates_log (request) VALUES ($1)`,
		request_type,
	)

	if err != nil {
		return fmt.Errorf("Ошибка внесения данных лога: %w", err)
	}
	return nil
}

func (pdb *PostgreSqlDB) GetRatesByPair(base, currency string) (database.Rate, error) {
	var rate database.Rate

	err := pdb.db.QueryRow(
		`SELECT base, currency, rate
		FROM exchange_rates 
		WHERE base = $1 AND currency = $2 AND updated_at = DATE($3)`, base, currency, time.Now(),
	).Scan(&rate.Base, &rate.Currency, &rate.Rate)
	if err == sql.ErrNoRows {
		return rate, nil
	}
	if err != nil {
		return rate, fmt.Errorf("Не удалось получить значение для пары (%s:%s): %w", base, currency, err)
	}

	return rate, nil
}

func (pdb *PostgreSqlDB) GetRatesByDate(bases map[string][]string, DateStr string) (rates []api.RateResponse, missingRates []database.Rate, err error) {
	date, err := time.Parse("2006-01-02", DateStr)
	if err != nil {
		return rates, missingRates, fmt.Errorf("неверный формат даты: %w", err)
	}

	rates = []api.RateResponse{}

	for base, currencies := range bases {
		currentRate := api.RateResponse{
			Base:  base,
			Rates: make(map[string]float64),
		}
		for _, currency := range currencies {
			var rateDB database.Rate

			err := pdb.db.QueryRow(
				`SELECT base, currency, rate
				FROM exchange_rates 
				WHERE base = $1 AND currency = $2 AND DATE(updated_at) = DATE($3)`, base, currency, date,
			).Scan(&rateDB.Base, &rateDB.Currency, &rateDB.Rate)

			if err == sql.ErrNoRows {
				rateDB.Base = base
				rateDB.Currency = currency
				missingRates = append(missingRates, rateDB)
				continue
			}

			if err != nil {
				return rates, missingRates, fmt.Errorf("Не удалось получить значение для пары (%s:%s): %w", base, currency, err)
			}

			currentRate.Rates[currency] = rateDB.Rate
		}
		rates = append(rates, currentRate)
	}

	return rates, missingRates, nil
}

func (pdb *PostgreSqlDB) VerifyApiKey(ApiKey string) (bool, error) {
	apiKeyDB := ""

	err := pdb.db.QueryRow(
		`SELECT key
		FROM api_keys 
		WHERE key = $1`, ApiKey,
	).Scan(&apiKeyDB)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("Не удалось проверить значения ключа %s", err)
	}

	return true, nil
}
