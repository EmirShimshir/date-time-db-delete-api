.PHONY: run down restart test

run:
	docker-compose up -d

down:
	docker-compose down -v

restart: down run

test: restart
	@echo "Ожидание 20 секунд..."
	@sleep 20
	sh ./test/curl.sh