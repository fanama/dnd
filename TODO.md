# TODO — Améliorations complètes

## 🔴 Priorité haute

### Retour d'action & feedback
- [ ] **Système de toasts/notifications** (succès action, erreur, loot) dans les deux frontends
- [ ] **Combat log temps réel + dégâts flottants** avec animation sur hit (joueur)
- [ ] **Afficher les chiffres HP** sur toutes les HPBar en combat (`showNumbers`)
- [ ] **Animation du vôtre propre personnage** quand il subit/Inflige des dégâts (flash rouge)

### Chat & communication
- [ ] **Chat joueur actif** — envoi de messages entre joueurs (WS `type: chat_msg`)
- [ ] **Chat MDJ ↔ joueurs** — le MDJ peut parler aux joueurs
- [ ] **Indicateur de messages non-lus** dans l'onglet Journal / badge

### Responsive DM
- [ ] **Fix responsive MDJ** — remplacer `display:none` du `.world-panel` < 1200px par des onglets/accordéon accessibles
- [ ] **Réorganisation mobile** — roster/éditeur/monde accessibles en onglets sur mobile
- [ ] **`min-height` / scroll corrects** des panneaux MDJ sur petit écran

## 🟠 Priorité moyenne

### États & statuts (joueur)
- [ ] **État de mort** — overlay/écran « tu es tombé », info résurrection en Taverne
- [ ] **Affichage des buffs permanents** sur la fiche (sort avec `buff` → badge +stat)
- [ ] **Avatars par classe** — remplacer 👤 par l'icône de la classe (header, roster, combat)
- [ ] **Tooltip/help** sur les stats dérivées (CA, mods, dés de dégâts)
- [ ] **Bouton looter un seul clic** avec confirmation visuelle d'ajout à l'inventaire

### MDJ
- [ ] **Modales de confirmation** stylées (remplacer `confirm()` natif)
- [ ] **Filtre/recherche roster** MDJ (par nom, classe, lieu)
- [ ] **Badges MOB/boss dans la liste PNJ** du manager
- [ ] **Aperçu des stats dérivées** lors de l'édition des stats (CA, PV max, dégâts)
- [ ] **Sélection du PNJ cliquable** depuis le manager (actuellement `selectPlayer` sans highlight)
- [ ] **Indicateur de sauvegarde** (état du monde non sauvegardé / sauvegardé)

### Monde & exploration
- [ ] **Carte du monde** — représentation visuelle des locations et des joueurs présents
- [ ] **Notion de liens entre locations** — déplacement uniquement vers les lieux adjacents (ou au moins affichage des connexions)
- [ ] **Background visuel des locations** — image/ambiance par lieu (au lieu du fond noir uniforme)

## 🟡 Priorité basse

### Polissage
- [ ] **Son/haptics** (hit, soin, loot, critique) avec toggle muet et persistance
- [ ] **Log filterable & paginé** (Joueur + MDJ) — filtres par type d'événement
- [ ] **Logout / switcher de personnage** côté joueur (bouton dans le header)
- [ ] **Raccourcis clavier** (navigation onglets, esc pour fermer)
- [ ] **Accessibilité** — `aria-label`, focus states visibles, contrastes, `role` corrects
- [ ] **Squelettes de chargement** pendant la reconnexion (au lieu d'un texte statique)
- [ ] **Animations de transition** entre les onglets (plus fluides)

### Contenu
- [ ] **Pseudo unique / vérification** — avertir si le pseudo existe déjà (pas de silo)
- [x] **Système d'or** — achat/vente d'items (le champ `prix` existe mais n'est pas utilisé)
- [ ] **Inventaire encumberment** — gestion du poids (`encombrement` existe mais ignoré)
- [ ] **XP / Niveaux** — progression au-delà de la création
- [ ] **Jet de sauvegarde / compétences** — UI pour lancer ses propres dés (1d20 + mod)

### DEV / Qualité
- [ ] **CI** — build + tests frontends et backend à chaque PR
- [ ] **Lint + format** intégré (le dépôt n'a pas de config lint frontend visible)
- [ ] **Tests e2e** du parcours joueur (login → création → combat → loot)
- [ ] **Storybook / catalogues** des composants atomiques
- [ ] **Changelog / versioning** visible (badge de version dans l'UI)
- [ ] **Vite preview incluse** pour tester le build de prod localement
