/*
	big-jq-regex - Utilitario like jq para uso com AUR json https://aur.archlinux.org/packages-meta-v1.json.gz
		go get github.com/go-ini/ini
    	Chili GNU/Linux - https://github.com/vcatafesta/chili/go
    	Chili GNU/Linux - https://chililinux.com
   	Chili GNU/Linux - https://chilios.com.br

   Created: 2024/08/10
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
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	_APP_       = "big-jq-regex"
	_PKGDESC_   = "Utilitario like jq para uso com AUR json https://aur.archlinux.org/packages-meta-v1.json.gz"
	_VERSION_   = "0.13.0-20240813"
	_COPYRIGHT_ = "Copyright (C) 2024 Vilmar Catafesta, <vcatafesta@gmail.com>"

	// Constantes para cores ANSI
	Reset        = "\x1b[0m"
	Black        = "\x1b[30m"       // `tput setaf 0`
	Red          = "\x1b[31m"       // `tput setaf 196`
	Green        = "\x1b[32m"       // `tput setaf 2`
	Yellow       = "\x1b[33m"       // `tput setaf 3`
	Blue         = "\x1b[34m"       // `tput setaf 4`
	Pink         = "\x1b[35m"       // `tput setaf 5`
	Magenta      = "\x1b[35m"       // `tput setaf 5` (Pink and Magenta are the same here)
	Cyan         = "\x1b[36m"       // `tput setaf 6`
	White        = "\x1b[37m"       // `tput setaf 7`
	Gray         = "\x1b[90m"       // `tput setaf 8`
	Orange       = "\x1b[38;5;202m" // `tput setaf 202`
	Purple       = "\x1b[38;5;125m" // `tput setaf 125`
	Violet       = "\x1b[38;5;61m"  // `tput setaf 61`
	LightRed     = "\x1b[38;5;9m"   // `tput setaf 9`
	LightGreen   = "\x1b[38;5;10m"  // `tput setaf 10`
	LightYellow  = "\x1b[38;5;11m"  // `tput setaf 11`
	LightBlue    = "\x1b[38;5;12m"  // `tput setaf 12`
	LightMagenta = "\x1b[38;5;13m"  // `tput setaf 13`
	LightCyan    = "\x1b[38;5;14m"  // `tput setaf 14`
	BrightWhite  = "\x1b[97m"       // `tput setaf 15` (White is often used for Bright White)
)

// Definição da estrutura Package
type Package struct {
	ID             int     `json:"ID"`
	Name           string  `json:"Name"`
	PackageBaseID  int     `json:"PackageBaseID"`
	PackageBase    string  `json:"PackageBase"`
	Version        string  `json:"Version"`
	Description    string  `json:"Description"`
	URL            string  `json:"URL"`
	NumVotes       int     `json:"NumVotes"`
	Popularity     float64 `json:"Popularity"`
	OutOfDate      *int    `json:"OutOfDate"` // Ajustado para *int para suportar nulo
	Maintainer     string  `json:"Maintainer"`
	Submitter      string  `json:"Submitter"`
	FirstSubmitted int     `json:"FirstSubmitted"`
	LastModified   int     `json:"LastModified"`
	URLPath        string  `json:"URLPath"`
}

// Declaração da variável global
var jsonFile string
var patterns []string //slice
var limit int
var verbose bool

func main() {
	if len(os.Args) < 2 {
		printUsageAndExit()
	}

	command, showJSON, useRegex := parseArgs()
	//	jsonFile := os.Args[2]

	fmt.Printf("Comando: %s, JSON: %t, Regex: %t, Arquivo: %s, Padrões: %v, Limite: %d, Verbose: %t\n",
		command, showJSON, useRegex, jsonFile, patterns, limit, verbose)

	if err := ensureJSONFileExists(jsonFile); err != nil {
		log.Fatalf("Erro ao garantir a existência do arquivo JSON: %v\n", err)
	}

	switch command {
	case "-S", "--search":
		searchAndPrintPackage(jsonFile, patterns, showJSON, useRegex)
	case "-L", "--list":
		handleList(jsonFile, showJSON)
	case "-C", "--create":
		handleCreate(jsonFile)
	default:
		fmt.Println("Comando inválido")
		printUsageAndExit()
	}
}

// Função para análise dos argumentos da linha de comando
func parseArgs() (command string, showJSON, useRegex bool) {
	limit = -1 // Sem limite por padrão
	verbose = false

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "-L", "--list", "-C", "--create":
			command = arg
		case "-S", "--search":
			command = arg
			// Capturar padrões que vêm imediatamente após o comando
			if i+1 < len(os.Args) {
				for j := i + 1; j < len(os.Args); j++ {
					nextArg := os.Args[j]
					// Verifica se o próximo argumento é um novo parâmetro
					if nextArg[0] == '-' {
						break
					}
					patterns = append(patterns, nextArg)
					i = j // Avança o índice para pular os padrões
				}
			}
		case "-j", "--json":
			showJSON = true
		case "-r", "--regex":
			useRegex = true
		case "--verbose":
			verbose = true
		case "-V", "--version":
			fmt.Println(Red + _APP_ + " - " + _PKGDESC_ + Reset)
			fmt.Println(Cyan + "big-jq-regex - v" + _VERSION_ + Reset)
			fmt.Println("   " + _COPYRIGHT_ + Reset)
			fmt.Println("")
			fmt.Println("   Este programa pode ser redistribuído livremente")
			fmt.Println("   sob os termos da Licença Pública Geral GNU.")
			os.Exit(0)
		case "-f", "--file":
			if i+1 < len(os.Args) {
				jsonFile = os.Args[i+1]
				i++ // Pula o valor do nome do arquivo
			} else {
				fmt.Println("Falta o nome do arquivo após o parâmetro -f ou --file")
				os.Exit(1)
			}
		case "--limit":
			if i+1 < len(os.Args) {
				parsedLimit, err := strconv.Atoi(os.Args[i+1])
				if err == nil && parsedLimit > 0 {
					limit = parsedLimit
				} else {
					fmt.Println("Valor inválido para o parâmetro --limit")
					os.Exit(1)
				}
				i++ // Pular o valor de --limit
			} else {
				fmt.Println("Falta valor para o parâmetro --limit")
				os.Exit(1)
			}
		}
	}
	return command, showJSON, useRegex
}

func printUsageAndExit() {
	fmt.Println(Red + _APP_ + " - " + _PKGDESC_ + Reset)
	fmt.Println(Cyan + "big-jq-regex - v" + _VERSION_ + Reset)
	fmt.Println("   " + _COPYRIGHT_ + Reset)
	fmt.Println(Cyan + "Uso:" + Reset)
	fmt.Println("  big-jq-regex -S|--search -f <arquivo_json> <search> [<search>...] [--json] [--limit] [--verbose]")
	fmt.Println("  big-jq-regex -S|--search -f <arquivo_json> <regex_pattern> [--json] [--regex] [--limit] [--verbose]")
	fmt.Println("  big-jq-regex -L|--list   -f <arquivo_json> [--json]")
	fmt.Println("  big-jq-regex -C|--create -f <arquivo_json> <id> <name> <package_base_id> <package_base> <version> <description> <url> <num_votes> <popularity> <out_of_date> <maintainer> <submitter> <first_submitted> <last_modified> <url_path>")
	os.Exit(1)
}

func ensureJSONFileExists(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		emptyArray := []Package{}
		jsonStr, _ := json.MarshalIndent(emptyArray, "", "    ")
		return ioutil.WriteFile(filePath, jsonStr, 0644)
	}
	return nil
}

func handleList(jsonFile string, showJSON bool) {
	listPackages(jsonFile, showJSON)
}

func handleCreate(jsonFile string) {
	if len(os.Args) < 17 {
		printUsageAndExit()
	}

	idStr := os.Args[3]
	name := os.Args[4]
	packageBaseIDStr := os.Args[5]
	packageBase := os.Args[6]
	version := os.Args[7]
	description := os.Args[8]
	url := os.Args[9]
	numVotes := os.Args[10]
	popularity := os.Args[11]
	outOfDate := os.Args[12]
	maintainer := os.Args[13]
	submitter := os.Args[14]
	firstSubmitted := os.Args[15]
	lastModified := os.Args[16]
	urlPath := os.Args[17]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Fatalf("Erro ao converter ID para inteiro: %v\n", err)
	}

	packageBaseID, err := strconv.Atoi(packageBaseIDStr)
	if err != nil {
		log.Fatalf("Erro ao converter PackageBaseID para inteiro: %v\n", err)
	}

	numVotesInt, err := strconv.Atoi(numVotes)
	if err != nil {
		log.Fatalf("Erro ao converter numVotes para inteiro: %v\n", err)
	}

	popularityFloat, err := strconv.ParseFloat(popularity, 64)
	if err != nil {
		log.Fatalf("Erro ao converter popularity para float: %v\n", err)
	}

	firstSubmittedInt, err := strconv.Atoi(firstSubmitted)
	if err != nil {
		log.Fatalf("Erro ao converter firstSubmitted para inteiro: %v\n", err)
	}

	lastModifiedInt, err := strconv.Atoi(lastModified)
	if err != nil {
		log.Fatalf("Erro ao converter lastModified para inteiro: %v\n", err)
	}

	var outOfDatePtr *int
	if outOfDate != "" {
		outOfDateInt, err := strconv.Atoi(outOfDate)
		if err != nil {
			log.Fatalf("Erro ao converter outOfDate para inteiro: %v\n", err)
		}
		outOfDatePtr = &outOfDateInt
	}

	var packages []Package
	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("Erro ao ler o arquivo JSON: %v\n", err)
	}

	if err := json.Unmarshal(data, &packages); err != nil {
		log.Fatalf("Erro ao decodificar o JSON: %v\n", err)
	}

	updated := createOrUpdatePackage(&packages, id, name, packageBaseID, packageBase, version, description, url, numVotesInt, popularityFloat, outOfDatePtr, maintainer, submitter, firstSubmittedInt, lastModifiedInt, urlPath)

	if updated {
		data, err = json.MarshalIndent(packages, "", "  ")
		if err != nil {
			log.Fatalf("Erro ao codificar o JSON: %v\n", err)
		}
		err = ioutil.WriteFile(jsonFile, data, os.ModePerm)
		if err != nil {
			log.Fatalf("Erro ao escrever no arquivo JSON: %v\n", err)
		}
		if verbose {
			log.Printf("%s %sSET: %s'%d'%s no arquivo %s - 200 OK%s\n", _APP_, Green, Yellow, id, Cyan, jsonFile, Reset)
		}
	} else {
		log.Printf("Nada a ser atualizado no arquivo JSON: %s\n", jsonFile)
	}
}

func searchAndPrintPackage(jsonFile string, patterns []string, showJSON, useRegex bool) {
	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("Erro ao ler o arquivo JSON: %v\n", err)
	}

	var packages []Package
	if err := json.Unmarshal(data, &packages); err != nil {
		log.Fatalf("Erro ao decodificar o JSON: %v\n", err)
	}

	packageMap := make(map[int]Package)
	for _, pkg := range packages {
		packageMap[pkg.ID] = pkg
	}

	if useRegex {
		searchWithRegex(jsonFile, packageMap, patterns[0], showJSON)
	} else {
		searchWithPatterns(jsonFile, packageMap, patterns, showJSON)
	}
}

func searchWithRegex(jsonFile string, data map[int]Package, pattern string, showJSON bool) {
	count := 0
	regex, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatalf("Erro ao compilar o regex: %v\n", err)
	}

	// Lista para armazenar os nomes dos pacotes encontrados
	var foundNames []string
	var foundAny bool

	// Canal para comunicação entre goroutines
	results := make(chan Package)
	done := make(chan struct{})

	// Função para processar pacotes
	go func() {
		defer close(done) // Garante que o canal 'done' será fechado quando a função terminar
		for _, pkg := range data {
			if limit > 0 && count >= limit {
				break
			}
			if regex.MatchString(pkg.Name) || regex.MatchString(pkg.Description) {
				foundNames = append(foundNames, pkg.Name)
				foundAny = true
				count++
				if verbose {
					log.Printf("%s %sGET: %s'%s'%s em %s %s- 200 OK%s\n", _APP_, Green, Yellow, strings.TrimSpace(pkg.Name), Cyan, jsonFile, Green, Reset)
				}
				if showJSON {
					printJSON(pkg)
				} else {
					printPackage(pkg)
				}
			}
		}
	}()

	// Iniciar goroutines para processar cada pacote
	for _, pkg := range data {
		go func(pkg Package) {
			results <- pkg
		}(pkg)
	}

	// Fechar o canal de resultados após todas as goroutines terminarem
	go func() {
		for i := 0; i < len(data); i++ {
			// Aguarda que todas as goroutines terminem
			<-done
		}
		close(results)
	}()

	// Aguardar que todas as goroutines terminem
	<-done

	if !foundAny {
		// Se nenhum pacote for encontrado, imprime a mensagem de erro
		if verbose {
			log.Printf("%s %sGET: %s'%s'%s em %s %s- 404 NOK%s\n", _APP_, Red, Yellow, pattern, Cyan, jsonFile, Red, Reset)
		}
	}
}

func searchWithPatterns(jsonFile string, data map[int]Package, patterns []string, showJSON bool) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	count := 0
	isMatchFound := false

	results := make(chan Package)
	done := make(chan struct{})

	// Função para processar pacotes em paralelo
	processPackages := func() {
		defer close(done) // Fecha o canal 'done' quando terminar
		for pkg := range results {
			isMatchFound = false
			for _, pattern := range patterns {
				if limit > 0 && count >= limit {
					return
				}
				if strings.Contains(pkg.Name, pattern) || strings.Contains(pkg.Description, pattern) {
					mu.Lock()
					count++
					isMatchFound = true
					mu.Unlock()
					break
				}
			}
			if isMatchFound {
				if verbose {
					log.Printf("%s %sGET: %s'%s'%s em %s %s- 200 OK%s\n", _APP_, Green, Yellow, strings.TrimSpace(pkg.Name), Cyan, jsonFile, Green, Reset)
				}
				if showJSON {
					printJSON(pkg)
				} else {
					printPackage(pkg)
				}
			}
		}
	}

	// Goroutine para processar pacotes
	wg.Add(1)
	go func() {
		defer wg.Done()
		processPackages()
	}()

	// Função para enviar pacotes para o canal de resultados
	sendPackages := func() {
		for _, pkg := range data {
			results <- pkg
		}
		close(results)
	}

	// Iniciar goroutine para enviar pacotes
	wg.Add(1)
	go func() {
		defer wg.Done()
		sendPackages()
	}()

	// Aguardar o término de todas as goroutines
	wg.Wait()
	<-done

	if !isMatchFound {
		if verbose {
			log.Printf("%s %sGET: %s'%s'%s em %s %s- 404 NOK%s\n", _APP_, Red, Yellow, strings.Join(patterns, ", "), Cyan, jsonFile, Red, Reset)
		}
	}
}

func printPackage(pkg Package) {
	fmt.Printf("ID: %d\n", pkg.ID)
	fmt.Printf("Name: %s\n", pkg.Name)
	fmt.Printf("PackageBaseID: %d\n", pkg.PackageBaseID)
	fmt.Printf("PackageBase: %s\n", pkg.PackageBase)
	fmt.Printf("Version: %s\n", pkg.Version)
	fmt.Printf("Description: %s\n", pkg.Description)
	fmt.Printf("URL: %s\n", pkg.URL)
	fmt.Printf("NumVotes: %d\n", pkg.NumVotes)
	fmt.Printf("Popularity: %f\n", pkg.Popularity)
	if pkg.OutOfDate != nil {
		fmt.Printf("OutOfDate: %d\n", *pkg.OutOfDate)
	} else {
		fmt.Printf("OutOfDate: NULL\n")
	}
	fmt.Printf("Maintainer: %s\n", pkg.Maintainer)
	fmt.Printf("Submitter: %s\n", pkg.Submitter)
	fmt.Printf("FirstSubmitted: %d\n", pkg.FirstSubmitted)
	fmt.Printf("LastModified: %d\n", pkg.LastModified)
	fmt.Printf("URLPath: %s\n", pkg.URLPath)
	fmt.Println()
}

func printJSON(pkg Package) {
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		log.Fatalf("Erro ao codificar o JSON: %v\n", err)
	}
	fmt.Println(string(data))
}

func createOrUpdatePackage(packages *[]Package, id int, name string, packageBaseID int, packageBase string, version string, description string, url string, numVotes int, popularity float64, outOfDate *int, maintainer string, submitter string, firstSubmitted int, lastModified int, urlPath string) bool {
	updated := false
	for i, pkg := range *packages {
		if pkg.ID == id {
			(*packages)[i] = Package{
				ID:             id,
				Name:           name,
				PackageBaseID:  packageBaseID,
				PackageBase:    packageBase,
				Version:        version,
				Description:    description,
				URL:            url,
				NumVotes:       numVotes,
				Popularity:     popularity,
				OutOfDate:      outOfDate,
				Maintainer:     maintainer,
				Submitter:      submitter,
				FirstSubmitted: firstSubmitted,
				LastModified:   lastModified,
				URLPath:        urlPath,
			}
			updated = true
			break
		}
	}

	if !updated {
		*packages = append(*packages, Package{
			ID:             id,
			Name:           name,
			PackageBaseID:  packageBaseID,
			PackageBase:    packageBase,
			Version:        version,
			Description:    description,
			URL:            url,
			NumVotes:       numVotes,
			Popularity:     popularity,
			OutOfDate:      outOfDate,
			Maintainer:     maintainer,
			Submitter:      submitter,
			FirstSubmitted: firstSubmitted,
			LastModified:   lastModified,
			URLPath:        urlPath,
		})
	}

	return updated
}

func listPackages(jsonFile string, showJSON bool) {
	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("Erro ao ler o arquivo JSON: %v\n", err)
	}

	var packages []Package
	if err := json.Unmarshal(data, &packages); err != nil {
		log.Fatalf("Erro ao decodificar o JSON: %v\n", err)
	}

	for _, pkg := range packages {
		if showJSON {
			printJSON(pkg)
		} else {
			printPackage(pkg)
		}
	}
}
