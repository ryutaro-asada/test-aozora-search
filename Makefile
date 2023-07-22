all : aozora-collector #aozora-search

aozora-collector : # ./cmd/aozora-collector/main.go
	go build -o aozora-collector ./cmd/aozora-collector
	./aozora-collector

#aozora-search : ./cmd/aozora-search/main.go
#	go build -o aozora-search ./cmd/aozora-search

.PHONY: clean

clean:
	rm aozora-collector # aozora-search 
