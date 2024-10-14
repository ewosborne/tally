fmt:
	go fmt

run:
	go run main.go

build:
	go build -o tally main.go

test:
	go run main.go < test.txt

git comment:
	git commit -am "{{comment}}"
	git push