package main

import (
	"bytes"
	"io"
	"math"
	"strings"
	"testing"
	"time"
)

func inputsPadrao() Inputs {
	return Inputs{
		IdadeAtual:       35,
		CapitalInicial:   140000,
		InflacaoMensal:   0.003,
		RendimentoMensal: 0.006,
		AporteMensal:     1000,
		RendaDesejada:    1000,
		ExpectativaVida:  87,
	}
}

func proximo(t *testing.T, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("got %.6f, want %.6f (tol %.1e)", got, want, tol)
	}
}

// Valor presente de uma anuidade antecipada crescente em fórmula fechada:
// W * (1 - q^n) / (1 - q), com q = (1+g)/(1+r).
func pvFechada(w float64, n int, r, g float64) float64 {
	q := (1 + g) / (1 + r)
	if math.Abs(q-1) < 1e-15 {
		return w * float64(n)
	}
	return w * (1 - math.Pow(q, float64(n))) / (1 - q)
}

func TestCalcularPatrimonioNecessarioFormulaFechada(t *testing.T) {
	casos := []struct {
		nome string
		r, g float64
		n    int
	}{
		{"rendimento maior que inflação", 0.006, 0.003, 600},
		{"rendimento igual à inflação", 0.006, 0.006, 600},
		{"rendimento menor que inflação", 0.003, 0.006, 600},
		{"sem rendimento nem inflação", 0, 0, 120},
		{"um mês", 0.006, 0.003, 1},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			got := calcularPatrimonioNecessario(1000, c.n, c.r, c.g)
			proximo(t, got, pvFechada(1000, c.n, c.r, c.g), 1e-6)
		})
	}
}

func TestCalcularPatrimonioNecessarioSemMeses(t *testing.T) {
	for _, n := range []int{0, -5} {
		if got := calcularPatrimonioNecessario(1000, n, 0.006, 0.003); got != 0 {
			t.Errorf("n=%d: got %v, want 0", n, got)
		}
	}
}

func TestPrimeiroSaqueNaoEDescontado(t *testing.T) {
	// Saque no início do mês: com 1 mês restante, a meta é exatamente a renda.
	proximo(t, calcularPatrimonioNecessario(1234.56, 1, 0.01, 0.005), 1234.56, 1e-9)
}

// Simulação independente (sem reutilizar simularAposentadoria) para conferir o
// primeiro mês em que o patrimônio cobre a meta.
func mesDeAposentadoriaReferencia(in Inputs) (int, bool) {
	total := (in.ExpectativaVida - in.IdadeAtual) * 12
	pat, aporte := in.CapitalInicial, in.AporteMensal
	for m := 0; m < total; m++ {
		renda := in.RendaDesejada * math.Pow(1+in.InflacaoMensal, float64(m))
		if pat >= pvFechada(renda, total-m, in.RendimentoMensal, in.InflacaoMensal) {
			return m, true
		}
		pat = pat*(1+in.RendimentoMensal) + aporte
		aporte *= 1 + in.InflacaoMensal
	}
	return 0, false
}

func TestSimularConfereComReferencia(t *testing.T) {
	variacoes := []func(*Inputs){
		func(in *Inputs) {},
		func(in *Inputs) { in.RendaDesejada = 2000 },
		func(in *Inputs) { in.RendaDesejada = 5000 },
		func(in *Inputs) { in.IdadeAtual = 25; in.AporteMensal = 3000 },
		func(in *Inputs) { in.RendimentoMensal = in.InflacaoMensal },
		func(in *Inputs) { in.CapitalInicial = 0; in.AporteMensal = 2500; in.RendaDesejada = 3000 },
	}
	for i, v := range variacoes {
		in := inputsPadrao()
		v(&in)
		want, wantOk := mesDeAposentadoriaReferencia(in)
		got := simularAposentadoria(in)
		if got.Atingiu != wantOk || (wantOk && got.MesesTrabalhados != want) {
			t.Errorf("variação %d: got (meses=%d, atingiu=%v), want (meses=%d, atingiu=%v)",
				i, got.MesesTrabalhados, got.Atingiu, want, wantOk)
		}
	}
}

