mac:
	GOOS=darwin GOARCH=amd64 go build -o ./strippdf main.go

windows:
	GOOS=windows GOARCH=amd64 go build -o ./strippdf.exe main.go

clean:
	rm *.pdf *.txt *.exe strippdf