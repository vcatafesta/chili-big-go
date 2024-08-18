package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Defina a estrutura para um item no array JSON, ajustando conforme necessário
type Package struct {
	ID             int         `json:"ID"`
	Name           string      `json:"Name"`
	PackageBaseID  int         `json:"PackageBaseID"`
	PackageBase    string      `json:"PackageBase"`
	Version        string      `json:"Version"`
	Description    string      `json:"Description"`
	URL            string      `json:"URL"`
	NumVotes       int         `json:"NumVotes"`
	Popularity     float64     `json:"Popularity"`
	OutOfDate      interface{} `json:"OutOfDate"` // Usar interface{} para lidar com múltiplos tipos
	Maintainer     string      `json:"Maintainer"`
	Submitter      string      `json:"Submitter"`
	FirstSubmitted int64       `json:"FirstSubmitted"`
	LastModified   int64       `json:"LastModified"`
	URLPath        string      `json:"URLPath"`
}

func printUsage() {
	fmt.Println("Uso:")
	fmt.Println("  -Ss, --search <nome do pacote> ...    Nome(s) do pacote(s) para buscar")
}

func main() {
	// Captura todos os argumentos da linha de comando
	args := os.Args[1:]

	// Verifica se o comando principal é -Ss ou --search
	if len(args) < 1 || (args[0] != "-Ss" && args[0] != "--search") {
		printUsage()
		return
	}

	// Captura todos os termos de busca após o comando principal
	searchTerms := args[1:]

	// Verifica se termos de busca foram fornecidos
	if len(searchTerms) == 0 {
		printUsage()
		return
	}

	// // Exibe o número de termos de busca e os próprios termos para depuração
	/*
	fmt.Printf("Número de termos de busca: %d\n", len(searchTerms))
	fmt.Println("Termos de busca:", searchTerms)
	*/

	// URL do arquivo JSON
	// url := "https://chililinux.com/packages-meta-v1.json.gz"
	url := "https://aur.archlinux.org/packages-meta-v1.json.gz"

	// Faz o download do arquivo JSON
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Erro ao fazer a requisição:", err)
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

	// Filtra os pacotes com base nos termos de busca
	var output []Package
	for _, pkg := range packages {
		for _, term := range searchTerms {
			if strings.Contains(strings.ToLower(pkg.Name), strings.ToLower(term)) ||
				strings.Contains(strings.ToLower(pkg.Description), strings.ToLower(term)) {
				output = append(output, pkg)
				break
			}
		}
	}

	// Começa o array JSON
	fmt.Print("[\n")

	// Exibe a saída JSON formatada com um pacote por linha
	for i, pkg := range output {
		if i > 0 {
			fmt.Print(",\n")
		}
		pkgJSON, err := json.Marshal(pkg)
		if err != nil {
			fmt.Println("Erro ao formatar o JSON:", err)
			continue
		}
		fmt.Print(string(pkgJSON))
	}

	// Fecha o array JSON
	fmt.Print("\n]\n")
}
