# Nom du binaire à produire
BINARY_NAME=forum.exe

# Liste explicite des fichiers source
SOURCES := main.go

# Commande de build
build:
	@echo "Construction du projet..."
	go build -o $(BINARY_NAME) $(SOURCES)

# Commande pour nettoyer le projet (supprimer le binaire)
clean:
	@echo "Nettoyage..."
ifeq ($(OS),Windows_NT)
	cmd /C del $(BINARY_NAME)
else
	rm $(BINARY_NAME)
endif

# Commande pour exécuter le programme
run: build
	@echo "Exécution du programme..."
	./$(BINARY_NAME)

# Option 'phony' pour indiquer que 'clean', 'run', et 'build' ne sont pas des fichiers
.PHONY: build clean run
