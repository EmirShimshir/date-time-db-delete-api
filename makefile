run:
	docker-compose up -d

down:
	docker-compose down -v

restart: down run