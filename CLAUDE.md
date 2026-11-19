# CLAUDE.md

## O que é

**habit-idle** (nome provisório): tracker de hábitos que vira um jogo idle de
torre. Cumprir hábitos dá **fôlego**; com fôlego o esquadrão sobe a torre
sozinho e ganha moedas, XP e itens. Um usuário só, sem notificações.

As regras do jogo estão em **`docs/regras.md`**, e os números delas em
`internal/game/balance.go`. Mudou um, muda o outro no mesmo commit.

O visual é o "L1 Apollo" do canvas de design (telas `M-*`):
https://claude.ai/artifact/EMqnb19jYs5aef2X51kq2f

## Stack

Um binário Go, sem CGO, que roda na Raspberry Pi de casa (ver a skill
`homelab`).

- `cmd/habit-idle`: o executável (`HOST`, `PORT` padrão 3002,
  `HABIT_IDLE_DB` padrão `./data/habit-idle.db`).
- `internal/db`: SQLite (`modernc.org/sqlite`) com migrations embutidas em
  `migrations/*.sql`, aplicadas em ordem no `Open`. Migration aplicada não se
  edita: crie a próxima.
- `internal/server`: HTTP (`/api/*`) e os arquivos de `web/`.
- `web/`: telas em JS puro, sem build, embutidas no binário (`web/embed.go`).

## Comandos

```bash
go test ./...
go run ./cmd/habit-idle          # http://127.0.0.1:3002
```

## Modelo

- `check_log` nunca é apagado: cada marcar/desmarcar é uma linha com a hora,
  porque o fôlego depende de quando o toque aconteceu. O estado atual de um
  hábito num dia é a última linha dele.
- `event` guarda as escolhas do jogador. Nível, moedas, andar e itens nunca são
  gravados: o motor recalcula tudo a partir de hábitos, checks, eventos e
  ajustes.

## API

`/api/today`, `/api/habits` (`GET`, `POST`), `/api/habits/{id}` (`GET` com
estatísticas de 12 semanas, `PUT`, `DELETE`), `PUT /api/checks`
(`{habitId, day, done}`; dia futuro dá 400) e `/api/settings` (`dayStart`
"HH:MM", padrão 06:00; `tz` padrão America/Sao_Paulo). Mudar frequência ou
arquivar passa por `core.Revise`: cria versão a partir do dia.

## Telas (`web/`)

- Hash routing em `js/app.js`; cada tela é um módulo em `js/views/` com
  `view(ctx)` que devolve `{ html, actions, mount }`.
- `html` (`js/ui.js`) escapa os valores; template aninhado entra cru.
  Eventos: `data-click`, `data-change` e `data-submit` com o nome da ação; o
  `#view` tem um handler só por tipo, porque os filhos são recriados.
- Cores e componentes do Apollo em `css/app.css`; fontes em `web/assets/fonts`
  (nada carregado de fora). Input com 16px para o Safari não dar zoom.
