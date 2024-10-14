fmt:
	go fmt

run:
	go run main.go

build: fmt
	go build -o tally main.go

test: fmt
	go run main.go < test.txt

git comment: fmt
	git commit -am "{{comment}}"
	git push