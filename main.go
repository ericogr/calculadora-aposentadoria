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

// Função para desenhar barra proporcional ao valor, com largura máxima
func drawBar(value, max float64, maxWidth int) string {
	if max == 0 {
		return ""
	}
	length := int((value / max) * float64(maxWidth))
	if length < 0 {
		length = 0
	}
	bar := strings.Repeat("█", length)
	return bar
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Valores padrão
	defaultIdadeAtual := 35
	defaultCapitalInicial := 140000.0
	defaultInflacaoMensal := 0.3
	defaultRendimentoMensal := 0.6
	defaultAporteMensal := 1000.0
	defaultRendaDesejada := 1000.0
	defaultExpectativaVida := 87

	// Função para ler entrada com valor padrão
	readWithDefault := func(prompt string, defaultValue string) string {
		fmt.Print(prompt)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			return defaultValue
		}
		return text
	}

	// Coleta de dados com explicações
	idadeAtualStr := readWithDefault(
		fmt.Sprintf("Idade atual (anos) [padrão: %d]: ", defaultIdadeAtual),
		fmt.Sprintf("%d", defaultIdadeAtual))
	idadeAtual, _ := strconv.Atoi(idadeAtualStr)

	capitalInicialStr := readWithDefault(
		fmt.Sprintf("Capital inicial disponível hoje em reais [padrão: %.2f]: ", defaultCapitalInicial),
		fmt.Sprintf("%.2f", defaultCapitalInicial))
	capitalInicial, _ := strconv.ParseFloat(capitalInicialStr, 64)

	inflacaoMensalStr := readWithDefault(
		fmt.Sprintf("Inflação mensal em %% (quanto os preços sobem por mês) [padrão: %.2f]: ", defaultInflacaoMensal),
		fmt.Sprintf("%.2f", defaultInflacaoMensal))
	inflacaoMensal, _ := strconv.ParseFloat(inflacaoMensalStr, 64)

	rendimentoMensalStr := readWithDefault(
		fmt.Sprintf("Rendimento mensal em %% (quanto o capital cresce por mês) [padrão: %.2f]: ", defaultRendimentoMensal),
		fmt.Sprintf("%.2f", defaultRendimentoMensal))
	rendimentoMensal, _ := strconv.ParseFloat(rendimentoMensalStr, 64)

	aporteMensalStr := readWithDefault(
		fmt.Sprintf("Aporte mensal (quanto você consegue investir por mês) em reais [padrão: %.2f]: ", defaultAporteMensal),
		fmt.Sprintf("%.2f", defaultAporteMensal))
	aporteMensal, _ := strconv.ParseFloat(aporteMensalStr, 64)

	rendaDesejadaStr := readWithDefault(
		fmt.Sprintf("Renda mensal desejada na aposentadoria (em valores de hoje) [padrão: %.2f]: ", defaultRendaDesejada),
		fmt.Sprintf("%.2f", defaultRendaDesejada))
	rendaDesejada, _ := strconv.ParseFloat(rendaDesejadaStr, 64)

	expectativaVidaStr := readWithDefault(
		fmt.Sprintf("Expectativa de vida (anos) [padrão: %d]: ", defaultExpectativaVida),
		fmt.Sprintf("%d", defaultExpectativaVida))
	expectativaVida, _ := strconv.Atoi(expectativaVidaStr)

	// Converte percentuais para decimais
	inflacaoMensal /= 100
	rendimentoMensal /= 100

	// Simulação mês a mês em valores NOMINAIS
	hoje := time.Now()
	patrimonio := capitalInicial
	mesesTrabalhados := 0
	aporteAtual := aporteMensal

	// Armazena histórico para gráfico
	type ponto struct {
		mes        int
		patrimonio float64
		meta       float64
	}
	var historico []ponto

	for {
		idade := idadeAtual + mesesTrabalhados/12
		mesesRestantesVida := (expectativaVida - idade) * 12

		rendaInicialAposentadoria := rendaDesejada * math.Pow(1+inflacaoMensal, float64(mesesTrabalhados))

		// Calcula patrimônio necessário para saques crescentes pela inflação
		patrimonioNecessario := 0.0
		saque := rendaInicialAposentadoria
		for m := 0; m < mesesRestantesVida; m++ {
			patrimonioNecessario += saque / math.Pow(1+rendimentoMensal, float64(m+1))
			saque *= (1 + inflacaoMensal)
		}

		historico = append(historico, ponto{
			mes:        mesesTrabalhados,
			patrimonio: patrimonio,
			meta:       patrimonioNecessario,
		})

		if patrimonio >= patrimonioNecessario {
			break
		}

		patrimonio *= (1 + rendimentoMensal)
		patrimonio += aporteAtual
		aporteAtual *= (1 + inflacaoMensal)

		mesesTrabalhados++
	}

	// Resultado final
	dataAposentadoria := hoje.AddDate(0, mesesTrabalhados, 0)
	idadeAposentadoria := idadeAtual + mesesTrabalhados/12
	rendaInicialAposentadoria := rendaDesejada * math.Pow(1+inflacaoMensal, float64(mesesTrabalhados))

	fmt.Println("\n========= RESULTADO =========")
	fmt.Printf("Meses até a aposentadoria: %d\n", mesesTrabalhados)
	fmt.Printf("Anos até a aposentadoria: %.1f\n", float64(mesesTrabalhados)/12)
	fmt.Printf("Data estimada da aposentadoria: %s\n", dataAposentadoria.Format("02/01/2006"))
	fmt.Printf("Idade na aposentadoria: %d anos\n", idadeAposentadoria)
	fmt.Printf("Renda inicial na aposentadoria (corrigida pela inflação): R$ %.2f\n", rendaInicialAposentadoria)
	fmt.Println("=============================")

	// --- Gráfico ASCII de barras comparativas ---

	maxWidth := 40
	var maxValor float64
	for _, p := range historico {
		if p.patrimonio > maxValor {
			maxValor = p.patrimonio
		}
		if p.meta > maxValor {
			maxValor = p.meta
		}
	}

	fmt.Println("\nEvolução do Patrimônio vs Meta (cada barra ~ proporcional ao valor):")
	for _, p := range historico {
		barPat := drawBar(p.patrimonio, maxValor, maxWidth)
		barMeta := drawBar(p.meta, maxValor, maxWidth)

		fmt.Printf("Mês %3d: Patrimônio %-*s  Meta %-*s\n",
			p.mes,
			maxWidth, barPat,
			maxWidth, barMeta)
		if p.mes%12 == 0 && p.mes != 0 {
			fmt.Println()
		}
	}
}
