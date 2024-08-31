package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

type Package struct {
	Name        string  `json:"Name"`
	Version     string  `json:"Version"`
	Description string  `json:"Description"`
	Maintainer  string  `json:"Maintainer"`
	NumVotes    int     `json:"NumVotes"`
	Popularity  float64 `json:"Popularity"`
	URL         string  `json:"URL"`
}

type CacheEntry struct {
	Results   []Package
	Timestamp time.Time
}

var (
	cache      = make(map[string]CacheEntry)
	cacheMutex = sync.Mutex{}
	cacheTTL   = time.Minute * 5
)

func printUsage() {
	const (
		reset   = "\033[0m"
		blue    = "\033[34m"
		green   = "\033[32m"
	)

	fmt.Println("Uso:")
	fmt.Println("  -Ss, --search " + green + "<palavras-chave> ... <opção>" + reset)
	fmt.Println("    <opção> pode ser:")
	fmt.Println("      " + blue + "--by-name" + reset + ", " + blue + "--by-name-desc" + reset + ", " + blue + "--by-maintainer" + reset + ", " + blue + "--by-depends" + reset + ", " + blue + "--by-makedepends" + reset + ", " + blue + "--by-optdepends" + reset + ", " + blue + "--by-checkdepends" + reset)
	fmt.Println("    <palavras-chave> são os termos de busca")
	fmt.Println("  " + blue + "--json" + reset + "     Saída em formato JSON")
	fmt.Println("  " + blue + "--raw" + reset + "      Saída formatada como texto simples com todos os campos")
	fmt.Println("  " + blue + "--sep" + reset + "      Separador dos campos na saída raw (padrão é '=')")
	fmt.Println("  " + blue + "--limit" + reset + "    Limite de pacotes encontrados")
	fmt.Println("  -Si " + green + "<pacote>" + reset + "     Consulta informações específicas do pacote")
}

func searchPackage(term string, searchField string, limit int, wg *sync.WaitGroup, ch chan<- Package) {
	defer wg.Done()

	cacheKey := term + "|" + searchField
	cacheMutex.Lock()
	if entry, found := cache[cacheKey]; found && time.Since(entry.Timestamp) < cacheTTL {
		count := 0
		for _, pkg := range entry.Results {
			if limit > 0 && count >= limit {
				break
			}
			ch <- pkg
			count++
		}
		cacheMutex.Unlock()
		return
	}
	cacheMutex.Unlock()

	baseURL := "https://aur.archlinux.org/rpc"
	queryParams := url.Values{}
	queryParams.Add("v", "5")
	queryParams.Add("type", "search")
	queryParams.Add("by", searchField)
	queryParams.Add("arg", term)

	// Construir a URL na ordem correta
	fullURL := fmt.Sprintf("%s?v=%s&type=%s&by=%s&arg=%s", baseURL, queryParams.Get("v"), queryParams.Get("type"), queryParams.Get("by"), queryParams.Get("arg"))

	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Printf("Erro ao fazer a requisição para o termo '%s': %s\n", term, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Erro na resposta para o termo '%s': %s\n", term, resp.Status)
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Erro ao ler a resposta para o termo '%s': %s\n", term, err)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(data, &response); err != nil {
		fmt.Printf("Erro ao decodificar o JSON para o termo '%s': %s\n", term, err)
		return
	}

	results, ok := response["results"].([]interface{})
	if !ok {
		fmt.Printf("Campo 'results' não encontrado na resposta para o termo '%s'\n", term)
		return
	}

	var packages []Package
	for _, result := range results {
		pkgData, err := json.Marshal(result)
		if err != nil {
			fmt.Printf("Erro ao converter o resultado para o termo '%s': %s\n", term, err)
			return
		}

		var pkg Package
		if err := json.Unmarshal(pkgData, &pkg); err != nil {
			fmt.Printf("Erro ao decodificar o pacote para o termo '%s': %s\n", term, err)
			return
		}

		packages = append(packages, pkg)
		if limit > 0 && len(packages) >= limit {
			break
		}
		ch <- pkg
	}

	cacheMutex.Lock()
	cache[cacheKey] = CacheEntry{Results: packages, Timestamp: time.Now()}
	cacheMutex.Unlock()
}

