.PHONY: help build reset remove cleanup

d=docker
dc=docker-compose
run=$(dc) run app

help: ## Show this help
	@echo "Targets:"
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/\(.*\):.*##[ \t]*/    \1 ## /' | sort | column -t -s '##'
	@echo

build: ## Build the gosaic app
	$(dc) up

reset: ## Removes the database
	rm -f ./db/gosaic.sqlite3

remove: ## Removes the application
	$(dc) down
	$(d) image rm gosaic_app

cleanup: reset remove ## Reset and remove
