# Regras do habit-idle

Fonte da verdade do jogo. Todo número aqui é provisório e mora em
`internal/game/balance.go`; mudou lá, muda aqui no mesmo commit.

O motor (`game.Play`) recalcula tudo a partir de hábitos, do registro de
checks, dos eventos do jogador e dos ajustes. Nível, moedas, andar e itens
nunca são gravados: saem do recálculo. Todo sorteio usa uma semente fixa
(`seed(tipo, andar, dia, …)`), então recalcular dá sempre o mesmo resultado.

## Calendário

- O dia vira no horário `dayStart` (padrão **06:00**), no fuso `tz` (padrão
  `America/Sao_Paulo`). "Que dia é" sempre passa por `DayOf(instante)`.
- A semana começa na segunda.
- Frequência do hábito: `daily`, `days` (dias da semana fixos, 0 = domingo) ou
  `weekly` (X vezes por semana). Mudar a frequência ou arquivar cria uma versão
  a partir do dia; dias anteriores seguem a versão antiga.
- **Missões do dia**: os hábitos `daily` e `days` que valem naquele dia. Os
  `weekly` não são missão do dia.

## Fôlego

Saldo em minutos. Com saldo positivo e pelo menos um herói, o esquadrão
explora e gasta 1 minuto de saldo por minuto de relógio. A conta começa no
primeiro check ou no início do jogo, o que vier antes.

- **Check**: marcar um hábito no próprio dia dá o fôlego do esforço dele
  (leve 60, médio 120, pesado 180 min) na hora do toque. Marcar um dia passado
  não dá nada (só histórico).
- **Desmarcar** tira exatamente o que aquele check deu. Se já foi gasto, o
  saldo fica negativo: é **dívida**, que não é gasta e é paga pelo próximo
  fôlego que entrar.
- **Teto**: o saldo nunca passa de **16h** (960 min); o excesso de um ganho se
  perde.
- **Dia completo** (só em dia com 2 missões ou mais: com uma só seria fácil
  demais): marcar a última missão do dia dá +60 min na hora
  (desmarcar uma missão do dia tira esses 60). Se o dia virar completo, o
  esquadrão ganha também um **baú de andar**.
- **Virada do dia**: saldo positivo cai pela metade (arredondado para baixo).
  Dívida não muda.
- **Semanais**: na virada de domingo para segunda, cada vez que faltou de um
  hábito `weekly` vira dívida igual ao fôlego do esforço dele. Só vale se o
  hábito foi semanal a semana inteira.

## Cansaço

Calculado na virada, para o dia que fechou, e vale a partir dela. Só conta
depois do início do jogo (o primeiro herói):

| Dia que fechou | Cansaço |
|---|---|
| nenhum check no próprio dia | +10% (máximo 30%) |
| todas as missões feitas | volta a 0% |
| algum check, mas não todas as missões | −10% (mínimo 0%) |
| sem missões naquele dia | não muda |

O cansaço multiplica o dano, a vida e o Poder por `1 − cansaço`.

## Heróis

| Herói | Classe | Especialidade |
|---|---|---|
| Oren | Mago | ATK |
| Kael | Cavaleiro | DEF |
| Brom | Anão | HP |
| Rasko | Homem-lagarto | VEL |
| Lyra | Elfa | SORTE |

- O jogo começa pelo evento `start` com o herói escolhido.
- Vencer os chefes dos andares 10, 20, 30 e 40 libera **um recrutamento**: o
  jogador escolhe um herói ainda fora do grupo (evento `recruit`). Ele entra
  no nível do herói mais baixo do grupo e vai para o esquadrão se houver vaga.
- **Esquadrão**: até 3 heróis, em ordem (o primeiro leva os golpes do chefe).
  O evento `squad` troca; fica valendo para as próximas lutas.
- Só quem está no esquadrão ganha XP.

### Pontos de status

Cada herói tem pontos em ATK, DEF, HP, VEL e SORTE:

