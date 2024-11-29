cover:
	go test ./... failfast -coverprofile=coverage.out

cover-v:
	go test ./... failfast -coverprofile=coverage.out -v

cover-html:
	go test ./... -failfast -coverprofile=coverage.out
	go tool cover -html=coverage.out

fmt:
	./scripts/format.sh

review:
	go test ./... -failfast -coverprofile=coverage.out
	./scripts/format.sh
	./scripts/check.sh