func TestSimularCenarioPadrao(t *testing.T) {
	res := simularAposentadoria(inputsPadrao())
	if !res.Atingiu {
		t.Fatal("cenário padrão deveria atingir a meta")
	}
	if res.MesesTrabalhados != 81 {
		t.Errorf("meses = %d, want 81", res.MesesTrabalhados)
	}
}

func TestSimularMinimalidade(t *testing.T) {
	// No mês da aposentadoria patrimônio >= meta; no mês anterior, não.
	in := inputsPadrao()
	res := simularAposentadoria(in)
	if !res.Atingiu || res.MesesTrabalhados == 0 {
		t.Fatalf("resultado inesperado: %+v", res)
	}
	ultimo := res.Historico[len(res.Historico)-1]
	if ultimo.Mes != res.MesesTrabalhados {
		t.Errorf("último ponto no mês %d, want %d", ultimo.Mes, res.MesesTrabalhados)
	}
	if ultimo.Patrimonio < ultimo.Meta {
		t.Errorf("patrimônio %.2f < meta %.2f no mês da aposentadoria", ultimo.Patrimonio, ultimo.Meta)
	}

	pat, aporte := in.CapitalInicial, in.AporteMensal
	for m := 0; m < res.MesesTrabalhados-1; m++ {
		pat = pat*(1+in.RendimentoMensal) + aporte
		aporte *= 1 + in.InflacaoMensal
	}
	m := res.MesesTrabalhados - 1
	total := (in.ExpectativaVida - in.IdadeAtual) * 12
	renda := in.RendaDesejada * math.Pow(1+in.InflacaoMensal, float64(m))
	if meta := pvFechada(renda, total-m, in.RendimentoMensal, in.InflacaoMensal); pat >= meta {
		t.Errorf("já havia atingido a meta no mês %d (%.2f >= %.2f)", m, pat, meta)
	}
}

func TestSimularHorizonteEmMesesExatos(t *testing.T) {
	// 12 meses de vida restantes, sem capital nem aporte: a meta no mês 0 cobre
	// 12 saques e, no último mês simulado (11), apenas 1 saque.
	in := inputsPadrao()
	in.IdadeAtual, in.ExpectativaVida = 35, 36
	in.CapitalInicial, in.AporteMensal = 0, 0

	res := simularAposentadoria(in)
	if res.Atingiu {
		t.Fatal("não deveria atingir a meta sem capital nem aporte")
	}
	if len(res.Historico) != 2 {
		t.Fatalf("histórico com %d pontos, want 2: %+v", len(res.Historico), res.Historico)
	}
	proximo(t, res.Historico[0].Meta, pvFechada(in.RendaDesejada, 12, in.RendimentoMensal, in.InflacaoMensal), 1e-6)

	ult := res.Historico[1]
	if ult.Mes != 11 {
		t.Errorf("último ponto no mês %d, want 11", ult.Mes)
	}
	proximo(t, ult.Meta, in.RendaDesejada*math.Pow(1+in.InflacaoMensal, 11), 1e-6)
}

func TestSimularImediato(t *testing.T) {
	in := inputsPadrao()
	in.CapitalInicial = 10_000_000
	res := simularAposentadoria(in)
	if !res.Atingiu || res.MesesTrabalhados != 0 {
		t.Errorf("got %+v, want atingiu com 0 meses", res)
	}
	if len(res.Historico) != 1 {
		t.Errorf("histórico com %d pontos, want 1", len(res.Historico))
	}
}

func TestSimularPatrimonioIgualAMetaAtinge(t *testing.T) {
	// Fronteira: patrimônio exatamente igual à meta já basta (>=, não >).
	in := inputsPadrao()
	total := (in.ExpectativaVida - in.IdadeAtual) * 12
	in.CapitalInicial = calcularPatrimonioNecessario(in.RendaDesejada, total, in.RendimentoMensal, in.InflacaoMensal)
	res := simularAposentadoria(in)
	if !res.Atingiu || res.MesesTrabalhados != 0 {
		t.Errorf("got %+v, want atingiu com 0 meses", res)
	}
}

