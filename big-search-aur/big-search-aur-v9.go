/*
  big-search-aur - Command-line AUR helper
    go get github.com/go-ini/ini
    Chili GNU/Linux - https://github.com/vcatafesta/chili/go
    Chili GNU/Linux - https://chililinux.com
    Chili GNU/Linux - https://chilios.com.br

  Created: 2024/08/13
  Altered: 2024/08/15

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
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	_APP_       = "big-search-aur"
	_PKGDESC_   = "Command-line AUR helper"
	_VERSION_   = "0.15.0-20240815"
	_COPYRIGHT_ = "Copyright (C) 2024 Vilmar Catafesta, <vcatafesta@gmail.com>"
)

// Constantes para cores ANSI
const (
	reset   = "\x1b[0m"
	red     = "\x1b[31m"
	green   = "\x1b[32m"
	yellow  = "\x1b[33m"
	blue    = "\x1b[34m"
	magenta = "\x1b[35m"
	cyan    = "\x1b[36m"
	white   = "\x1b[37m"
	Reset   = "\x1b[0m"
	Red     = "\x1b[31m"
	Green   = "\x1b[32m"
	Yellow  = "\x1b[33m"
	Blue    = "\x1b[34m"
	Magenta = "\x1b[35m"
	Cyan    = "\x1b[36m"
	White   = "\x1b[37m"
)

type Package struct {
	Name        string  `json:"Name"`
	Version     string  `json:"Version"`
	Description string  `json:"Description"`
	Maintainer  string  `json:"Maintainer"`
//	NumVotes    int     `json:"NumVotes"`
	NumVotes    string  `json:"NumVotes"`
//	Popularity  float64 `json:"Popularity"`
	Popularity  string `json:"Popularity"`
	URL         string  `json:"URL"`
	fullURL     string
	count       int
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

const baseURL = "https://aur.archlinux.org/rpc"

// Declaração da variável global
var verbose bool
var fullURL string
var count int

func main() {
	args := os.Args[1:]

	if len(args) < 2 {
		printUsage()
		return
	}

	var searchField string
	var searchTerms []string
	var outputFormat string
	var separator string = "|"
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
		case "--verbose":
			verbose = true
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
			go infoPackage(pkgName, limit, &wg, ch)
		}
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var results []Package
	for pkg := range ch {
		if verbose {
			log.Printf("%s %sGET:%s %02d '%s'%s em %s %s- 200 OK%s\n", _APP_, Green, Yellow, pkg.count, strings.TrimSpace(pkg.Name), Reset, pkg.fullURL, Green, Reset)
		}
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
//		// Itera sobre o array de structs `Package` e formata a saída
//		for _, pkg := range results {
//			fmt.Printf("Name%s%s\nVersion%s%s\nDescription%s%s\nMaintainer%s%s\nNumVotes%s%d\nPopularity%s%.2f\nURL%s%s\n\n",
//				separator, pkg.Name, separator, pkg.Version, separator, pkg.Description,
//				separator, pkg.Maintainer, separator, pkg.NumVotes, separator, pkg.Popularity,
//				separator, pkg.URL)
//		}

		// Itera sobre o array de structs `Package` e formata a saída
		for _, pkg := range results {
//			fmt.Printf("%s %s %s %s %s %s %s %s %d %s %.2f %s %s %s %s\n",
//				pkg.Name, separator, pkg.Version, separator, pkg.Description,
//				separator, pkg.Maintainer, separator, pkg.NumVotes, separator, pkg.Popularity,
//				separator, pkg.URL, pkg.count)

			fmt.Println(pkg.Name + separator + pkg.Version + separator + pkg.Description + separator + pkg.Maintainer + separator, pkg.NumVotes, separator, pkg.Popularity,
				separator, pkg.URL, separator, pkg.count)
		}

		//		// Cria uma string para armazenar toda a saída
		//		var output string
		//		// Itera sobre o array de structs `Package` e formata a saída
		//		for _, pkg := range results {
		//			output += fmt.Sprintf("%s%s%s%s%s%s%d%s%.2f%s%s%s%d\n",
		//				pkg.Name,
		//				separator,
		//				pkg.Version,
		//				separator,
		//				pkg.Description,
		//				separator,
		//				pkg.Maintainer,
		//				separator,
		//				pkg.NumVotes,
		//				separator,
		//				pkg.Popularity,
		//				separator,
		//				pkg.URL,
		//				separator,
		//				pkg.fullURL,
		//				separator,
		//				pkg.count)
		//		}
		//
		//		// Imprime toda a saída como uma única string
		//		fmt.Print(output)
	}
}

func printUsage() {
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

func infoPackage(pkgName string, limit int, wg *sync.WaitGroup, ch chan<- Package) {
	defer wg.Done()

	count = 0
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

	queryParams := url.Values{}
	queryParams.Add("v", "5")
	queryParams.Add("type", "info")
	queryParams.Add("arg[]", pkgName)

	// Construir a URL na ordem correta para o tipo info
	fullURL = fmt.Sprintf("%s?v=%s&type=%s&arg[]=%s", baseURL, queryParams.Get("v"), queryParams.Get("type"), queryParams.Get("arg[]"))

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
//	if err := json.Unmarshal(data, &response); err != nil {
	if err := json.Unmarshal([]byte(data), &response); err != nil {
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
//		if err := json.Unmarshal(pkgData, &pkg); err != nil {
		if err := json.Unmarshal([]byte(pkgData), &pkg); err != nil {
			fmt.Printf("Erro ao decodificar o pacote '%s': %s\n", pkgName, err)
			return
		}

		count++
		if verbose {
			// Preenche o campo `url` com `fullURL`
			fullURL = fmt.Sprintf("%s?v=%s&type=%s&by=%s&arg=%s", baseURL, queryParams.Get("v"), queryParams.Get("type"), queryParams.Get("by"), pkg.Name)
			pkg.fullURL = fullURL
			pkg.count = count
		}

		packages = append(packages, pkg)
		// Envia todos os pacotes se limit for <= 0
		if limit <= 0 || count < limit {
			ch <- pkg
		} else {
			break
		}
	}

	cacheMutex.Lock()
	cache[cacheKey] = CacheEntry{Results: packages, Timestamp: time.Now()}
	cacheMutex.Unlock()
}

func searchPackage(term string, searchField string, limit int, wg *sync.WaitGroup, ch chan<- Package) {
	defer wg.Done()

	cacheKey := term + "|" + searchField
	cacheMutex.Lock()
	if entry, found := cache[cacheKey]; found && time.Since(entry.Timestamp) < cacheTTL {
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

	queryParams := url.Values{}
	queryParams.Add("v", "5")
	queryParams.Add("type", "search")
	queryParams.Add("by", searchField)
	queryParams.Add("arg", term)

	fullURL = fmt.Sprintf("%s?v=%s&type=%s&by=%s&arg=%s", baseURL, queryParams.Get("v"), queryParams.Get("type"), queryParams.Get("by"), queryParams.Get("arg"))

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
//	if err := json.Unmarshal(data, &response); err != nil {
	if err := json.Unmarshal([]byte(data), &response); err != nil {
		fmt.Printf("Erro ao decodificar o JSON para o termo '%s': %s\n", term, err)
		return
	}

	results, ok := response["results"].([]interface{})
	if !ok {
		fmt.Printf("Campo 'results' não encontrado na resposta para o termo '%s'\n", term)
		return
	}

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
//		if err := json.Unmarshal(pkgData, &pkg); err != nil {
		if err := json.Unmarshal([]byte(pkgData), &pkg); err != nil {
			fmt.Printf("Erro ao decodificar o pacote para o termo '%s': %s\n", term, err)
			return
		}

		count++
		if verbose {
			// Preenche o campo `url` com `fullURL`
			fullURL = fmt.Sprintf("%s?v=%s&type=%s&by=%s&arg=%s", baseURL, queryParams.Get("v"), queryParams.Get("type"), queryParams.Get("by"), pkg.Name)
			pkg.fullURL = fullURL
			pkg.count = count
		}

		ch <- pkg
	}

	cacheMutex.Lock()
	cache[cacheKey] = CacheEntry{Results: nil, Timestamp: time.Now()} // Cache vazio pois já foi processado
	cacheMutex.Unlock()
}
