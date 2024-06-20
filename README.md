# Gaming Universe Forum

Gaming Universe Forum est une application web de discussion où les utilisateurs peuvent créer des comptes, publier des messages, commenter des publications, liker des messages et interagir avec d'autres utilisateurs.

## Table des matières

- [Fonctionnalités](#fonctionnalités)
- [Technologies Utilisées](#technologies-utilisées)
- [Installation](#installation)
- [Configuration](#configuration)
- [Utilisation du Makefile](#utilisation-du-makefile)
- [Démarrage](#démarrage)
- [Structure du Projet](#structure-du-projet)
- [Dépendances](#dépendances)
- [Contribuer](#contribuer)
- [Licence](#licence)

## Fonctionnalités

### Authentification

- **Inscription** : Les utilisateurs peuvent créer un compte en fournissant un nom d'utilisateur, un email et un mot de passe.
- **Connexion** : Les utilisateurs peuvent se connecter à leur compte.
- **Déconnexion** : Les utilisateurs peuvent se déconnecter de leur compte.

### Gestion des Posts

- **Créer un post** : Les utilisateurs connectés peuvent créer un post.
- **Modifier un post** : Les utilisateurs peuvent modifier leurs propres posts.
- **Supprimer un post** : Les utilisateurs peuvent supprimer leurs propres posts.
- **Voir un post** : Les utilisateurs peuvent voir les détails d'un post.
- **Liker un post** : Les utilisateurs peuvent liker un post.

### Commentaires

- **Ajouter un commentaire** : Les utilisateurs connectés peuvent commenter un post.
- **Voir les commentaires** : Les utilisateurs peuvent voir tous les commentaires d'un post.
- **Liker un commentaire** : Les utilisateurs peuvent liker un commentaire.

### Profil

- **Voir le profil** : Les utilisateurs peuvent voir leur propre profil et celui des autres.
- **Modifier le profil** : Les utilisateurs peuvent modifier leur propre profil.
- **Suivre un utilisateur** : Les utilisateurs peuvent suivre d'autres utilisateurs.
- **Se désabonner d'un utilisateur** : Les utilisateurs peuvent se désabonner d'autres utilisateurs.
- **Voir les abonnés** : Les utilisateurs peuvent voir la liste de leurs abonnés.
- **Voir les abonnements** : Les utilisateurs peuvent voir la liste des utilisateurs qu'ils suivent.

### Réinitialisation du mot de passe

- **Mot de passe oublié** : Les utilisateurs peuvent réinitialiser leur mot de passe en répondant à des questions de sécurité.

## Technologies Utilisées

- **Backend** : Go (Golang) avec les bibliothèques Gorilla Mux et GORM.
- **Frontend** : HTML, CSS, JavaScript.
- **Base de Données** : SQLite.

## Installation

1. Assurez-vous d'avoir Go installé sur votre machine.
2. Clonez ce repository :

   ```sh
   git clone https://github.com/votre-utilisateur/forum.git
   ```

3. Accédez au répertoire du projet :
   ```sh
   cd forum
   ```
4. Installez les dépendances :

   ```sh
   go mod tidy
   ```

5. Installez TDM-GCC-64. Vous pouvez le télécharger [ici](https://jmeubank.github.io/tdm-gcc/).

### Utilisation de TDM-GCC-64

1. **Téléchargement et installation :**

   - Téléchargez l'installeur depuis [ce lien](https://jmeubank.github.io/tdm-gcc/download/).
   - Choissez l'installeur `tdm64-gcc-10.3.0-2.exe`.
   - Exécutez l'installeur et suivez les instructions pour installer TDM-GCC-64 sur votre système.

2. **Configuration de l'environnement :**
   - Après l'installation, ajoutez le répertoire `bin` de TDM-GCC-64 à votre variable d'environnement `PATH`. Cela permet à votre système de trouver les exécutables nécessaires.
   - Pour ce faire, suivez ces étapes :
     1. Ouvrez les "Paramètres système avancés" sur Windows.
     2. Cliquez sur "Variables d'environnement".
     3. Dans la section "Variables système", trouvez la variable `Path` et cliquez sur "Modifier".
     4. Ajoutez le chemin du répertoire `bin` de TDM-GCC-64 (par exemple, `C:\TDM-GCC-64\bin`).
     5. Cliquez sur "OK" pour enregistrer les modifications.

## Configuration

Le fichier `main.go` contient la configuration principale de l'application, y compris la configuration de la base de données et la configuration des routes.

## Utilisation du Makefile

Le projet inclut un `Makefile` pour simplifier les tâches courantes telles que la compilation, l'exécution et le nettoyage. **Assurez-vous d'être à la racine du projet** pour exécuter les commandes `make`.

Voici les cibles disponibles dans le `Makefile` :

### `make build`

Cette cible compile le projet et génère un fichier exécutable.

Commande :

```sh
make build
```

Ce que cela fait :

- Exécute la commande `go build -o forum` pour compiler le projet et générer un fichier exécutable nommé `forum`.

### `make run`

Cette cible compile le projet (si ce n'est pas déjà fait) et exécute l'application.

Commande :

```sh
make run
```

Ce que cela fait :

- Exécute la commande `go run main.go` pour démarrer l'application.

### `make clean`

Cette cible supprime les fichiers générés (notamment l'exécutable).

Commande :

```sh
make clean
```

Ce que cela fait :

- Supprime le fichier exécutable généré `forum` et tout autre fichier temporaire.

Exemple d'utilisation :

```sh
make build
make run
make clean
```

## Démarrage

Compilez et lancez l'application en utilisant le Makefile :

```sh
make build
make run
```

Ouvrez votre navigateur et accédez à `http://localhost:8080`.

## Structure du Projet

- **main.go** : Point d'entrée principal de l'application.
- **common** : Contient les fichiers de configuration et les fonctions utilitaires.
- **handlers** : Contient les gestionnaires pour les différentes fonctionnalités (authentification, général, posts, profil).
- **models** : Contient les définitions des modèles de données.
- **static** : Contient les fichiers CSS et JavaScript.
- **templates** : Contient les fichiers HTML pour le rendu côté client.

## Dépendances

Les principales dépendances du projet sont listées dans le fichier `go.mod` :

- [Gorilla Mux](https://github.com/gorilla/mux) : Routeur HTTP pour Go.
- [Gorilla Sessions](https://github.com/gorilla/sessions) : Gestion des sessions pour Go.
- [GORM](https://gorm.io/) : ORM pour Go.
- [SQLite Driver for GORM](https://gorm.io/docs/connecting_to_the_database.html#SQLite) : Driver SQLite pour GORM.

## Contribuer

Les contributions sont les bienvenues ! Veuillez soumettre des pull requests pour toutes les améliorations ou corrections de bugs.

## Licence

Ce projet est sous licence MIT. Voir le fichier `LICENSE` pour plus d'informations.

---

Pour toute question ou assistance, veuillez ouvrir une issue sur le dépôt GitHub.

Merci d'utiliser Gaming Universe Forum !
