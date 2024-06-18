# Nom de l'exécutable
BINARY := forum.exe

# Répertoire des sources
SRC_DIR := .

# Commande de build
build:
	@echo "Construction du projet..."
	@go build -o $(BINARY) $(SRC_DIR)/main.go

# Commande pour nettoyer le projet (supprimer les fichiers binaires)
clean:
	@echo "Nettoyage..."
	@if exist $(BINARY) del $(BINARY)

# Commande pour exécuter le programme
run: build
	@echo "Exécution du programme..."
	@.\$(BINARY)

# Option 'phony' pour indiquer que 'clean', 'run', et 'build' ne sont pas des fichiers
.PHONY: build clean run