func TestSimularInalcancavel(t *testing.T) {
	in := inputsPadrao()
	in.CapitalInicial, in.AporteMensal, in.RendaDesejada = 1000, 0, 10000
	res := simularAposentadoria(in)
	if res.Atingiu {
		t.Errorf("meta não deveria ser atingida: %+v", res)
	}
	if len(res.Historico) == 0 {
		t.Error("histórico vazio")
	}
}

func TestSimularTerminaComCapitalNegativo(t *testing.T) {
	// Entradas assim são rejeitadas por Validar, mas a simulação nunca deve travar.
	in := inputsPadrao()
	in.CapitalInicial, in.AporteMensal = -50000, 0
	if res := simularAposentadoria(in); res.Atingiu {
		t.Errorf("meta não deveria ser atingida: %+v", res)
	}
}

func TestHistoricoIncluiPontoFinal(t *testing.T) {
	res := simularAposentadoria(inputsPadrao()) // 81 meses: não é múltiplo de 12
	meses := make([]int, len(res.Historico))
	for i, p := range res.Historico {
		meses[i] = p.Mes
	}
	want := []int{0, 12, 24, 36, 48, 60, 72, 81}
	if len(meses) != len(want) {
		t.Fatalf("meses = %v, want %v", meses, want)
	}
	for i := range want {
		if meses[i] != want[i] {
			t.Fatalf("meses = %v, want %v", meses, want)
		}
	}
}

func TestValidar(t *testing.T) {
	casos := []struct {
		nome   string
		altera func(*Inputs)
		valido bool
		contem string
	}{
		{"padrão", func(in *Inputs) {}, true, ""},
		{"idade negativa", func(in *Inputs) { in.IdadeAtual = -1 }, false, "idade"},
		{"idade igual à expectativa", func(in *Inputs) { in.IdadeAtual = 87 }, false, "menor"},
		{"idade maior que a expectativa", func(in *Inputs) { in.IdadeAtual = 90 }, false, "menor"},
		{"expectativa absurda", func(in *Inputs) { in.ExpectativaVida = 1_000_000 }, false, "máximo"},
		{"capital negativo", func(in *Inputs) { in.CapitalInicial = -1 }, false, "capital"},
		{"capital zero", func(in *Inputs) { in.CapitalInicial = 0 }, true, ""},
		{"aporte negativo", func(in *Inputs) { in.AporteMensal = -1 }, false, "aporte"},
		{"renda zero", func(in *Inputs) { in.RendaDesejada = 0 }, false, "renda"},
		{"inflação -100%", func(in *Inputs) { in.InflacaoMensal = -1 }, false, "inflação"},
		{"rendimento -100%", func(in *Inputs) { in.RendimentoMensal = -1 }, false, "rendimento"},
		{"deflação moderada", func(in *Inputs) { in.InflacaoMensal = -0.001 }, true, ""},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			in := inputsPadrao()
			c.altera(&in)
			err := in.Validar()
			if c.valido && err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if !c.valido {
				if err == nil {
					t.Fatal("esperava erro")
				}
				if !strings.Contains(err.Error(), c.contem) {
					t.Errorf("erro %q não contém %q", err, c.contem)
				}
			}
		})
	}
}

func TestParseFloat(t *testing.T) {
	validos := map[string]float64{
		"0.3":          0.3,
		"0,3":          0.3,
		" 0,6 ":        0.6,
		"140000":       140000,
		"140000,50":    140000.5,
		"1.000,50":     1000.5,
		"1.234.567,89": 1234567.89,
		"-2":           -2,
	}
	for s, want := range validos {
		got, err := parseFloat(s)
		if err != nil {
			t.Errorf("parseFloat(%q): erro %v", s, err)
			continue
		}
		proximo(t, got, want, 1e-9)
	}
	for _, s := range []string{"", "abc", "NaN", "Inf", "-Inf", "1,2,3", "R$ 10"} {
		if v, err := parseFloat(s); err == nil {
			t.Errorf("parseFloat(%q) = %v, esperava erro", s, v)
		}
	}
}

func TestLerInputsPadroes(t *testing.T) {
	var saida bytes.Buffer
	got := lerInputs(strings.NewReader("\n\n\n\n\n\n\n"), &saida)
	if got != inputsPadrao() {
		t.Errorf("got %+v, want %+v", got, inputsPadrao())
	}
}

