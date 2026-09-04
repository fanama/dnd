# D&D - Async Tabletop RPG Engine (Go & Svelte Edition)

A multiplayer, asynchronous, ultra-lightweight tabletop RPG (TTRPG) platform. The backend is built in **Go** with WebSockets and **SQLite** persistence. The frontend is a reactive **Svelte 4** app with **TypeScript**, powered by **Vite**.

The platform manages real-time virtual tabletop sessions: player movement, physical and magic combat, inventory management, location exploration, and instant event logging.

---

## Architecture Overview

The project follows a clean modular architecture with strict separation between domain, services, and persistence:

### Project Structure
```
/
├── backend/                    # Go game server
│   ├── go.mod                  # Module & dependencies (SQLite, Gorilla WebSocket)
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # Entry point, Dependency Injection
│   └── internal/
│       ├── domain/             # Pure business logic (Models)
│       │   └── models.go       # Character, Stats, Item, Spell, World...
│       ├── repository/         # Data layer (Persistence)
│       │   └── sqlite_repo.go  # SQLite operations (Save/Get Character)
│       └── services/           # Orchestration logic
│           └── game_manager.go # WebSocket management, Combat, Sync
│
└── frontend/                   # Svelte + TypeScript web app
    ├── src/
    │   ├── lib/
    │   │   ├── components/     # Atomic Design (Atoms, Molecules, Organisms)
    │   │   │   ├── atoms/      # Button, HPBar, Input, StatLabel
    │   │   │   ├── molecules/  # CharacterSheet, InventoryItem, LootItem
    │   │   │   └── organisms/  # ChatBox, LocationExplorer, pages/
    │   │   └── stores/
    │   │       └── game.ts     # Global store & WebSocket communication (TypeScript)
    │   └── routes/
    │       └── +page.svelte    # Main route
    ├── app.css                 # Tailwind config, D&D theme, animations
    └── index.html              # SPA entry point
```

### Data Flow
```mermaid
graph TD
    classDef feNode fill:#2563eb,stroke:#1d4ed8,stroke-width:2px,color:#fff;
    classDef beNode fill:#059669,stroke:#047857,stroke-width:2px,color:#fff;
    classDef dbNode fill:#d97706,stroke:#b45309,stroke-width:2px,color:#fff;

    subgraph Frontend["Frontend"]
        UI[Svelte UI]:::feNode --> Store[game.ts Store]:::feNode
        Store --> WS_C[WebSocket Client]:::feNode
    end

    subgraph Backend["Backend"]
        WS_S[WebSocket Server]:::beNode --> GM[GameManager Service]:::beNode
        GM --> Domain[Domain Models]:::beNode
        GM --> Repo[SQLite Repository]:::beNode
        Repo --> DB[(SQLite DB)]:::dbNode
    end

    WS_C <-->|JSON via WS| WS_S

    style Frontend fill:#eff6ff,stroke:#3b82f6,stroke-width:2px,color:#1e3a8a
    style Backend fill:#ecfdf5,stroke:#10b981,stroke-width:2px,color:#064e3b
    linkStyle default stroke:#64748b,stroke-width:2px;
    linkStyle 4 stroke:#f59e0b,stroke-width:3px;
```

---

## Installation & Setup

### 1. Backend (Go)
Requires Go installed.
```bash
cd backend
go run cmd/server/main.go
```
Server starts on `:8000`. The `game.db` SQLite database is created automatically.

### 2. Frontend (Svelte + TypeScript)
Requires Node.js or Bun.
```bash
cd frontend
npm install    # or bun install
npm run dev    # or bun run dev
```
Access at **`http://localhost:5173`**.

### Build
```bash
cd frontend
npm run build       # Production build
npm run check       # Type-check with svelte-check
```

---

## WebSocket Protocol

All actions are transmitted as JSON objects.

### Client -> Server Actions
* **Init (`type: "init"`)**: Sends character name and class. The server uses the URL pseudo to restore existing stats if the player already exists.
* **Move (`type: "move"`)**: `{"type": "move", "destination": "Donjon"}`
* **Attack (`type: "attack"`)**: `{"type": "attack", "cible": "Pseudo"}`
* **Spell (`type: "cast_spell"`)**: `{"type": "cast_spell", "sort": "Nom", "cible": "Pseudo"}`
* **Consumable (`type: "use_consumable"`)**: `{"type": "use_consumable", "item_index": 0}`
* **Loot (`type: "loot"`)**: `{"type": "loot", "loot_index": 0}`

