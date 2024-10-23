help: Makefile
	@echo
	@echo "Choose a make command to run in"$(PROJECTNAME)":"
	@echo
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'
	@echo

start:
	docker compose -p wow -f deploy/docker/docker-compose.yaml up --build

test:
	go clean --testcache
	go test ./...
