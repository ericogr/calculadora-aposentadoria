package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// Limite superior aceito para a expectativa de vida (anos).
const maxExpectativaVida = 130

type Inputs struct {
	IdadeAtual       int
	CapitalInicial   float64
	InflacaoMensal   float64
	RendimentoMensal float64
	AporteMensal     float64
	RendaDesejada    float64
	ExpectativaVida  int
}

type Ponto struct {
	Mes        int // meses a partir de hoje
	Patrimonio float64
	Meta       float64
}

type Resultado struct {
	Historico        []Ponto
	MesesTrabalhados int  // só tem significado quando Atingiu é true
	Atingiu          bool // false: a meta não é alcançada antes da expectativa de vida
}

func main() {
	inputs := getUserInputs()
	if err := inputs.Validar(); err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}

	resultado := simularAposentadoria(inputs)
	exibirResultado(inputs, resultado)
	exibirLimitacoes(os.Stdout)
	exibirGraficoAnual(resultado.Historico)
}

func (in Inputs) Validar() error {
	switch {
	case in.IdadeAtual < 0:
		return fmt.Errorf("a idade atual não pode ser negativa")
	case in.ExpectativaVida > maxExpectativaVida:
		return fmt.Errorf("a expectativa de vida deve ser de no máximo %d anos", maxExpectativaVida)
	case in.IdadeAtual >= in.ExpectativaVida:
		return fmt.Errorf("a idade atual (%d) deve ser menor que a expectativa de vida (%d)", in.IdadeAtual, in.ExpectativaVida)
	case in.CapitalInicial < 0:
		return fmt.Errorf("o capital inicial não pode ser negativo")
	case in.AporteMensal < 0:
		return fmt.Errorf("o aporte mensal não pode ser negativo")
	case in.RendaDesejada <= 0:
		return fmt.Errorf("a renda mensal desejada deve ser maior que zero")
	case in.InflacaoMensal <= -1:
		return fmt.Errorf("a inflação mensal deve ser maior que -100%%")
	case in.RendimentoMensal <= -1:
		return fmt.Errorf("o rendimento mensal deve ser maior que -100%%")
	}
	return nil
}

func getUserInputs() Inputs {
	return lerInputs(os.Stdin, os.Stdout)
}

func lerInputs(r io.Reader, w io.Writer) Inputs {
	reader := bufio.NewReader(r)

	readWithDefault := func(prompt, def string) string {
		fmt.Fprintf(w, "%s [padrão: %s]: ", prompt, def)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			return def
		}
		return text
	}

	toInt := func(prompt, def string) int {
		s := readWithDefault(prompt, def)
		v, err := strconv.Atoi(s)
		if err != nil {
			fmt.Fprintf(w, "Entrada inválida %q, usando valor padrão %s\n", s, def)
			v, _ = strconv.Atoi(def)
		}
		return v
	}

	toFloat := func(prompt, def string) float64 {
		s := readWithDefault(prompt, def)
		v, err := parseFloat(s)
		if err != nil {
			fmt.Fprintf(w, "Entrada inválida %q, usando valor padrão %s\n", s, def)
			v, _ = parseFloat(def)
		}
		return v
	}

	idadeAtual := toInt("Idade atual (anos)", "35")
	capitalInicial := toFloat("Capital inicial disponível hoje em reais", "140000.00")
	inflacaoMensal := toFloat("Inflação mensal em % (quanto os preços sobem por mês)", "0.3")
	rendimentoMensal := toFloat("Rendimento mensal em % (quanto o capital cresce por mês)", "0.6")
	aporteMensal := toFloat("Aporte mensal (quanto você consegue investir por mês) em reais", "1000.00")
	rendaDesejada := toFloat("Renda mensal desejada na aposentadoria (em valores de hoje)", "1000.00")
	expectativaVida := toInt("Expectativa de vida (anos)", "87")

	// Converte percentuais para decimais
	inflacaoMensal /= 100
	rendimentoMensal /= 100

	return Inputs{
		IdadeAtual:       idadeAtual,
		CapitalInicial:   capitalInicial,
		InflacaoMensal:   inflacaoMensal,
		RendimentoMensal: rendimentoMensal,
		AporteMensal:     aporteMensal,
		RendaDesejada:    rendaDesejada,
		ExpectativaVida:  expectativaVida,
	}
}

// parseFloat aceita vírgula ou ponto como separador decimal. Quando há vírgula,
// os pontos são tratados como separador de milhar ("1.000,50" -> 1000.50).
func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("valor inválido: %s", s)
	}
	return v, nil
}

// calcularPatrimonioNecessario devolve o valor presente de mesesRestantes saques
// mensais feitos no início de cada mês: o primeiro de rendaInicial, na data da
// aposentadoria, e os seguintes corrigidos pela inflação e descontados pelo rendimento.
func calcularPatrimonioNecessario(rendaInicial float64, mesesRestantes int, rendimentoMensal float64, inflacaoMensal float64) float64 {
	razao := (1 + inflacaoMensal) / (1 + rendimentoMensal)
	patrimonioNecessario := 0.0
	fator := 1.0
	for m := 0; m < mesesRestantes; m++ {
		patrimonioNecessario += rendaInicial * fator
		fator *= razao
	}
	return patrimonioNecessario
}