### Server -> Client Events
1. **Chat (`type: "chat"`)**: Broadcasts game events to all players.
2. **Sync (`type: "sync"`)**: Full state of all players and locations.

---

## Game Engine Rules

* **Max HP**: Constitution x 10.0
* **Physical Damage**: Force x 2.0
* **Armor**: Vitesse x 1.5
* **Combat**: Damage = Attacker Damage - Target Armor (minimum 1.0)
* **Persistence**: Characters are saved to SQLite after every action

---

## Frontend Features

* **HP Bar**: Videogame-style health bar with instant green fill, red damage trail that catches up slowly, and a red flash on hit. Pulses red when health drops below 25%.
* **Atomic Design**: Component architecture split into Atoms (Button, HPBar, Input, StatLabel), Molecules (CharacterSheet, InventoryItem, LootItem), and Organisms (ChatBox, LocationExplorer, LoginPage, GamePage).
* **Responsive Layout**: Sticky sidebar on desktop, tab-based navigation on mobile.
* **Chat Log**: Parchment-styled event journal with numbered entries and auto-scroll.
* **Class Selection**: Visual card-based class picker (Guerrier / Magicien) instead of a raw dropdown.
* **Form Validation**: Real-time validation with error states on the login form.

---

## Domain Model

```mermaid
classDiagram
    class Controller {
        <<Entity>>
        +string pseudo
        +List~Personnage~ personnages
    }
    class Joueur {
        <<Entity>>
        keyboardInput() String
    }
    class Writer {
        <<Entity>>
        write()
    }
    class IA {
        <<Entity>>
        generateText() String
        selectAction()
    }
    class World {
        <<Entity>>
        +uuid id
        Liste~Joueur~ joueurs
    }
    class Caractéristiques {
        <<Entity>>
        +String nom
        +String background
        -Number force
        -Number constitution
        -Number vitesse
        -Number charisme
        -Number Instinct
        -Number Savoir
        +CalculateLifePoints() Number
        +CalculateArmor() Number
        +CalculateDamage() Number
        +getMod(param:String) Number
    }
    class Personnage {
        <<Entity>>
        +String alignement
        +Number currentPV
        +List~Object~ inventaire
        +List~Quete~ quêtes
        +speak(text:String)
        +attack(ennemy: Personnage)
        +takeDamage(damage: Number)
        +equip(object: Object)
        +buy(object:Object)
        +sell(object:Object)
    }
    class Ennemy {
        <<Entity>>
        +List~Object~ inventaire
        +Number Experience
    }
    class Lieu {
        <<Entity>>
        +Position position
        +Caractéristiques stats
        +List~Personnage~ personnages
        +List~Object~ objects
    }
    class Quete {
        <<Entity>>
        +uuid id
        +String titre
        +String description
        +String statut
        +Number xpRecompense
        +verifierObjectifs() Boolean
    }
    class Objet {
        <<Entity>>
        +Number prix
        +Number encombrement
        +Caractéristiques stats
    }
    class Equipement {
        <<Entity>>
        +String emplacement
        +Number portée
    }
    class Consommable {
        <<Entity>>
        +Number Durée
    }
    class Vetement {
        <<Entity>>
        +uuid id
    }
    class Arme {
        <<Entity>>
        +uuid id
    }
    class Classe {
        <<Entity>>
        +Number niveau
        +Caractéristiques stats
        +Liste~Sort~ sorts
        +Liste~Competences~ compétences
        +levelUp()
    }
    class Race {
        <<Entity>>
        +String nomRace
        +Caractéristiques stats
        +Liste~Sort~ sorts
        +Liste~Competences~ compétences
    }
    class Competence {
        <<Entity>>
        +String nom
        +String stat
        +String description
        +Number maitrise
    }
    class Sort {
        <<Entity>>
        +String nom
        +Number niveauSort
        +String ecoleMagie
        +String portee
        +String duree
        +lancerSort(cible: Personnage)
    }
    Controller *--"1" Writer
    Controller <|-- IA
    Controller <|-- Joueur
    Equipement --|> Objet
    Consommable --|> Objet
    Vetement --|> Equipement
    Arme --|> Equipement
    Personnage "1"o--"1" Classe
    Personnage "1"*--"1" Race
    Personnage <|-- Ennemy
    World "1"o-->"*" Lieu
    Lieu "1"o-->"*" Quete
```
