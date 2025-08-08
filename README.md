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

    **Campos solicitados:**

    *   `Idade atual (anos)`: Sua idade hoje.
    *   `Capital inicial disponível hoje em reais`: Quanto dinheiro você já tem investido.
    *   `Inflação mensal em %`: A taxa média de aumento dos preços (ex: 0.3 para 0.3%).
    *   `Rendimento mensal em %`: O crescimento médio dos seus investimentos (ex: 0.6 para 0.6%).
    *   `Aporte mensal em reais`: O valor que você investe todo mês.
    *   `Renda mensal desejada na aposentadoria`: Quanto você gostaria de receber por mês ao se aposentar (em valores de hoje).
    *   `Expectativa de vida (anos)`: Até que idade você espera viver.

3.  **Veja o resultado:**

    Após preencher os dados, o programa exibirá:

    *   O tempo restante para a aposentadoria (em meses e anos).
    *   A data estimada para a aposentadoria.
    *   Sua idade ao se aposentar.
    *   A renda inicial na aposentadoria (corrigida pela inflação).

## Explicação Detalhada

O programa funciona como um simulador financeiro mês a mês, projetando o futuro do seu patrimônio com base em duas lógicas principais:

1.  **Fase de Acumulação (enquanto você trabalha):**

    *   A cada mês, seu patrimônio cresce com base na `taxa de rendimento mensal`.
    *   Você adiciona o `aporte mensal`, que também é corrigido pela `inflação` para manter seu poder de compra.
    *   O programa continua simulando até que seu patrimônio atinja o valor necessário para a aposentadoria.

2.  **Cálculo da Meta de Aposentadoria:**

    *   Para cada mês simulado no futuro, o programa calcula qual seria o **patrimônio necessário** para custear sua vida até a `expectativa de vida`.
    *   Essa meta é calculada de forma inteligente: ela considera que, mesmo aposentado, seu dinheiro continuará rendendo, mas você fará saques mensais para viver.
    *   Os saques mensais também são corrigidos pela inflação, garantindo que sua `renda desejada` mantenha o poder de compra ao longo dos anos.

Quando o seu patrimônio acumulado (`gráfico de Patrimônio`) se encontra com o patrimônio necessário (`gráfico de Meta`), a simulação para, e o programa informa que você atingiu seu objetivo.

### Gráfico de Evolução

O programa também exibe um gráfico de barras simples que compara a evolução do seu **Patrimônio** com a **Meta** de aposentadoria ao longo do tempo. Isso ajuda a visualizar o quão perto ou longe você está do seu objetivo.
