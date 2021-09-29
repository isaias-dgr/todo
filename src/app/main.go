package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/isaias-dgr/todo/src/domain"
	"github.com/isaias-dgr/todo/src/task/deliver/http"
	_TaskRepo "github.com/isaias-dgr/todo/src/task/repository/mysql"
	useCase "github.com/isaias-dgr/todo/src/task/usecase"
	"go.uber.org/zap"
)

func SetUpLog() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	return logger.Sugar()
}

func SetUpRepository(logger *zap.SugaredLogger) (*sql.DB, domain.TaskRepository) {
	logger.Info("💾 Set up Database.")
	connection := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
	)
	logger.Debug(connection)
	dbConn, err := sql.Open(`mysql`, connection)
	if err != nil {
		logger.Info(err)
	}
	err = dbConn.Ping()
	if err != nil {
		logger.Info(err)
	}
	return dbConn, _TaskRepo.NewtaskRepository(dbConn, logger)
}

func main() {
	log := SetUpLog()
	msg := fmt.Sprintf(
		"🤓 SetUp %s_%s:%s..",
		os.Getenv("PROJ_NAME"),
		os.Getenv("PROJ_ENV"),
		"8080")
	log.Info(msg)
	log.Info("🚀 API V1.")
	dbConn, task_repo := SetUpRepository(log)
	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	useCase := useCase.NewTaskUseCase(task_repo)
	http.NewTaskHandler(useCase, log)
}
