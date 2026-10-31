# Product overview

## What we are building
**Realmforge** is an engine for persistent, browser- and mobile-first, multiplayer strategy games of the economic-military kind. Players grow an empire over days and weeks while the world keeps running when they are offline. Each game built on it is a separate product defined by a **Game Pack**: names, setting, races, resources, buildings, units, technologies, quests, balance, translations and visual theme. Fixes and improvements to the engine reach every game.

## Game 1: Ashen Crowns (original IP)
*When the sky shattered, the old crowns fell. From the ash rise new houses, hungry for starglass.*

- **Peoples:** four races in two factions. The **Dawnpact** are the Thornkin and the Ironvow Houses; the **Duskbound** are the Ashborn and the Gravewardens. Each has its own bonuses and a unique unit, and attacking your own race or faction carries a penalty.
- **Economy:** seven resources: timber, grain, stone, iron, gold, lore (research) and starglass (premium, earned only in-game).
  - Population brings both workers and taxpayers: employed people produce, idle people pay gold.
  - Army wages are paid in gold, and unpaid troops desert.
  - Stock above storage capacity slowly rots.
- **Growth:** 22 buildings and 18 sciences, including mutually exclusive "Paths" that act as specialisations. Up to 9 cities, founded with settlers.
- **Warfare:** 16 trainable units with a counter loop: infantry beat ranged, ranged beat beasts, beasts beat cavalry, cavalry beat infantry.
  - Battles run in up to 20 rounds, with walls, retreat thresholds, plunder and detailed reports.
  - Missions: scouting with detection risk, transport, reinforcement of allies, occupying and capturing territories.
- **PvE:** wolf dens, brigand camps, ancient barrows and starfall craters, each in 5 levels. They respawn and drop hero items.
- **Heroes:** 4 classes, skills, equipment slots and experience from battle; heroes can be wounded.
- **Social:** alliances with roles and diplomacy (ally, non-aggression pact, war), a player market with escrow, a merchant exchange, mail with report/block, rankings, and a quest-driven tutorial.
- **Fair play:** newbie protection (72 h, ends early if you attack), at most 2 attacks per target per day, no attacks between allies, and server-side validation of everything.

## Game 2 (demo): Void Frontier
A sci-fi colony setting on the same core with different resources and races. It also switches several rules: biomass pays army upkeep, population grows from idle people, and storage has a hard cap. It shows how a new game is produced; it is not a finished game.

## Audience & platforms
- **Audience:** players of long-session browser strategy who check in several times a day, and mobile players who need short sessions.
- **Platforms:** the web (served by the game server), Yandex Games (HTML5 build) and Google Play (Capacitor wrapper). Both stores need extra work before submission: see `docs/spec/PLATFORM_REQUIREMENTS.md`.

## Principles
1. Persistent, fair, server-authoritative world: no client-side cheating surface.
2. Meaningful decisions: housing vs. jobs, wages vs. army size, specialisation choice, raid vs. defend, collect vs. guard territories.
3. No dead ends: tutorial quests, protection, merchant exchange and PvE rewards keep new players moving.
4. No pay-to-win in the MVP: no real-money sales, and the premium currency is earned in play.

## Status
The core loop, empire growth, map & PvE, army & PvP, heroes, social economy, Game Pack engine and web client are implemented and tested. Platform submission, load testing and several deeper systems from the genre are still open; see `docs/IMPLEMENTATION_STATUS.md` and `docs/ROADMAP.md`.

Full product requirements: [`docs/spec/PRD.md`](docs/spec/PRD.md). Game design: [`docs/spec/GAME_DESIGN_DOCUMENT.md`](docs/spec/GAME_DESIGN_DOCUMENT.md).
