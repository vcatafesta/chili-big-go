package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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
	fmt.Println("  <opção> pode ser:")
	fmt.Println("    " + blue + "--by-name" + reset + ", " + blue + "--by-name-desc" + reset + ", " + blue + "--by-maintainer" + reset + ", " + blue + "--by-depends" + reset + ", " + blue + "--by-makedepends" + reset + ", " + blue + "--by-optdepends" + reset + ", " + blue + "--by-checkdepends" + reset)
	fmt.Println("  <palavras-chave> são os termos de busca")
	fmt.Println("  " + blue + "--json" + reset + "     Saída em formato JSON")
	fmt.Println("  " + blue + "--raw" + reset + "      Saída formatada como texto simples com todos os campos")
	fmt.Println("  " + blue + "--sep" + reset + "      Separador dos campos na saída raw (padrão é '=')")
}

// Função para buscar pacotes com cache
func searchPackage(term string, searchField string, wg *sync.WaitGroup, ch chan<- Package) {
	defer wg.Done()

	cacheKey := term + "|" + searchField
	cacheMutex.Lock()
	if entry, found := cache[cacheKey]; found && time.Since(entry.Timestamp) < cacheTTL {
		for _, pkg := range entry.Results {
			ch <- pkg
		}
		cacheMutex.Unlock()
		return
	}
	cacheMutex.Unlock()

	baseURL := "https://aur.archlinux.org/rpc?v=5&type=search"
	queryParams := url.Values{}
	queryParams.Add("by", searchField)
	queryParams.Add("arg", term)

	fullURL := fmt.Sprintf("%s&%s", baseURL, queryParams.Encode())

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
		ch <- pkg
	}

	cacheMutex.Lock()
	cache[cacheKey] = CacheEntry{Results: packages, Timestamp: time.Now()}
	cacheMutex.Unlock()
}

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		printUsage()
		return
	}

	var searchField string
	var searchTerms []string
	var outputFormat string
	var separator string = "="

	optionToField := map[string]string{
		"--by-name":         "name",
		"--by-name-desc":    "name-desc",
		"--by-maintainer":   "maintainer",
		"--by-depends":      "depends",
		"--by-makedepends":  "makedepends",
		"--by-optdepends":   "optdepends",
		"--by-checkdepends": "checkdepends",
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json", "--raw":
			outputFormat = args[i]
			args = append(args[:i], args[i+1:]...)
			i--
		case "--sep":
			if i+1 < len(args) {
				separator = args[i+1]
				args = append(args[:i], args[i+2:]...)
				i--
			}
		case "--by-name", "--by-name-desc", "--by-maintainer", "--by-depends", "--by-makedepends", "--by-optdepends", "--by-checkdepends":
			if field, valid := optionToField[args[i]]; valid {
				searchField = field
				args = append(args[:i], args[i+1:]...)
				i--
			}
		}
	}

	if searchField == "" {
		searchField = "name"
	}

	if len(args) < 1 {
		printUsage()
		return
	}

	searchTerms = args

	var wg sync.WaitGroup
	results := make(chan Package, len(searchTerms))

	for _, term := range searchTerms {
		wg.Add(1)
		go searchPackage(term, searchField, &wg, results)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allPackages []Package
	for pkg := range results {
		allPackages = append(allPackages, pkg)
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
			fmt.Printf("Name" + separator + "%s\nVersion" + separator + "%s\nDescription" + separator + "%s\nMaintainer" + separator + "%s\nNumVotes" + separator + "%d\nPopularity" + separator + "%.2f\nURL" + separator + "%s\n\n",
				pkg.Name, pkg.Version, pkg.Description, pkg.Maintainer, pkg.NumVotes, pkg.Popularity, pkg.URL)
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
