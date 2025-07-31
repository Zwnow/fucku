test:
	go test -race -parallel 32 -v -count=1 ./...

db:
	docker exec -it postgres psql -U postgres
