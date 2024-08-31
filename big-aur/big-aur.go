package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Estrutura para representar um resultado do AUR
type AurPackage struct {
	Name        string `json:"Name"`
	Version     string `json:"Version"`
	Description string `json:"Description"`
	URL         string `json:"URL"`
}

// Estrutura para a resposta JSON
type AurResponse struct {
	ResultCount int          `json:"resultcount"`
	Results     []AurPackage `json:"results"`
}

func main() {
	url := "https://aur.archlinux.org/rpc?arg=elisa&by=name-desc&type=search&v=5"

	// Faz a requisição HTTP GET
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Erro ao fazer requisição HTTP: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Lê o corpo da resposta
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Erro ao ler corpo da resposta: %v\n", err)
		return
	}

	// Decodifica o JSON
	var aurResponse AurResponse
	err = json.Unmarshal(body, &aurResponse)
	if err != nil {
		fmt.Printf("Erro ao decodificar JSON: %v\n", err)
		return
	}

	// Exibe os resultados
	fmt.Printf("Total de pacotes encontrados: %d\n\n", aurResponse.ResultCount)
	for _, pkg := range aurResponse.Results {
		fmt.Printf("Nome: %s\nVersão: %s\nDescrição: %s\nURL: %s\n\n",
			pkg.Name, pkg.Version, pkg.Description, pkg.URL)
	}
}
