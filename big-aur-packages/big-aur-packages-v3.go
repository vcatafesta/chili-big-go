package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Defina a estrutura para um item no array JSON, ajuste conforme necessário
type Package struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	// Outros campos, ajuste conforme a estrutura do JSON
}

func main() {
	// URL do arquivo JSON
//	url := "https://aur.archlinux.org/packages-meta-v1.json.gz"
	url := "https://chililinux.com/packages-meta-v1.json.gz"

	// Faz a requisição para obter o arquivo JSON
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Erro ao fazer a requisição:", err)
		return
	}
	defer resp.Body.Close()

	// Lê o conteúdo do arquivo JSON
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao ler o conteúdo:", err)
		return
	}

	// Decodifica o JSON
	var packages []Package
	if err := json.Unmarshal(data, &packages); err != nil {
		fmt.Println("Erro ao decodificar o JSON:", err)
		return
	}

	// Determina o número de pacotes a ser processado
	numToProcess := 10
	if len(packages) < numToProcess {
		numToProcess = len(packages)
	}

	// Contagem de pacotes processados
	fmt.Printf("Número de pacotes processados: %d\n", numToProcess)

	// Processa e imprime os primeiros pacotes
	processPackages(packages[:numToProcess])
}

// Função para processar o JSON
func processPackages(packs []Package) {
	// Adicione o código para processar os pacotes aqui
	// Contador de pacotes processados
	for i, pkg := range packs {
		fmt.Printf("Pacote %d: Nome: %s, Versão: %s\n", i+1, pkg.Name, pkg.Version)
	}
}
