<<<<<<< HEAD
# Nom du binaire à produire
BINARY_NAME=forum.exe

# Liste explicite des fichiers source
SOURCES := main.go
=======
# Nom de l'exécutable
BINARY := forum.exe

# Répertoire des sources
SRC_DIR := .
>>>>>>> 29a298a9d4487b330fd2b06109be617dd112274b

# Commande de build
build:
	@echo "Construction du projet..."
<<<<<<< HEAD
	go build -o $(BINARY_NAME) $(SOURCES)

# Commande pour nettoyer le projet (supprimer le binaire)
clean:
	@echo "Nettoyage..."
ifeq ($(OS),Windows_NT)
	del $(BINARY_NAME)
else
	rm $(BINARY_NAME)
endif
=======
	@go build -o $(BINARY) $(SRC_DIR)/main.go

# Commande pour nettoyer le projet (supprimer les fichiers binaires)
clean:
	@echo "Nettoyage..."
	@if exist $(BINARY) del $(BINARY)
>>>>>>> 29a298a9d4487b330fd2b06109be617dd112274b

# Commande pour exécuter le programme
run: build
	@echo "Exécution du programme..."
<<<<<<< HEAD
	./$(BINARY_NAME)
=======
	@.\$(BINARY)
>>>>>>> 29a298a9d4487b330fd2b06109be617dd112274b

# Option 'phony' pour indiquer que 'clean', 'run', et 'build' ne sont pas des fichiers
.PHONY: build clean run
