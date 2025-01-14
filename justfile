fmt:
	go fmt

run:
	go run main.go

build: fmt
	go build -o tally main.go

test: fmt build
	go test
	./tally < test.txt
	./tally test.txt
	./tally test.txt test.txt
	./tally test.txt test.txt test.txt
	./tally test.txt test.txt test.txt --json
	./tally test.txt test.txt test.txt --text


git comment: fmt
	git commit -am "{{comment}}"
	git push

install: test
	go install
