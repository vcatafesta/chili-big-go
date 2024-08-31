package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Defina a estrutura para um item no array JSON, ajuste conforme necessário
type Package struct {
	// Adicione os campos necessários aqui conforme a estrutura do JSON
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	// Outros campos
}

func main() {
	// URL do arquivo JSON
	url := "https://aur.archlinux.org/packages-meta-v1.json.gz"

	// Faz o download do arquivo JSON
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Erro ao fazer o download do arquivo:", err)
		return
	}
	defer resp.Body.Close()

	// Lê o conteúdo do arquivo JSON
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao ler o arquivo:", err)
		return
	}

	// Decodifica o JSON
	var packages []Package
	if err := json.Unmarshal(data, &packages); err != nil {
		fmt.Println("Erro ao decodificar o JSON:", err)
		return
	}

	// Exemplo: imprimir os pacotes decodificados
	for _, pkg := range packages {
		fmt.Printf("Nome: %s, Versão: %s, Descrição: %s\n", pkg.Name, pkg.Version, pkg.Description)
	}

	// Processar os pacotes conforme necessário
	processPackages(packages)
}

// Função para processar o JSON
func processPackages(packs []Package) {
	// Adicione o código para processar os pacotes aqui
	// Exemplo: Iterar sobre os pacotes
	for _, pkg := range packs {
		fmt.Printf("Processando pacote: Nome: %s, Versão: %s\n", pkg.Name, pkg.Version)
	}
}