- nível 1: **4** em cada status, **10** na especialidade;
- cada nível acima: **+1** em cada status, **+2** na especialidade;
- **compra com moedas** (evento `buy`): +1 ponto no status escolhido, custo
  `50 × 1,15^(pontos já comprados naquele status daquele herói)`, arredondado;
- **equipamento**: pontos do item (ver Equipamento).

**Poder** = soma dos pontos dos heróis do esquadrão × 10 × (1 − cansaço).

### De pontos para combate

| Status | Valor |
|---|---|
| ATK | dano por golpe = 4 × pontos |
| DEF | redução de dano = pontos ÷ (pontos + 40) |
| HP | vida = 30 × pontos |
| VEL | golpes por segundo = 0,5 × (1 + pontos ÷ 20) |
| SORTE | chance de crítico = min(60%, 1% × pontos); crítico dá dano ×2 |

- **DPS do herói** = dano × golpes/s × (1 + crítico).
- **Vida efetiva do herói** = vida ÷ (1 − redução).

### XP e nível

- XP para ir do nível L para L+1: `100 × 1,12^(L−1)`, arredondado.
- O XP de cada inimigo vai inteiro para cada herói do esquadrão.

## Torre

Infinita. Cada andar tem **12 inimigos**. O esquadrão caminha **5 min** de
fôlego até cada um e luta com ele; 1 segundo de luta gasta 1 segundo de
fôlego.

- **Moedas** `10 + 3n` e **XP** `20 + 5n` do andar n são divididos entre os 12
  inimigos e entram a cada inimigo vencido (a soma fecha o total exato).
- **Baú**: 20% de chance ao fechar o andar (sorteio por andar e passagem).

### Inimigos

Calibrados por um esquadrão "na média" do andar: heróis no nível que o XP dos
andares anteriores dá, um herói a mais a cada 10 andares (até 3), cada um com
a média dos status dos 5 heróis (justo com qualquer herói inicial).

- **Vida** = DPS dessa média × 300 s: ela leva **5 min** para vencer cada
  inimigo, ou seja, **~2h por andar** (1h caminhando, 1h lutando). Mais forte
  vence mais rápido; a caminhada fica.
- **Dano**: 1 golpe/s no primeiro herói vivo, reduzido pela DEF dele, para a
  média perder **14%** da vida efetiva por luta.
- **Andares 1 a 5**: o inimigo bate só 30% disso, para nenhum herói inicial
  cair no começo.
- A luta é simulada segundo a segundo, como a do chefe, com sorteio fixo.

### Vida e queda

- A vida volta cheia no começo de cada andar e não volta entre as lutas.
  Caminhando, cada herói vivo recupera **2% da vida máxima por minuto** (~10%
  entre um inimigo e outro).
- Herói com vida zero fica caído até o fim do andar; o inimigo passa a bater
  no próximo.
- **Todos caídos**: o esquadrão descansa **10 min** de fôlego e recomeça o
  andar do primeiro inimigo, com vida cheia. Moedas e XP já ganhos ficam.

### Chefe (andares 10, 20, 30…)

- Ao chegar no andar de chefe pela primeira vez, a luta acontece na hora.
- **Luta** simulada em passos de 0,1 s, até 120 s:
  - cada herói vivo golpeia no ritmo da sua VEL; cada golpe pode ser crítico
    (sorteio pela semente do chefe, do dia e da tentativa);
  - o chefe golpeia o primeiro herói vivo do esquadrão, com o dano reduzido
    pela DEF dele;
  - vence quem zera a vida do outro; no tempo esgotado, o chefe vence.
- **Status do chefe** k (andar 10k) vêm de um esquadrão de referência: os
  heróis base, no nível que um herói alcança com o XP dos andares antes do
  chefe, sem compras nem itens, com `min(3, k)` heróis (ATK, DEF e HP como
  especialistas, nessa ordem). O chefe tem vida = DPS da referência × 60 s e
  DPS = vida efetiva da referência ÷ 75 s. A referência vence com folga; o
  **poder recomendado** é o Poder dessa referência.
