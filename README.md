# Calculadora de Aposentadoria

Este programa em Go simula a sua jornada para a aposentadoria, calculando em quanto tempo você alcançará a independência financeira com base em suas informações atuais.

## Para que serve?

A calculadora ajuda a visualizar o crescimento do seu patrimônio ao longo do tempo, considerando variáveis como aportes mensais, rendimentos e inflação. O objetivo é estimar a data em que você poderá se aposentar com a renda mensal desejada.

## Como usar

1.  **Compile e execute o programa:**

    ```bash
    make run
    ```

2.  **Responda às perguntas:**

    O programa solicitará algumas informações financeiras. Você pode simplesmente pressionar `Enter` para usar os valores padrão sugeridos entre colchetes `[]`.

    Os números aceitam vírgula ou ponto como separador decimal (`0,3` ou `0.3`). Se houver vírgula, os pontos são tratados como separador de milhar (`1.000,50` = 1000,50). Se o texto digitado não for um número válido, o programa avisa e usa o valor padrão.

    **Campos solicitados:**

    *   `Idade atual (anos)`: Sua idade hoje.
    *   `Capital inicial disponível hoje em reais`: Quanto dinheiro você já tem investido.
    *   `Inflação mensal em %`: A taxa média de aumento dos preços (ex: 0.3 para 0.3%).
    *   `Rendimento mensal em %`: O crescimento médio dos seus investimentos (ex: 0.6 para 0.6%).
    *   `Aporte mensal em reais`: O valor que você investe todo mês.
    *   `Renda mensal desejada na aposentadoria`: Quanto você gostaria de receber por mês ao se aposentar (em valores de hoje).
    *   `Expectativa de vida (anos)`: Até que idade você espera viver.

    **Validações:** a idade atual deve ser menor que a expectativa de vida (no máximo 130 anos); capital inicial e aporte não podem ser negativos; a renda desejada deve ser maior que zero; inflação e rendimento devem ser maiores que -100%. Com dados inválidos o programa mostra o erro e encerra.

3.  **Veja o resultado:**

    Após preencher os dados, o programa exibirá:

    *   O tempo restante para a aposentadoria (em meses e anos).
    *   A data estimada para a aposentadoria.
    *   Sua idade ao se aposentar.
    *   A renda inicial na aposentadoria (corrigida pela inflação).

    Se, com os dados informados, o patrimônio não conseguir sustentar a renda desejada em nenhum momento antes da expectativa de vida, o programa avisa que a **meta não foi atingida** em vez de informar uma data de aposentadoria.

## Explicação Detalhada

O programa funciona como um simulador financeiro mês a mês, projetando o futuro do seu patrimônio com base em duas lógicas principais:

1.  **Fase de Acumulação (enquanto você trabalha):**

    *   A cada mês, seu patrimônio cresce com base na `taxa de rendimento mensal`.
    *   Você adiciona o `aporte mensal` ao fim de cada mês; o primeiro é o valor informado e os seguintes são corrigidos pela `inflação` para manter seu poder de compra.
    *   O programa continua simulando até que seu patrimônio atinja o valor necessário para a aposentadoria ou até a `expectativa de vida`, o que ocorrer primeiro.

2.  **Cálculo da Meta de Aposentadoria:**

    *   Para cada mês simulado no futuro, o programa calcula qual seria o **patrimônio necessário** para custear sua vida até a `expectativa de vida`.
    *   Essa meta é calculada de forma inteligente: ela considera que, mesmo aposentado, seu dinheiro continuará rendendo, mas você fará saques mensais para viver.
    *   Os saques mensais também são corrigidos pela inflação, garantindo que sua `renda desejada` mantenha o poder de compra ao longo dos anos.
    *   O saque é feito no **início** de cada mês: o primeiro ocorre na própria data da aposentadoria e é igual à "renda inicial" exibida (a renda desejada corrigida pela inflação até lá).
    *   O horizonte de vida é contado em meses exatos (`expectativa de vida` menos a idade atual, menos os meses já trabalhados), então a meta diminui suavemente mês a mês.
    *   Todas as taxas são nominais: o rendimento informado já deve ser o rendimento bruto, não o real (descontada a inflação).

Quando o seu patrimônio acumulado se iguala ou supera o patrimônio necessário (a **Meta**), a simulação para, e o programa informa que você atingiu seu objetivo. A data estimada é a data de hoje somada ao número de meses simulados (se o dia não existir no mês de destino, usa-se o último dia do mês).

### Gráfico de Evolução

O programa também exibe um gráfico de barras simples que compara a evolução do seu **Patrimônio** com a **Meta** de aposentadoria. Há um ponto a cada 12 meses e um ponto final no mês da aposentadoria (ou no último mês simulado, se a meta não for atingida), com o valor em reais ao lado de cada barra. Isso ajuda a visualizar o quão perto ou longe você está do seu objetivo.

## Limitações do cálculo

O programa também imprime este resumo logo após o resultado:

*   **Patrimônio consumido:** os saques esgotam o patrimônio exatamente na expectativa de vida. Não há perpetuidade: o principal é gasto e nada sobra no final.
*   **Sem margem de segurança:** a aposentadoria é declarada no primeiro mês em que o patrimônio iguala a meta. Se você viver mais ou o rendimento ficar abaixo do informado, o dinheiro não dura.
*   **Taxas constantes:** rendimento e inflação são fixos todos os meses, sem oscilação de mercado. O prazo é muito sensível ao rendimento real, `(1+rendimento)/(1+inflação)-1`; vale testar valores mais conservadores.
*   **Sem impostos e taxas:** não há IR sobre rendimentos nem taxa de administração.
*   **Sem outras rendas:** INSS, previdência privada e aluguéis não são considerados.
*   **Convenções de tempo:** saques no início do mês; aportes no fim do mês (o primeiro sem correção pela inflação); a idade atual é tratada como completa hoje.
*   **Tipo de taxa:** as taxas devem ser mensais e nominais (não reais). Converta taxas anuais antes de digitar.

## Testes

```bash
make test
```

Os testes conferem o patrimônio necessário contra a fórmula fechada de valor presente, o mês da aposentadoria contra uma simulação independente, os casos de borda (meta inalcançável, aposentadoria imediata, capital negativo), a leitura de números com vírgula, as validações e o cálculo de datas.