func TestLerInputsVirgulaDecimal(t *testing.T) {
	entrada := "40\n200.000,00\n0,4\n0,7\n1500,50\n2000\n90\n"
	got := lerInputs(strings.NewReader(entrada), io.Discard)
	if got.IdadeAtual != 40 || got.ExpectativaVida != 90 {
		t.Errorf("idades: got %d e %d, want 40 e 90", got.IdadeAtual, got.ExpectativaVida)
	}
	proximo(t, got.CapitalInicial, 200000, 1e-9)
	proximo(t, got.InflacaoMensal, 0.004, 1e-12)
	proximo(t, got.RendimentoMensal, 0.007, 1e-12)
	proximo(t, got.AporteMensal, 1500.5, 1e-9)
	proximo(t, got.RendaDesejada, 2000, 1e-9)
}

func TestLerInputsInvalidoUsaPadrao(t *testing.T) {
	var saida bytes.Buffer
	got := lerInputs(strings.NewReader("abc\nxyz\n\n\n\n\n1.5\n"), &saida)
	if got != inputsPadrao() {
		t.Errorf("got %+v, want %+v", got, inputsPadrao())
	}
	if n := strings.Count(saida.String(), "Entrada inválida"); n != 3 {
		t.Errorf("%d avisos de entrada inválida, want 3:\n%s", n, saida.String())
	}
}

func TestLerInputsEOF(t *testing.T) {
	got := lerInputs(strings.NewReader(""), io.Discard)
	if got != inputsPadrao() {
		t.Errorf("got %+v, want %+v", got, inputsPadrao())
	}
}

func TestAdicionarMeses(t *testing.T) {
	data := func(a int, m time.Month, d int) time.Time {
		return time.Date(a, m, d, 10, 30, 0, 0, time.UTC)
	}
	casos := []struct {
		nome   string
		inicio time.Time
		meses  int
		want   time.Time
	}{
		{"zero meses", data(2026, 10, 6), 0, data(2026, 10, 6)},
		{"mês simples", data(2026, 10, 6), 3, data(2027, 1, 6)},
		{"31/01 + 1 não transborda para março", data(2026, 1, 31), 1, data(2026, 2, 28)},
		{"29/02 de ano bissexto", data(2028, 1, 31), 1, data(2028, 2, 29)},
		{"31/03 + 1 mês", data(2026, 3, 31), 1, data(2026, 4, 30)},
		{"31/08 + 6 meses", data(2026, 8, 31), 6, data(2027, 2, 28)},
		{"dezembro + 1", data(2026, 12, 15), 1, data(2027, 1, 15)},
		{"vários anos", data(2026, 10, 6), 81, data(2033, 7, 6)},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if got := adicionarMeses(c.inicio, c.meses); !got.Equal(c.want) {
				t.Errorf("got %s, want %s", got, c.want)
			}
		})
	}
}

func TestDrawBar(t *testing.T) {
	casos := []struct {
		valor, maximo float64
		largura       int
		want          int
	}{
		{50, 100, 40, 20},
		{100, 100, 40, 40},
		{0, 100, 40, 0},
		{-10, 100, 40, 0},
		{200, 100, 40, 40}, // nunca passa da largura
		{10, 0, 40, 0},
	}
	for _, c := range casos {
		if got := strings.Count(drawBar(c.valor, c.maximo, c.largura), "█"); got != c.want {
			t.Errorf("drawBar(%v, %v, %d) = %d barras, want %d", c.valor, c.maximo, c.largura, got, c.want)
		}
	}
}

func TestExibirLimitacoes(t *testing.T) {
	var saida bytes.Buffer
	exibirLimitacoes(&saida)
	texto := saida.String()
	for _, trecho := range []string{
		"Limitações deste cálculo",
		"Patrimônio consumido",
		"margem de segurança",
		"Taxas constantes",
		"impostos",
		"INSS",
		"início do mês",
		"mensais e nominais",
	} {
		if !strings.Contains(texto, trecho) {
			t.Errorf("limitações não mencionam %q:\n%s", trecho, texto)
		}
	}
	if strings.Contains(strings.ToLower(texto), "outras calculadoras") {
		t.Errorf("o texto não deve comparar com outras calculadoras:\n%s", texto)
	}
}
