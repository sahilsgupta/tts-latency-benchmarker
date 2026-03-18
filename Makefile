.PHONY: test-e2e

test-e2e:
	export $$(cat .env | xargs) && go test ./tests/... -v -run TestE2E
