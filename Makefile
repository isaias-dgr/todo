default: start

SHELL := /bin/bash
project:=mstodo
project-path:=github.com/isaias-dgr/todo
service:=ms-todo
ENV?=dev
COMMIT_HASH = $(shell git rev-parse --verify HEAD)

.PHONY: start
start:
	source envs_template
	docker-compose -p ${project} up -d

.PHONY: stop
stop:
	docker-compose -p ${project} down

.PHONY: restart
restart: stop start

.PHONY: logs
logs:
	docker-compose -p ${project} logs -f ${service}

.PHONY: logs-db
logs-db:
	docker-compose -p ${project} logs -f ${service}-db

.PHONY: ps
ps:
	docker-compose -p ${project} ps

.PHONY: build
build:
	docker-compose -p ${project} build --no-cache

.PHONY: clean
clean: stop build start

.PHONY: add
add: install-package-in-container build

.PHONY: install-package-in-container
install-package-in-container:
	docker-compose -p ${project} exec ${service} go get -u ${package}

.PHONY: add-dev
add-dev: install-dev-package-in-container build

.PHONY: install-dev-package-in-container
install-dev-package-in-container: start
	docker-compose -p ${project} exec ${service} go get ${package}

.PHONY: migration-create
migration-create: start
	docker-compose -p ${project} exec ${service} goose -dir ./migrate mysql "${MYSQL_CONN}" create $(name) sql
	sudo chown -R $$USER ./migrate

.PHONY: migrate
migrate: start
	docker-compose -p ${project} exec ${service} goose -dir ./migrate mysql "${MYSQL_CONN}" up

.PHONY: rollback
rollback: start
	docker-compose -p ${project} exec ${service} goose -dir ./migrate mysql "${MYSQL_CONN}" down

.PHONY: reset-db
reset-db: start
	docker-compose -p ${project} exec ${service} goose -dir ./migrate mysql "${MYSQL_CONN}" reset

.PHONY: shell
shell:
	docker-compose -p ${project} exec ${service} sh

.PHONY: mysql
mysql:
	docker-compose -p ${project} exec ${service}-db mysql -u root -p

.PHONY: test
test: test-exec

.PHONY: test-exec
test-exec:
	docker-compose -p ${project} exec -T ${service} go test ${project-path}/...

.PHONY: lint
lint:
	docker-compose -p ${project} exec -T ${service} gofmt -d -l -s -e .
	docker-compose -p ${project} exec -T ${service} go vet ${project-path}/...
	docker-compose -p ${project} exec -T ${service} staticcheck ./...

.PHONY: lint-fix
lint-fix:
	docker-compose -p ${project} exec ${service} go fmt ${project-path}/...

.PHONY: test-cov
test-cov:
	docker-compose -p ${project} exec -T ${service} go test -coverprofile=./tmp/profile.out ${project-path}/...
	docker-compose -p ${project} exec -T ${service} go tool cover -func=./tmp/profile.out

.PHONY: commit-hash
commit-hash:
	@echo $(COMMIT_HASH)

.PHONY: build-release
build-release:
	docker build --target release -t local/${service}:${COMMIT_HASH} .

.PHONY: run-release
run-release:
	docker run -d --name ${service}_${COMMIT_HASH} -p :8080 local/${service}:${COMMIT_HASH}
	docker logs -f ${service}_${COMMIT_HASH}