- **Vitória**: moedas `250 × k`, XP `200 × k`, um **baú de chefe**, libera o
  próximo trecho e, nos chefes 10 a 40, um recrutamento.
- **Derrota**: o esquadrão volta para o andar anterior e treina ali (repete o
  andar, ganhando moedas, XP e baús normalmente). Nova tentativa só pelo
  evento `challenge` (botão Desafiar), uma por dia.

## Equipamento

- Cada herói tem 3 espaços: **arma** (ATK), **armadura** (DEF) e **amuleto**
  (SORTE ou VEL, conforme o item).
- Raridade e pontos do status principal:

  | Raridade | Multiplicador | Status extras |
  |---|---|---|
  | comum | 1 | 0 |
  | rara | 2 | 1 |
  | épica | 3,5 | 1 a 2 |
  | lendária | 6 | 2 |

  Pontos principais = `round(2 × multiplicador × (1 + andar ÷ 10))`; cada extra
  é um status sorteado (diferente do principal) com 40% disso.
- **Baú de andar**: comum 75%, rara 25%. **Baú de chefe**: épica 85%,
  lendária 15%. Tipo e nome do item são sorteados numa lista fixa (dois nomes
  por tipo e raridade).
- O item vai para o **inventário**. Eventos: `equip` (o que estava no espaço
  volta ao inventário), `sell` (moedas = `10 × multiplicador × (1 + andar ÷ 10)`)
  e `refine` (dois itens de mesmo nome e raridade viram um com +1 de refino,
  até +10; cada +1 soma 10% aos pontos do item).

## Ritmo esperado

Gerado por `go run ./cmd/habit-idle balance` (sem compras nem itens; "Kaels"
é o mesmo número de heróis, todos iguais ao Kael, o que menos bate).
Regerar ao mudar `balance.go`.

| Andar | Nível | Heróis | Vida do inimigo | Min/andar | Vida por luta | Min/andar (Kaels) | Vida por luta (Kaels) | Moedas | XP |
|---|---|---|---|---|---|---|---|---|---|
| 1 | 1 | 1 | 4078 | 120 | 4% | 142 | 7% | 13 | 25 |
| 5 | 2 | 1 | 5313 | 120 | 4% | 141 | 6% | 25 | 45 |
| 10 | 4 | 1 | 8133 | 120 | 14% | 140 | 20% | 40 | 70 |
| 15 | 6 | 2 | 22881 | 120 | 14% | 140 | 20% | 55 | 95 |
| 20 | 9 | 2 | 34729 | 120 | 14% | 141 | 19% | 70 | 120 |
| 25 | 11 | 3 | 65968 | 120 | 14% | 141 | 19% | 85 | 145 |
| 30 | 13 | 3 | 81544 | 120 | 14% | 141 | 19% | 100 | 170 |
| 35 | 15 | 3 | 98892 | 120 | 14% | 142 | 18% | 115 | 195 |
| 40 | 17 | 3 | 118080 | 120 | 14% | 142 | 18% | 130 | 220 |
| 45 | 19 | 3 | 139177 | 120 | 14% | 142 | 18% | 145 | 245 |
| 50 | 20 | 3 | 150464 | 120 | 14% | 142 | 18% | 160 | 270 |

| Chefe | Nível da referência | Heróis | Poder recomendado | Vida | Dano/s |
|---|---|---|---|---|---|
| 10 | 4 | 1 | 440 | 2773 | 3 |
| 20 | 9 | 2 | 1480 | 8172 | 14 |
| 30 | 13 | 3 | 2940 | 16537 | 40 |
| 40 | 17 | 3 | 3660 | 23616 | 54 |
| 50 | 20 | 3 | 4200 | 29830 | 65 |
