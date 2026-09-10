package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	err error
)

func InitGorm() *gorm.DB {
	// dsn := os.Getenv("DB_DSN")
	// if dsn == "" {
	dsn := "host=localhost user=root password=root dbname=root port=5432 sslmode=disable TimeZone=America/Sao_Paulo"
	// }

	// stringConexao := "host=localhost user=root password=root dbname=root port=5432 sslmode=disable TimeZone=America/Sao_Paulo"
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Não foi possível conectar com o banco de dados")
	}

	// DB.AutoMigrate(&models.Aluno{})

	return DB
}
