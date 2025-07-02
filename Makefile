run:
	CompileDaemon -build="go build -o ticket_app ./app" -command="./ticket_app" -build-dir=. -run-dir=. \
	-directory=. -directory=app -directory=auth -directory=domain -directory=health -directory=internal \
	-directory=internal/repository -directory=internal/repository/user -directory=internal/rest \
	-directory=internal/rest/middleware -directory=migrations \
	-pattern="\\.go$$" -log-prefix -verbose -exclude=".git" -exclude=".github" -exclude="misc"
test:
	go test ./internal/rest -v

create-database:
	docker compose -f docker/ticket_database/docker-compose.yml up -d 