package core

import "time"

// Day é um dia do calendário, "AAAA-MM-DD". Contas de dia não usam fuso.
type Day = string

const layout = "2006-01-02"

// Calendar diz a que dia um instante pertence: no fuso Loc, com o dia virando
// DayStart minutos depois da meia-noite (a madrugada antes disso ainda é o dia
// anterior).
type Calendar struct {
	Loc      *time.Location
	DayStart int
}

func (c Calendar) DayOf(t time.Time) Day {
	local := t.In(c.Loc)
	return local.Add(-time.Duration(c.DayStart) * time.Minute).Format(layout)
}

// Start é o instante em que day começa (a virada).
func (c Calendar) Start(day Day) time.Time {
	d := parse(day)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, c.DayStart, 0, 0, c.Loc)
}

func parse(day Day) time.Time {
	t, err := time.Parse(layout, day)
	if err != nil {
		panic("dia inválido: " + day)
	}
	return t
}

func AddDays(day Day, n int) Day {
	return parse(day).AddDate(0, 0, n).Format(layout)
}

// Weekday do dia, 0 = domingo.
func Weekday(day Day) int {
	return int(parse(day).Weekday())
}

// WeekOf é a segunda-feira da semana de day.
func WeekOf(day Day) Day {
	return AddDays(day, -((Weekday(day) + 6) % 7))
}
