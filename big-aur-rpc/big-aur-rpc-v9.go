package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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
	fmt.Println("  -Ss, --search <palavras-chave> ... <opção>")
	fmt.Println("    <opção> pode ser:")
	fmt.Println("      --by-name, --by-name-desc, --by-maintainer, --by-depends, --by-makedepends, --by-optdepends, --by-checkdepends")
	fmt.Println("    <palavras-chave> são os termos de busca")
	fmt.Println("  --json     Saída em formato JSON")
	fmt.Println("  --raw      Saída formatada como texto simples com todos os campos")
}

// Função principal
func main() {
	// Captura todos os argumentos da linha de comando
	args := os.Args[1:]

	// Verifica se o comando principal é -Ss ou --search
	if len(args) < 2 || (args[0] != "-Ss" && args[0] != "--search") {
		printUsage()
		return
	}

	// Mapeia as opções para os campos correspondentes
	optionToField := map[string]string{
		"--by-name":         "name",
		"--by-name-desc":    "name-desc",
		"--by-maintainer":   "maintainer",
		"--by-depends":      "depends",
		"--by-makedepends":  "makedepends",
		"--by-optdepends":   "optdepends",
		"--by-checkdepends": "checkdepends",
		"-by-name":         "name",
		"-by-name-desc":    "name-desc",
		"-by-maintainer":   "maintainer",
		"-by-depends":      "depends",
		"-by-makedepends":  "makedepends",
		"-by-optdepends":   "optdepends",
		"-by-checkdepends": "checkdepends",
		"--name":            "name",
		"--name-desc":       "name-desc",
		"--maintainer":      "maintainer",
		"--depends":         "depends",
		"--makedepends":     "makedepends",
		"--optdepends":      "optdepends",
		"--checkdepends":    "checkdepends",
		"-name":            "name",
		"-name-desc":       "name-desc",
		"-maintainer":      "maintainer",
		"-depends":         "depends",
		"-makedepends":     "makedepends",
		"-optdepends":      "optdepends",
		"-checkdepends":    "checkdepends",
	}

	// Define a opção padrão e o tipo de saída
	defaultOption := "--by-name"
	var outputFormat string

	// Localiza a opção de saída e remove os argumentos correspondentes
	for i := len(args) - 1; i >= 1; i-- {
		switch args[i] {
		case "--json", "--raw":
			outputFormat = args[i]
			args = append(args[:i], args[i+1:]...) // Remove a opção de saída
		}
	}

	// Localiza a opção de pesquisa e separa as palavras-chave
	var searchField string
	var searchTerms []string

	for i := len(args) - 1; i >= 1; i-- { // Inverte a ordem de verificação
		if field, valid := optionToField[args[i]]; valid {
			searchField = field
			searchTerms = args[1:i] // Palavras-chave são todos os argumentos antes da opção
			break
		}
	}

	// Se nenhuma opção válida foi fornecida, usa a opção padrão
	if searchField == "" {
		searchField = optionToField[defaultOption]
		searchTerms = args[1:] // Se não houver opção, todos os argumentos são palavras-chave
	}

	// Processa cada termo de busca individualmente
	var allPackages []Package
	for _, term := range searchTerms {
		// Cria a URL para a requisição RPC
		baseURL := "https://aur.archlinux.org/rpc?v=5&type=search"
		queryParams := url.Values{}
		queryParams.Add("by", searchField)
		queryParams.Add("arg", term)

		// Constrói a URL completa
		fullURL := fmt.Sprintf("%s&%s", baseURL, queryParams.Encode())

		// Faz a requisição GET ao serviço RPC
		resp, err := http.Get(fullURL)
		if err != nil {
			fmt.Printf("Erro ao fazer a requisição para o termo '%s': %s\n", term, err)
			continue
		}
		defer resp.Body.Close()

		// Verifica se o status da resposta é 200 OK
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Erro na resposta para o termo '%s': %s\n", term, resp.Status)
			continue
		}

		// Lê o conteúdo da resposta
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Erro ao ler a resposta para o termo '%s': %s\n", term, err)
			continue
		}

		// Decodifica o JSON da resposta
		var response map[string]interface{}
		if err := json.Unmarshal(data, &response); err != nil {
			fmt.Printf("Erro ao decodificar o JSON para o termo '%s': %s\n", term, err)
			continue
		}

		// Verifica se o campo 'results' está presente na resposta
		results, ok := response["results"].([]interface{})
		if !ok {
			fmt.Printf("Campo 'results' não encontrado na resposta para o termo '%s'\n", term)
			continue
		}

		// Converte o resultado em um slice de pacotes
		for _, result := range results {
			pkgData, err := json.Marshal(result)
			if err != nil {
				fmt.Printf("Erro ao converter o resultado para o termo '%s': %s\n", term, err)
				continue
			}

			var pkg Package
			if err := json.Unmarshal(pkgData, &pkg); err != nil {
				fmt.Printf("Erro ao decodificar o pacote para o termo '%s': %s\n", term, err)
				continue
			}

			allPackages = append(allPackages, pkg)
		}
	}

	// Exibe a saída conforme o formato especificado
	switch outputFormat {
	case "--json":
		// Exibe a saída JSON formatada com um pacote por linha
		for _, pkg := range allPackages {
			pkgJSON, err := json.MarshalIndent(pkg, "", "  ")
			if err != nil {
				fmt.Println("Erro ao formatar o JSON:", err)
				continue
			}
			fmt.Println(string(pkgJSON))
		}
	case "--raw":
		// Exibe a saída formatada como texto simples com todos os campos
		for _, pkg := range allPackages {
			fmt.Printf("Name: %s\nVersion: %s\nDescription: %s\nMaintainer: %s\nNumVotes: %d\nPopularity: %.2f\nURL: %s\n\n",
				pkg.Name, pkg.Version, pkg.Description, pkg.Maintainer, pkg.NumVotes, pkg.Popularity, pkg.URL)
		}
	default:
		// Se nenhum formato de saída especificado, exibe no formato JSON por padrão
		for _, pkg := range allPackages {
			pkgJSON, err := json.MarshalIndent(pkg, "", "  ")
			if err != nil {
				fmt.Println("Erro ao formatar o JSON:", err)
				continue
			}
			fmt.Println(string(pkgJSON))
		}
	}
}
