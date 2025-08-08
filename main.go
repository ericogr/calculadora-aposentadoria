package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

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
	Mes        int
	Patrimonio float64
	Meta       float64
}

func main() {
	inputs := getUserInputs()
	historico, mesesTrabalhados := simularAposentadoriaAnual(inputs)
	exibirResultado(inputs, mesesTrabalhados)
	exibirGraficoAnual(historico)
}

func getUserInputs() Inputs {
	reader := bufio.NewReader(os.Stdin)

	// Valores padrão
	defaults := map[string]string{
		"Idade atual (anos)":                                             "35",
		"Capital inicial disponível hoje em reais":                       "140000.00",
		"Inflação mensal em % (quanto os preços sobem por mês)":          "0.3",
		"Rendimento mensal em % (quanto o capital cresce por mês)":       "0.6",
		"Aporte mensal (quanto você consegue investir por mês) em reais": "1000.00",
		"Renda mensal desejada na aposentadoria (em valores de hoje)":    "1000.00",
		"Expectativa de vida (anos)":                                     "87",
	}

	readWithDefault := func(prompt, def string) string {
		fmt.Printf("%s [padrão: %s]: ", prompt, def)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			return def
		}
		return text
	}

	toInt := func(s string) int {
		v, err := strconv.Atoi(s)
		if err != nil {
			fmt.Printf("Entrada inválida, usando valor padrão %s\n", s)
			return 0
		}
		return v
	}

	toFloat := func(s string) float64 {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			fmt.Printf("Entrada inválida, usando valor padrão %s\n", s)
			return 0
		}
		return v
	}

	idadeAtual := toInt(readWithDefault("Idade atual (anos)", defaults["Idade atual (anos)"]))
	capitalInicial := toFloat(readWithDefault("Capital inicial disponível hoje em reais", defaults["Capital inicial disponível hoje em reais"]))
	inflacaoMensal := toFloat(readWithDefault("Inflação mensal em % (quanto os preços sobem por mês)", defaults["Inflação mensal em % (quanto os preços sobem por mês)"]))
	rendimentoMensal := toFloat(readWithDefault("Rendimento mensal em % (quanto o capital cresce por mês)", defaults["Rendimento mensal em % (quanto o capital cresce por mês)"]))
	aporteMensal := toFloat(readWithDefault("Aporte mensal (quanto você consegue investir por mês) em reais", defaults["Aporte mensal (quanto você consegue investir por mês) em reais"]))
	rendaDesejada := toFloat(readWithDefault("Renda mensal desejada na aposentadoria (em valores de hoje)", defaults["Renda mensal desejada na aposentadoria (em valores de hoje)"]))
	expectativaVida := toInt(readWithDefault("Expectativa de vida (anos)", defaults["Expectativa de vida (anos)"]))

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

func calcularPatrimonioNecessario(rendaInicial float64, mesesRestantes int, rendimentoMensal float64, inflacaoMensal float64) float64 {
	patrimonioNecessario := 0.0
	saque := rendaInicial
	for m := 0; m < mesesRestantes; m++ {
		patrimonioNecessario += saque / math.Pow(1+rendimentoMensal, float64(m+1))
		saque *= (1 + inflacaoMensal)
	}
	return patrimonioNecessario
}

func exibirResultado(inputs Inputs, mesesTrabalhados int) {
	hoje := time.Now()
	dataAposentadoria := hoje.AddDate(0, mesesTrabalhados, 0)
	idadeAposentadoria := inputs.IdadeAtual + mesesTrabalhados/12
	rendaInicialAposentadoria := inputs.RendaDesejada * math.Pow(1+inputs.InflacaoMensal, float64(mesesTrabalhados))

	fmt.Println("\n========= RESULTADO =========")
	fmt.Printf("Meses até a aposentadoria: %d\n", mesesTrabalhados)
	fmt.Printf("Anos até a aposentadoria: %.1f\n", float64(mesesTrabalhados)/12)
	fmt.Printf("Data estimada da aposentadoria: %s\n", dataAposentadoria.Format("02/01/2006"))
	fmt.Printf("Idade na aposentadoria: %d anos\n", idadeAposentadoria)
	fmt.Printf("Renda inicial na aposentadoria (corrigida pela inflação): R$ %.2f\n", rendaInicialAposentadoria)
	fmt.Println("=============================")
}

func drawBar(value, max float64, maxWidth int) string {
	if max == 0 {
		return ""
	}
	length := int((value / max) * float64(maxWidth))
	if length < 0 {
		length = 0
	}
	return strings.Repeat("█", length)
}

func simularAposentadoriaAnual(inputs Inputs) ([]Ponto, int) {
	patrimonio := inputs.CapitalInicial
	mesesTrabalhados := 0
	aporteAtual := inputs.AporteMensal
	historico := make([]Ponto, 0)

	for {
		idade := inputs.IdadeAtual + mesesTrabalhados/12
		mesesRestantesVida := (inputs.ExpectativaVida - idade) * 12

		rendaInicialAposentadoria := inputs.RendaDesejada * math.Pow(1+inputs.InflacaoMensal, float64(mesesTrabalhados))

		patrimonioNecessario := calcularPatrimonioNecessario(
			rendaInicialAposentadoria,
			mesesRestantesVida,
			inputs.RendimentoMensal,
			inputs.InflacaoMensal,
		)

		// Armazena histórico só a cada 12 meses (1 ano)
		if mesesTrabalhados%12 == 0 {
			historico = append(historico, Ponto{
				Mes:        mesesTrabalhados / 12, // aqui Mes vira Ano
				Patrimonio: patrimonio,
				Meta:       patrimonioNecessario, // mantemos para referência, pode ser ignorado no gráfico
			})
		}

		if patrimonio >= patrimonioNecessario {
			break
		}

		patrimonio = patrimonio*(1+inputs.RendimentoMensal) + aporteAtual
		aporteAtual *= (1 + inputs.InflacaoMensal)
		mesesTrabalhados++
	}

	return historico, mesesTrabalhados
}

func exibirGraficoAnual(historico []Ponto) {
	maxWidth := 40
	maxValor := 0.0

	for _, p := range historico {
		if p.Patrimonio > maxValor {
			maxValor = p.Patrimonio
		}
	}

	fmt.Println("\nEvolução anual do Patrimônio (cada barra ~ proporcional ao valor):")
	for _, p := range historico {
		barPat := drawBar(p.Patrimonio, maxValor, maxWidth)

		fmt.Printf("Ano %3d: Patrimônio %-*s\n",
			p.Mes, // aqui Mes é Ano
			maxWidth, barPat)
	}
}
