# VibeRiichiMahjong

Photograph your Riichi Mahjong hand → get the tiles recognized → get your score.

A personal research project exploring AI-assisted development workflows. Not commercial.

## Components

- **backend/** — own Riichi scoring engine (yaku, han, fu, points) + HTTP API. Language TBD.
- **frontend/** — React + TypeScript + Vite web app.
- **vision/** — Python service, fine-tuned YOLO tile detector trained on photos of my own tile set.
- **fixtures/hands/** — shared, language-agnostic scoring test cases.

Plan: backend first, then frontend, then vision, then Docker-based deployment.

## Docs

- [Architecture](docs/architecture.md)
- [Rules & scope](docs/domain/rules.md)
- [Glossary & tile notation](docs/domain/glossary.md)
- [Decision records](docs/adr/)
- [Workflow](docs/workflow.md)
- [Agent instructions](AGENTS.md)

Planning and tickets live in GitHub Issues, Milestones and the [project board](https://github.com/users/lukamrko/projects/1).

## License

MIT
