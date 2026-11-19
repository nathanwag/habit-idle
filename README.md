# habit-idle

Tracker de hábitos que vira um jogo idle de torre: cada hábito feito dá fôlego,
e com fôlego o esquadrão sobe a torre sozinho.

Roda num binário Go na Raspberry Pi de casa, acessado pelo Tailscale.

```bash
go test ./...
go run ./cmd/habit-idle   # http://127.0.0.1:3002
```

Regras do jogo: [docs/regras.md](docs/regras.md).

Arte pixel: 0x72 DungeonTileset II (CC0).

Licença: [PolyForm Noncommercial 1.0.0](LICENSE.md).
