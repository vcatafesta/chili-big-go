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

// Função para exibir o uso correto do comando com cores
func printUsage() {
	const (
		reset   = "\033[0m"    // Reset de cor
		blue    = "\033[34m"   // Azul
		green   = "\033[32m"   // Verde
	)

	fmt.Println("Uso:")
	fmt.Println("  -Ss, --search " + green + "<palavras-chave> ... <opção>" + reset)
	fmt.Println("    <opção> pode ser:")
	fmt.Println("      " + blue + "--by-name" + reset + ", " + blue + "--by-name-desc" + reset + ", " + blue + "--by-maintainer" + reset + ", " + blue + "--by-depends" + reset + ", " + blue + "--by-makedepends" + reset + ", " + blue + "--by-optdepends" + reset + ", " + blue + "--by-checkdepends" + reset)
	fmt.Println("    <palavras-chave> são os termos de busca")
	fmt.Println("  " + blue + "--json" + reset + "     Saída em formato JSON")
	fmt.Println("  " + blue + "--raw" + reset + "      Saída formatada como texto simples com todos os campos")
	fmt.Println("  " + blue + "--sep" + reset + "      Separador dos campos na saída raw (padrão é '=')")
}

// Função principal
func main() {
	args := os.Args[1:]

	if len(args) < 2 {
		printUsage()
		return
	}

	var searchField string
	var searchTerms []string
	var outputFormat string
	var separator string = "=" // Valor padrão para o separador

	optionToField := map[string]string{
		"--by-name":         "name",
		"--by-name-desc":    "name-desc",
		"--by-maintainer":   "maintainer",
		"--by-depends":      "depends",
		"--by-makedepends":  "makedepends",
		"--by-optdepends":   "optdepends",
		"--by-checkdepends": "checkdepends",
	}

	// Processa os argumentos
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json", "--raw":
			outputFormat = args[i]
			args = append(args[:i], args[i+1:]...)
			i-- // Ajusta o índice após a remoção
		case "--sep":
			if i+1 < len(args) {
				separator = args[i+1]
				args = append(args[:i], args[i+2:]...)
				i-- // Ajusta o índice após a remoção
			}
		case "--by-name", "--by-name-desc", "--by-maintainer", "--by-depends", "--by-makedepends", "--by-optdepends", "--by-checkdepends":
			if field, valid := optionToField[args[i]]; valid {
				searchField = field
				args = append(args[:i], args[i+1:]...)
				i-- // Ajusta o índice após a remoção
			}
		}
	}

	if searchField == "" {
		searchField = "name" // Valor padrão, se necessário
	}

	fmt.Println(searchField)
	return

	if len(args) < 1 {
		printUsage()
		return
	}

	searchTerms = args

	var allPackages []Package
	for _, term := range searchTerms {
		baseURL := "https://aur.archlinux.org/rpc?v=5&type=search"
		queryParams := url.Values{}
		queryParams.Add("by", searchField)
		queryParams.Add("arg", term)

		fullURL := fmt.Sprintf("%s&%s", baseURL, queryParams.Encode())

		resp, err := http.Get(fullURL)
		if err != nil {
			fmt.Printf("Erro ao fazer a requisição para o termo '%s': %s\n", term, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Erro na resposta para o termo '%s': %s\n", term, resp.Status)
			continue
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Erro ao ler a resposta para o termo '%s': %s\n", term, err)
			continue
		}

		var response map[string]interface{}
		if err := json.Unmarshal(data, &response); err != nil {
			fmt.Printf("Erro ao decodificar o JSON para o termo '%s': %s\n", term, err)
			continue
		}

		results, ok := response["results"].([]interface{})
		if !ok {
			fmt.Printf("Campo 'results' não encontrado na resposta para o termo '%s'\n", term)
			continue
		}

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

	switch outputFormat {
	case "--json":
		for _, pkg := range allPackages {
			pkgJSON, err := json.MarshalIndent(pkg, "", "  ")
			if err != nil {
				fmt.Println("Erro ao formatar o JSON:", err)
				continue
			}
			fmt.Println(string(pkgJSON))
		}
	case "--raw":
		for _, pkg := range allPackages {
			fmt.Printf("Name%s%s\nVersion%s%s\nDescription%s%s\nMaintainer%s%s\nNumVotes%s%d\nPopularity%s%.2f\nURL%s%s\n\n",
				separator, pkg.Name, separator, pkg.Version, separator, pkg.Description, separator, pkg.Maintainer,
				separator, pkg.NumVotes, separator, pkg.Popularity, separator, pkg.URL)
		}
	default:
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
