/*
  big-search-aur - Command-line AUR helper
    go get github.com/go-ini/ini
    Chili GNU/Linux - https://github.com/vcatafesta/chili/go
    Chili GNU/Linux - https://chililinux.com
    Chili GNU/Linux - https://chilios.com.br

  Created: 2024/08/13
  Altered: 2024/08/13

  Copyright (c) 2024-2024, Vilmar Catafesta <vcatafesta@gmail.com>
  All rights reserved.

  Redistribution and use in source and binary forms, with or without
  modification, are permitted provided that the following conditions
  are met:
  1. Redistributions of source code must retain the above copyright
    notice, this list of conditions and the following disclaimer.
  2. Redistributions in binary form must reproduce the above copyright
    notice, this list of conditions and the following disclaimer in the
    documentation and/or other materials provided with the distribution.

  THIS SOFTWARE IS PROVIDED BY THE AUTHOR ``AS IS'' AND ANY EXPRESS OR
  IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
  OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
  IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT,
  INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
  NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
  DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
  THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
  (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
  THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

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
  reset = "\033[0m"
  blue  = "\033[34m"
  green = "\033[32m"
  cyan  = "\033[36m"
  red   = "\033[31m"
)

	fmt.Println("Uso:")
	fmt.Printf("%s%-20s %s%s%s%s%s\n", blue, "  --Ss, --search", green, "<palavra-chave> ... <opção>", cyan, " # pesquisa no repositório AUR por palavras coincidentes", reset)
	fmt.Printf("%s%-20s %s%s%s%s%s\n", blue, "  --Si, --info", green, "<palavra-chave> ... <opção>", cyan, " # pesquisa no repositório AUR por palavras coincidentes", reset)
	fmt.Println("    <palavras-chave> são os termos/pacotes de busca")
	fmt.Println("    <opção> podem ser:")
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --by-name", reset, "Pesquisa pelo nome do pacote apenas (padrão)", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --by-name-desc", reset, "Pesquisa pelo nome e descrição do pacote", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --by-maintainer", reset, "Pesquisa pelo mantenedor do pacote", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --by-dependsr", reset, "Pesquisa pacotes que são dependências por palavras-chaves", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --by-makedependsr", reset, "Pesquisa pacotes que são dependências para compilação por palavras-chaves", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --by-optdependsr", reset, "Pesquisa pacotes que são dependências opcionais por palavras-chaves", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --by-checkdependsr", reset, "Pesquisa pacotes que são dependências para verificação por palavras-chaves", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --json", reset, "Saída em formato JSON (padrão)", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --raw", reset, "Saída formatada como texto simples com todos os campos", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --sep", reset, "Separador dos campos na saída raw (padrão é '=')", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --limit", reset, "Limite de pacotes encontrados", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --verbose", reset, "Liga modo verboso", reset)
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
	var limit int = -1 // Usar -1 para indicar que não há limite
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

	// Verificar se o parâmetro -Ss, -Si, --search ou --info está presente
	if len(args) > 0 {
		switch args[0] {
		case "-Ss", "--search":
			searchMode = "search"
			args = args[1:] // Remove o -Ss ou --search da lista de argumentos
		case "-Si", "--info":
			searchMode = "info"
			args = args[1:] // Remove o -Si ou --info da lista de argumentos
		default:
			printUsage()
			return
		}
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

	if outputFormat == "" {
		outputFormat = "--json" // Define o formato padrão como json
	}

	if outputFormat == "--json" {
		//		jsonData, err := json.MarshalIndent(results, "", "  ")
		jsonData, err := json.Marshal(results) // compacto, like jq -c
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

func searchPackage(term string, searchField string, limit int, wg *sync.WaitGroup, ch chan<- Package) {
	defer wg.Done()

	cacheKey := term + "|" + searchField
	cacheMutex.Lock()
	if entry, found := cache[cacheKey]; found && time.Since(entry.Timestamp) < cacheTTL {
		count := 0
		for _, pkg := range entry.Results {
			// Envia todos os pacotes se limit for <= 0
			if limit <= 0 || count < limit {
				ch <- pkg
				count++
			} else {
				break
			}
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

	count := 0
	for _, result := range results {
		// Verifica se o limite é maior que zero e se o contador é menor que o limite
		if limit > 0 && count >= limit {
			break
		}

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

		ch <- pkg
		count++
	}

	cacheMutex.Lock()
	cache[cacheKey] = CacheEntry{Results: nil, Timestamp: time.Now()} // Cache vazio pois já foi processado
	cacheMutex.Unlock()
}
