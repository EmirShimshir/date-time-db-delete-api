.PHONY: run down restart test

run:
	docker-compose up -d

down:
	docker-compose down -v

restart: down run

test: restart
	sh ./test/curl.sh