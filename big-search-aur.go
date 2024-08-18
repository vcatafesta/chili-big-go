/*
  big-search-aur - Command-line AUR helper
    go get github.com/go-ini/ini
    Chili GNU/Linux - https://github.com/vcatafesta/chili/go
    Chili GNU/Linux - https://chililinux.com
    Chili GNU/Linux - https://chilios.com.br

  Created: 2024/08/13
  Altered: 2024/08/17 - 02h22

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
	_VERSION_   = "0.17.0222-20240817"
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
	NumVotes    int     `json:"NumVotes"`
	Popularity  float64 `json:"Popularity"`
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
var nlenArgs int

var searchField string
var searchTerms []string
var outputFormat string
var separator string = "|"
var limit int = -1 // Usar -1 para indicar que não há limite
var searchMode string
var args []string
var optionToField map[string]string

// Inline
var msgError = func(msg string) { fmt.Println(Red + msg + Reset) }
var echo = func(args ...interface{}) { fmt.Println(args...) }
var logError = func(args ...interface{}) { log.Println(Red + fmt.Sprint(args...) + Reset) }

func main() {
	args = os.Args[1:]
	nlenArgs = len(args)
	if nlenArgs < 1 {
		printUsage()
		return
	}

	if parseArgs() {
		if searchMode == "search" && len(searchTerms) == 0 {
			echo(Red + "Erro: Nenhuma palavra-chave de busca fornecida" + Reset)
			return
		}
		runSearchPackages()
	}
}

func parseArgs() bool {
	optionToField = map[string]string{
		"--by-name":         "name",
		"--by-name-desc":    "name-desc",
		"--by-maintainer":   "maintainer",
		"--by-depends":      "depends",
		"--by-makedepends":  "makedepends",
		"--by-optdepends":   "optdepends",
		"--by-checkdepends": "checkdepends",
	}

	for i := 0; i < nlenArgs; i++ {
		//		logError("args[", i, "]", args[i])
		switch args[i] {
		case "-Ss", "--search":
			searchMode = "search"
		case "-Si", "--info":
			searchMode = "info"
		case "--json", "--raw", "--pairs":
			outputFormat = args[i]
		case "--sep":
			if i+1 < nlenArgs && !strings.HasPrefix(args[i+1], "--") {
				separator = args[i+1]
				i++
			} else {
				logError("Erro: --sep requer um argumento válido.")
				return false
			}
		case "--verbose":
			verbose = true
		case "--help":
			printUsage()
		case "--bash":
			helpBash()
    case "-V", "--version":
      fmt.Println(Red + _APP_ + " - " + _PKGDESC_ + Reset)
      fmt.Println(Cyan + _APP_ + " - v" + _VERSION_ + Reset)
      fmt.Println("   " + _COPYRIGHT_ + Reset)
      fmt.Println("")
      fmt.Println("   Este programa pode ser redistribuído livremente")
      fmt.Println("   sob os termos da Licença Pública Geral GNU.")
      os.Exit(0)
		case "--limit":
			if i+1 < nlenArgs {
				parsedLimit, err := strconv.Atoi(args[i+1])
				if err != nil || parsedLimit < 1 {
					logError("Erro: --limit requer um número positivo")
					return false
				}
				limit = parsedLimit
				i++
			} else {
				logError("Erro: --limit requer um argumento")
				return false
			}
		default:
			if field, ok := optionToField[args[i]]; ok {
				searchField = field
			} else {
				searchTerms = append(searchTerms, args[i])
			}
		}
	}
	return true
}

func helpBash() {
	text := `
#!/usr/bin/env bash
# -*- coding: utf-8 -*-

by_pairs_with_read() {
  # Processar cada linha e atribuir valores às variáveis
  while read -r line; do
    # Substitui "=" por "_=" e avalia a linha
    eval "${line//=/_=}"

    # Se a linha estiver vazia, pula para a próxima iteração
    if [[ -n "$line" ]]; then
      echo "Name: $Name_"
      echo "Version: $Version_"
      echo "Description: $Description_"
      echo "Maintainer: $Maintainer_"
      echo "NumVotes: $NumVotes_"
      echo "Popularity: $Popularity_"
      echo "URL: $URL_"
      echo ""
    fi
    unset Name_ Version_ Description_ Maintainer_ NumVotes_ Popularity_ URL_
  done < <(big-search-aur -Si elisa-git brave-bin --pairs --sep '=')
}

by_raw_with_mapfile_read() {
  # Define o separador
  separator="|"
  output=$(big-search-aur -Ss elisa-git brave-bin falkon-git --raw)
  mapfile -t packages <<<"$output"
  # Itera sobre cada linha (pacote)
  for package_line in "${packages[@]}"; do
    IFS="$separator" read -r name version description maintainer num_votes popularity url full_url count <<<"$package_line"
    echo "Linha do pacote: $package_line"

    # Exibe as informações do pacote
    echo "Name: $name"
    echo "Version: $version"
    echo "Description: $description"
    echo "Maintainer: $maintainer"
    echo "NumVotes: $num_votes"
    echo "Popularity: $popularity"
    echo "URL: $url"
    echo "fullURL: $full_url"
    echo "Count: $count"
    echo "###############################################################################################################"
  done
}

#by_raw_with_mapfile_read() {
by_pairs_with_read() {

`
	fmt.Println(Cyan + text + Reset)
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
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --raw", reset, "Saída formatada como texto simples com todos os campos (util para usar com mapfile/read do bash)", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --pairs", reset, "Usa o formato de saída texto chave='valor' (util para usar com mapfile/read do bash)", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --sep", reset, "Separador dos campos na saída raw (padrão é '=')", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --limit", reset, "Limite de pacotes encontrados", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --verbose", reset, "Liga modo verboso", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --bash", reset, "Mostra exemplo de uso com bash", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --version", reset, "Mostra a versão do aplicativo", reset)
	fmt.Printf("%s%-20s %s%s%s\n", blue, "  --help", reset, "Este help", reset)
}

func runSearchPackages() {
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
		jsonData, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			fmt.Println("Erro ao formatar saída JSON:", err)
			return
		}
		fmt.Println(string(jsonData))
	} else if outputFormat == "--pairs" {
		separator = "="
		for _, pkg := range results {
			fmt.Printf("Name%s'%s' Version%s'%s' Description%s'%s' Maintainer%s'%s' NumVotes%s'%d' Popularity%s'%.2f' URL%s'%s'\n",
				separator, pkg.Name, separator, pkg.Version, separator, pkg.Description,
				separator, pkg.Maintainer, separator, pkg.NumVotes, separator, pkg.Popularity,
				separator, pkg.URL)
		}
	} else {
		for _, pkg := range results {
			echo(pkg.Name + separator +
				pkg.Version + separator +
				pkg.Description + separator +
				pkg.Maintainer + separator +
				strconv.Itoa(pkg.NumVotes) + separator +
				strconv.FormatFloat(pkg.Popularity, 'f', -1, 64) + separator +
				pkg.URL + separator +
				strconv.Itoa(pkg.count))
		}
	}
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

		count++
		if verbose {
			// Preenche o campo `url` com `fullURL`
			fullURL = fmt.Sprintf("%s?v=%s&type=%s&arg[]=%s", baseURL, queryParams.Get("v"), queryParams.Get("type"), pkg.Name)
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

func getStringField(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		switch v := value.(type) {
		case string:
			return v
		case float64:
			return fmt.Sprintf("%f", v)
		case bool:
			return fmt.Sprintf("%t", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
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

	// Verificar se algum parâmetro 'by-' foi fornecido
	if queryParams.Get("by") == "" {
		fullURL = fmt.Sprintf("%s?v=%s&type=%s&arg=%s", baseURL, queryParams.Get("v"), queryParams.Get("type"), queryParams.Get("arg"))
	} else {
		fullURL = fmt.Sprintf("%s?v=%s&type=%s&by=%s&arg=%s", baseURL, queryParams.Get("v"), queryParams.Get("type"), queryParams.Get("by"), queryParams.Get("arg"))
	}

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

		count++
		if verbose {
			// Verificar se algum parâmetro 'by-' foi fornecido
			if queryParams.Get("by") == "" {
				fullURL = fmt.Sprintf("%s?v=%s&type=%s&arg=%s", baseURL, queryParams.Get("v"), queryParams.Get("type"), pkg.Name)
			} else {
				fullURL = fmt.Sprintf("%s?v=%s&type=%s&by=%s&arg=%s", baseURL, queryParams.Get("v"), queryParams.Get("type"), queryParams.Get("by"), pkg.Name)
			}
			pkg.fullURL = fullURL
			pkg.count = count
		}

		ch <- pkg
	}

	cacheMutex.Lock()
	cache[cacheKey] = CacheEntry{Results: nil, Timestamp: time.Now()} // Cache vazio pois já foi processado
	cacheMutex.Unlock()
}
