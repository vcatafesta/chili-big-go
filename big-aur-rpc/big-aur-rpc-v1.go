package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type RPCResponse struct {
	Version     int      `json:"version"`
	Type        string   `json:"type"`
	ResultCount int      `json:"resultcount"`
	Results     []Package `json:"results"`
}

type Package struct {
	Name        string `json:"Name"`
	Version     string `json:"Version"`
	Description string `json:"Description"`
	URL         string `json:"URL"`
	Maintainer  string `json:"Maintainer"`
}

func main() {
	// Captura o termo de pesquisa a partir dos argumentos da linha de comando
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go <termo_de_busca>")
		return
	}
	searchTerm := os.Args[1]

	// Constrói a URL de pesquisa RPC
	url := fmt.Sprintf("https://aur.archlinux.org/rpc/?v=5&type=search&arg=%s", searchTerm)

	// Faz a requisição HTTP
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Erro ao fazer a requisição:", err)
		return
	}
	defer resp.Body.Close()

	// Decodifica a resposta JSON
	var rpcResponse RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
		fmt.Println("Erro ao decodificar o JSON:", err)
		return
	}

	// Exibe os resultados
	for _, pkg := range rpcResponse.Results {
		fmt.Printf("Pacote: %s\nVersão: %s\nDescrição: %s\nURL: %s\nMantenedor: %s\n\n",
			pkg.Name, pkg.Version, pkg.Description, pkg.URL, pkg.Maintainer)
	}
}
