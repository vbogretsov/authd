NAME			:=	$$(basename $$(pwd))
ENV				?=	dev
DOCKER			=	docker
COMPOSE			=	docker-compose
HOST			:=	$$(echo $$DOCKER_HOST  | cut -d : -f 2 | cut -d / -f 3)
GO				?=	go
GOBUILD			=	env GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build
GCFLAGS			?=	""
LDFLAGS			?=	""
GOOS 			:=	$$($(GO) env GOOS)
GOARCH			:=	$$($(GO) env GOARCH)
SRC				=	$(wildcard app/*.go)
BIN				=	bin
EXE				=	$(BIN)/$(NAME)
DBMIGRATE_URL	=	github.com/mattes/migrate/cli
DBMIGRATE		=	$(BIN)/dbmigrate
WAITPG_URL		=	github.com/vbogretsov/waitpg
WAITPG			=	$(BIN)/waitpg

include .$(ENV)

DSN				=	postgres://postgres@$(HOST)/$(DBNAME)?sslmode=disable

default: $(EXE)


up: $(WAITPG) $(DBMIGRATE)
	$(COMPOSE) up -d
	$(WAITPG) $(DSN)
	$(DBMIGRATE) -database $(DSN) -source file://db/migrations/ up

down:
	$(COMPOSE) down

$(BIN):
	mkdir -p $(BIN)

$(EXE): $(SRC) $(BIN)
	$(GO) get -d ./...
	$(GOBUILD) -o $(EXE) -gcflags $(GCFLAGS) -ldflags=$(LDFLAGS) ./cmd/$(NAME)

$(DBMIGRATE): $(BIN)
	$(GO) get -d $(DBMIGRATE_URL) github.com/lib/pq
	$(GOBUILD) -o $(DBMIGRATE) -tags 'postgres' -ldflags="-s -w" $(DBMIGRATE_URL)

$(WAITPG): $(BIN)
	$(GO) get -d $(WAITPG_URL)
	$(GOBUILD) -o $(WAITPG) -ldflags="-s -w" $(WAITPG_URL)

clean:
	$(GO) clean
	rm -rf $(BIN)
