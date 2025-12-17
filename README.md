# 🎵 GROUPIE TRACKER

> Plateforme de jeux musicaux multijoueur en temps réel

**Projet Bachelor 1 Cybersécurité** - Go + WebSockets


## 🚀 Installation rapide

```bash
# 1. Cloner le projet
git clone https://github.com/votre-username/groupie-tracker.git
cd groupie-tracker

# 2. Installer les dépendances
go mod download

# 3. Lancer le serveur
go run main.go
```

Ouvrir **http://localhost:8080**

## 🎮 Lancer le projet

### Première utilisation

1. **Créer un compte** (`/register`)
   - Pseudo avec MAJUSCULE au début
   - Mot de passe CNIL (12+ caractères, maj/min/chiffre/symbole)

2. **Se connecter** (`/login`)
   - Avec pseudo OU email

3. **Créer une partie**
   - Choisir Blind Test ou Petit Bac
   - Noter le code de salle (6 caractères)
   - Partager avec des amis

4. **Jouer**
   - Communication temps réel par WebSocket


## ✨ Fonctionnalités

**Authentification**
- Inscription avec validation CNIL
- Pseudo avec majuscule obligatoire
- Connexion par pseudo OU email

**Salles de jeu**
- Création avec code unique
- WebSocket temps réel

**Blind Test**
- Faute de temps

**Petit Bac**
- 5 catégories de base
- Lettres aléatoires
- Points : 0/1/2

**Scoreboard**
- Affichage pseudos + scores
- Médailles 🥇🥈🥉

## 🛠️ Technologies

- **Go** - Backend
- **SQLite** - Base de données
- **WebSocket** - Temps réel (gorilla/websocket)
- **bcrypt** - Sécurité mots de passe
- **HTML/CSS** - Frontend
- **JavaScript** - WebSocket client uniquement

Bon jeu a toi ! (ustre)
