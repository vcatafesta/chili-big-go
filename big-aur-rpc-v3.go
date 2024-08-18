package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Estrutura para armazenar as informações do pacote
type Package struct {
	Name        string  `json:"Name"`
	Version     string  `json:"Version"`
	Description string  `json:"Description"`
	Maintainer  string  `json:"Maintainer"`
	NumVotes    int     `json:"NumVotes"`
	Popularity  float64 `json:"Popularity"`
	URL         string  `json:"URL"`
}

// Função para exibir o uso correto do comando
func printUsage() {
	fmt.Println("Uso:")
	fmt.Println("  -Ss, --search <campo> <palavras-chave> ...")
	fmt.Println("    <campo> pode ser:")
	fmt.Println("      name, name-desc, maintainer, depends, makedepends, optdepends, checkdepends")
	fmt.Println("    <palavras-chave> são os termos de busca")
}

// Função principal
func main() {
	// Captura todos os argumentos da linha de comando
	args := os.Args[1:]

	// Verifica se o comando principal é -Ss ou --search
	if len(args) < 3 || (args[0] != "-Ss" && args[0] != "--search") {
		printUsage()
		return
	}

	// Captura o campo de pesquisa e os termos de busca
	searchField := args[1]
	searchTerms := args[2:]

	// Verifica se o campo de pesquisa é válido
	validFields := []string{"name", "name-desc", "maintainer", "depends", "makedepends", "optdepends", "checkdepends"}
	if !contains(validFields, searchField) {
		fmt.Println("Campo de pesquisa inválido:", searchField)
		printUsage()
		return
	}

	// Cria a URL para a requisição RPC
	baseURL := "https://aur.archlinux.org/rpc?v=5&type=search"
	queryParams := url.Values{}
	queryParams.Add("by", searchField)
	queryParams.Add("arg", strings.Join(searchTerms, " "))

	// Constrói a URL completa
	fullURL := fmt.Sprintf("%s&%s", baseURL, queryParams.Encode())

	// Faz a requisição GET ao serviço RPC
	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Println("Erro ao fazer a requisição:", err)
		return
	}
	defer resp.Body.Close()

	// Verifica se o status da resposta é 200 OK
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Erro na resposta: %s\n", resp.Status)
		return
	}

	// Lê o conteúdo da resposta
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao ler a resposta:", err)
		return
	}

	// Decodifica o JSON da resposta
	var response map[string]interface{}
	if err := json.Unmarshal(data, &response); err != nil {
		fmt.Println("Erro ao decodificar o JSON:", err)
		return
	}

	// Verifica se o campo 'results' está presente na resposta
	results, ok := response["results"].([]interface{})
	if !ok {
		fmt.Println("Campo 'results' não encontrado na resposta")
		return
	}

	// Converte o resultado em um slice de pacotes
	var packages []Package
	for _, result := range results {
		pkgData, err := json.Marshal(result)
		if err != nil {
			fmt.Println("Erro ao converter o resultado:", err)
			continue
		}

		var pkg Package
		if err := json.Unmarshal(pkgData, &pkg); err != nil {
			fmt.Println("Erro ao decodificar o pacote:", err)
			continue
		}

		packages = append(packages, pkg)
	}

	// Exibe a saída JSON formatada com um pacote por linha
	for _, pkg := range packages {
		pkgJSON, err := json.MarshalIndent(pkg, "", "  ")
		if err != nil {
			fmt.Println("Erro ao formatar o JSON:", err)
			continue
		}
		fmt.Println(string(pkgJSON))
	}
}

// Função auxiliar para verificar se um valor está em um slice
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