func infoPackage(pkgName string, wg *sync.WaitGroup, ch chan<- Package) {
	defer wg.Done()

	cacheKey := "info|" + pkgName
	cacheMutex.Lock()
	if entry, found := cache[cacheKey]; found && time.Since(entry.Timestamp) < cacheTTL {
		for _, pkg := range entry.Results {
			ch <- pkg
		}
		cacheMutex.Unlock()
		return
	}
	cacheMutex.Unlock()

	baseURL := "https://aur.archlinux.org/rpc"
	queryParams := url.Values{}
	queryParams.Add("v", "5")
	queryParams.Add("type", "info")
	queryParams.Add("arg[]", pkgName)

	// Construir a URL na ordem correta para o tipo info
	fullURL := fmt.Sprintf("%s?v=%s&type=%s&arg[]=%s", baseURL, queryParams.Get("v"), queryParams.Get("type"), queryParams.Get("arg[]"))

	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Printf("Erro ao fazer a requisição para o pacote '%s': %s\n", pkgName, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Erro na resposta para o pacote '%s': %s\n", pkgName, resp.Status)
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Erro ao ler a resposta para o pacote '%s': %s\n", pkgName, err)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(data, &response); err != nil {
		fmt.Printf("Erro ao decodificar o JSON para o pacote '%s': %s\n", pkgName, err)
		return
	}

	results, ok := response["results"].([]interface{})
	if !ok {
		fmt.Printf("Campo 'results' não encontrado na resposta para o pacote '%s'\n", pkgName)
		return
	}

	var packages []Package
	for _, result := range results {
		pkgData, err := json.Marshal(result)
		if err != nil {
			fmt.Printf("Erro ao converter o resultado para o pacote '%s': %s\n", pkgName, err)
			return
		}

		var pkg Package
		if err := json.Unmarshal(pkgData, &pkg); err != nil {
			fmt.Printf("Erro ao decodificar o pacote '%s': %s\n", pkgName, err)
			return
		}

		packages = append(packages, pkg)
		ch <- pkg
	}

	cacheMutex.Lock()
	cache[cacheKey] = CacheEntry{Results: packages, Timestamp: time.Now()}
	cacheMutex.Unlock()
}

func main() {
	args := os.Args[1:]

	if len(args) < 2 {
		printUsage()
		return
	}

	var searchField string
	var searchTerms []string
	var outputFormat string
	var separator string = "="
	var limit int = -1  // Usar -1 para indicar que não há limite
	var searchMode string

	optionToField := map[string]string{
		"--by-name":         "name",
		"--by-name-desc":    "name-desc",
		"--by-maintainer":   "maintainer",
		"--by-depends":      "depends",
		"--by-makedepends":  "makedepends",
		"--by-optdepends":   "optdepends",
		"--by-checkdepends": "checkdepends",
	}

	// Verificar se o parâmetro -Ss ou -Si está presente
	if args[0] == "-Ss" {
		searchMode = "search"
		args = args[1:] // Remover o -Ss da lista de argumentos
	} else if args[0] == "-Si" {
		searchMode = "info"
		args = args[1:] // Remover o -Si da lista de argumentos
	} else {
		printUsage()
		return
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json", "--raw":
			outputFormat = args[i]
		case "--sep":
			if i+1 < len(args) {
				separator = args[i+1]
				i++
			} else {
				fmt.Println("Erro: --sep requer um argumento")
				return
			}
		case "--limit":
			if i+1 < len(args) {
				parsedLimit, err := strconv.Atoi(args[i+1])
				if err != nil || parsedLimit < 1 {
					fmt.Println("Erro: --limit requer um número positivo")
					return
				}
				limit = parsedLimit
				i++
			} else {
				fmt.Println("Erro: --limit requer um argumento")
				return
			}
		default:
			if field, ok := optionToField[args[i]]; ok {
				searchField = field
			} else {
				searchTerms = append(searchTerms, args[i])
			}
		}
	}

	if searchMode == "search" && len(searchTerms) == 0 {
		fmt.Println("Erro: Nenhuma palavra-chave de busca fornecida")
		return
	}

	ch := make(chan Package)
	var wg sync.WaitGroup

	if searchMode == "search" {
		for _, term := range searchTerms {
			wg.Add(1)
			go searchPackage(term, searchField, limit, &wg, ch)
		}
	} else if searchMode == "info" {
		for _, pkgName := range searchTerms {
			wg.Add(1)
			go infoPackage(pkgName, &wg, ch)
		}
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var results []Package
	for pkg := range ch {
		results = append(results, pkg)
	}

	if outputFormat == "--json" {
		jsonData, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			fmt.Println("Erro ao formatar saída JSON:", err)
			return
		}
		fmt.Println(string(jsonData))
	} else {
		for _, pkg := range results {
			fmt.Printf("Name%s%s\nVersion%s%s\nDescription%s%s\nMaintainer%s%s\nNumVotes%s%d\nPopularity%s%.2f\nURL%s%s\n\n",
				separator, pkg.Name, separator, pkg.Version, separator, pkg.Description,
				separator, pkg.Maintainer, separator, pkg.NumVotes, separator, pkg.Popularity,
				separator, pkg.URL)
		}
	}
}