// adicionarMeses soma meses a t limitando o dia ao último dia do mês de destino
// (31/01 + 1 mês = 28/02, e não 03/03).
func adicionarMeses(t time.Time, meses int) time.Time {
	ano, mes, dia := t.Date()
	primeiroDoMes := time.Date(ano, mes+time.Month(meses), 1, 0, 0, 0, 0, t.Location())
	if ultimo := primeiroDoMes.AddDate(0, 1, -1).Day(); dia > ultimo {
		dia = ultimo
	}
	return time.Date(primeiroDoMes.Year(), primeiroDoMes.Month(), dia, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func exibirResultado(inputs Inputs, resultado Resultado) {
	fmt.Println("\n========= RESULTADO =========")
	if !resultado.Atingiu {
		fmt.Printf("Meta não atingida: com os dados informados o patrimônio não sustenta a renda\n")
		fmt.Printf("desejada em nenhum momento antes dos %d anos (expectativa de vida).\n", inputs.ExpectativaVida)
		fmt.Println("Tente aumentar o aporte, o capital inicial ou o rendimento, ou reduzir a renda desejada.")
		fmt.Println("=============================")
		return
	}

	mesesTrabalhados := resultado.MesesTrabalhados
	dataAposentadoria := adicionarMeses(time.Now(), mesesTrabalhados)
	idadeAposentadoria := inputs.IdadeAtual + mesesTrabalhados/12
	rendaInicialAposentadoria := inputs.RendaDesejada * math.Pow(1+inputs.InflacaoMensal, float64(mesesTrabalhados))

	fmt.Printf("Meses até a aposentadoria: %d\n", mesesTrabalhados)
	fmt.Printf("Anos até a aposentadoria: %.1f\n", float64(mesesTrabalhados)/12)
	fmt.Printf("Data estimada da aposentadoria: %s\n", dataAposentadoria.Format("02/01/2006"))
	fmt.Printf("Idade na aposentadoria: %d anos\n", idadeAposentadoria)
	fmt.Printf("Renda inicial na aposentadoria (corrigida pela inflação): R$ %.2f\n", rendaInicialAposentadoria)
	fmt.Println("=============================")
}

// exibirLimitacoes lista as premissas e limitações do modelo.
func exibirLimitacoes(w io.Writer) {
	linhas := []string{
		"",
		"--- Limitações deste cálculo ---",
		"- Patrimônio consumido: os saques esgotam o patrimônio exatamente na expectativa de",
		"  vida. Não há perpetuidade: o principal é gasto e nada sobra no final.",
		"- Sem margem de segurança: a aposentadoria é declarada no primeiro mês em que o",
		"  patrimônio iguala a meta. Se você viver mais ou o rendimento ficar abaixo do",
		"  informado, o dinheiro não dura.",
		"- Taxas constantes: rendimento e inflação são fixos todos os meses, sem oscilação de",
		"  mercado. O prazo é muito sensível ao rendimento real, (1+rendimento)/(1+inflação)-1;",
		"  vale testar valores mais conservadores.",
		"- Sem impostos e taxas: não há IR sobre rendimentos nem taxa de administração.",
		"- Sem outras rendas: INSS, previdência privada e aluguéis não são considerados.",
		"- Convenções de tempo: saques no início do mês; aportes no fim do mês (o primeiro sem",
		"  correção pela inflação); a idade atual é tratada como completa hoje.",
		"- Taxas devem ser mensais e nominais (não reais). Converta taxas anuais antes de digitar.",
	}
	for _, l := range linhas {
		fmt.Fprintln(w, l)
	}
}

func drawBar(value, maximo float64, maxWidth int) string {
	if maximo <= 0 {
		return ""
	}
	length := int((value / maximo) * float64(maxWidth))
	length = max(0, min(length, maxWidth))
	return strings.Repeat("█", length)
}

// simularAposentadoria avança mês a mês até que o patrimônio acumulado cubra a
// meta (patrimônio necessário para sustentar a renda desejada até a expectativa
// de vida). O horizonte de vida é contado em meses exatos, e a simulação termina
// no máximo na expectativa de vida, mesmo que a meta nunca seja atingida.
func simularAposentadoria(inputs Inputs) Resultado {
	totalMeses := (inputs.ExpectativaVida - inputs.IdadeAtual) * 12
	patrimonio := inputs.CapitalInicial
	aporteAtual := inputs.AporteMensal
	rendaAtual := inputs.RendaDesejada
	historico := make([]Ponto, 0)

	for mes := 0; mes < totalMeses; mes++ {
		patrimonioNecessario := calcularPatrimonioNecessario(
			rendaAtual,
			totalMeses-mes,
			inputs.RendimentoMensal,
			inputs.InflacaoMensal,
		)
		atingiu := patrimonio >= patrimonioNecessario

		// Histórico a cada 12 meses, mais o último mês simulado
		if mes%12 == 0 || atingiu || mes == totalMeses-1 {
			historico = append(historico, Ponto{
				Mes:        mes,
				Patrimonio: patrimonio,
				Meta:       patrimonioNecessario,
			})
		}

		if atingiu {
			return Resultado{Historico: historico, MesesTrabalhados: mes, Atingiu: true}
		}

		patrimonio = patrimonio*(1+inputs.RendimentoMensal) + aporteAtual
		aporteAtual *= (1 + inputs.InflacaoMensal)
		rendaAtual *= (1 + inputs.InflacaoMensal)
	}

	return Resultado{Historico: historico}
}

func exibirGraficoAnual(historico []Ponto) {
	maxWidth := 40
	maxValor := 0.0

	for _, p := range historico {
		maxValor = max(maxValor, p.Patrimonio, p.Meta)
	}

	fmt.Println("\nEvolução do Patrimônio e da Meta (barras proporcionais ao valor):")
	for _, p := range historico {
		fmt.Printf("Ano %5.1f: Patrimônio %-*s R$ %12.2f\n",
			float64(p.Mes)/12, maxWidth, drawBar(p.Patrimonio, maxValor, maxWidth), p.Patrimonio)
		fmt.Printf("           Meta       %-*s R$ %12.2f\n",
			maxWidth, drawBar(p.Meta, maxValor, maxWidth), p.Meta)
	}
}
