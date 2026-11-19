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
go run ./cmd/habit-idle balance  # tabela de ritmo da torre (vai para docs/regras.md)
HABIT_IDLE_DB=./data/demo.db go run ./cmd/habit-idle demo   # banco novo com 20 dias de dados
```

## Deploy

`sh scripts/deploy.sh`: testa, compila para arm64, envia e roda
`deploy/install.sh` na Pi, que é idempotente: serviço `habit-idle` (porta
3002, dados em `~/data/habit-idle`), backup aos domingos 3h20 no
`homelab-backups` (`habit-idle backup <arquivo>`), Caddy em
`http://habit-idle.local` e Tailscale Serve na porta 10000. Os arquivos de
`deploy/` e os `.sh` ficam com LF (`.gitattributes`): a Pi não aceita CRLF.

## Modelo

- `check_log` nunca é apagado: cada marcar/desmarcar é uma linha com a hora,
  porque o fôlego depende de quando o toque aconteceu. O estado atual de um
  hábito num dia é a última linha dele.
- `event` guarda as escolhas do jogador. Nível, moedas, andar e itens nunca são
  gravados: o motor recalcula tudo a partir de hábitos, checks, eventos e
  ajustes.

## Motor (`internal/game`)

- `Folego` (fôlego, dívida, virada, cansaço, dias completos) recebe hábitos e
  o registro de checks e devolve os intervalos de exploração. `Play` recebe
  esses intervalos e os eventos e recalcula torre, heróis, moedas e itens.
  Os dois são puros; o servidor monta tudo a cada `GET /api/state`.
- Todo sorteio passa por `seed(...)`: o mesmo resultado em qualquer recálculo.
  Não use `rand` sem semente.
- Números só em `balance.go`. Mudou: rode `balance` e atualize
  `docs/regras.md`.
- Evento inválido (sem moedas, herói que não existe, desafio repetido no dia)
  é ignorado pelo motor, não recusado pela API.

## API

`/api/today`, `/api/habits` (`GET`, `POST`), `/api/habits/{id}` (`GET` com
estatísticas de 12 semanas, `PUT`, `DELETE`), `PUT /api/checks`
(`{habitId, day, done}`; dia futuro dá 400) e `/api/settings` (`dayStart`
"HH:MM", padrão 06:00; `tz` padrão America/Sao_Paulo). Mudar frequência ou
arquivar passa por `core.Revise`: cria versão a partir do dia.

Jogo: `GET /api/state` (hoje, fôlego, jogo, catálogo, `lastSeen`),
`POST /api/events` (`{kind, data}`: start, recruit, squad, challenge, buy,
equip, unequip, sell, refine), `PUT /api/seen`, `DELETE /api/game`
(recomeçar a torre; hábitos ficam) e `/api/backup` (`GET` exporta tudo,
`PUT` substitui tudo numa transação).

## Telas (`web/`)

- Hash routing em `js/app.js`; cada tela é um módulo em `js/views/` com
  `view(ctx)` que devolve `{ html, actions, mount }`.
- `html` (`js/ui.js`) escapa os valores; template aninhado entra cru.
  Eventos: `data-click`, `data-change` e `data-submit` com o nome da ação; o
  `#view` tem um handler só por tipo, porque os filhos são recriados.
- Ao abrir o app (ou voltar a ele), `app.js` mostra o Retorno se a última
  visita foi há mais de 2h e algo aconteceu; senão marca a visita.
- Sprites: 0x72 DungeonTileset II (CC0) em `web/assets/sprites`, fitas de 8
  quadros 16×28 (0–3 correndo, 4–7 parado). Os quadros soltos do pacote
  (`frames/*_idle_anim_f*` e `*_run_anim_f*`) viram uma fita assim.
- `battleScene` (`js/game.js`): explorando, a cena roda o ciclo de luta de
  8 s do protótipo (só visual, com o dano de ATK de cada herói); sem fôlego,
  o esquadrão fica parado. Animações em CSS, desligadas com
  `prefers-reduced-motion`.
- Cores e componentes do Apollo em `css/app.css`; fontes em `web/assets/fonts`
  (nada carregado de fora). Input com 16px para o Safari não dar zoom.
