/*
 *	  big-pacman-to-json - process output pacman to json
 *	  go get github.com/go-ini/ini
 *    Chili GNU/Linux - https://github.com/vcatafesta/chili/go
 *    Chili GNU/Linux - https://chililinux.com
 *    Chili GNU/Linux - https://chilios.com.br
 *
 *    Created: 2023/10/01
 *    Altered: 2024/08/12
 *
 *    Copyright (c) 2023-2023, Vilmar Catafesta <vcatafesta@gmail.com>
 *    All rights reserved.
 *
 *    Redistribution and use in source and binary forms, with or without
 *    modification, are permitted provided that the following conditions
 *    are met:
 *    1. Redistributions of source code must retain the above copyright
 *        notice, this list of conditions and the following disclaimer.
 *    2. Redistributions in binary form must reproduce the above copyright
 *        notice, this list of conditions and the following disclaimer in the
 *        documentation and/or other materials provided with the distribution.
 *
 *    THIS SOFTWARE IS PROVIDED BY THE AUTHOR ``AS IS'' AND ANY EXPRESS OR
 *    IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
 *    OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
 *    IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT,
 *    INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
 *    NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 *    DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 *    THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 *    (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
 *    THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	_APP_     = "big-pacman-to-json"
	_VERSION_ = "0.8.0-20240812"
	_COPY_    = "Copyright (C) 2023 Vilmar Catafesta, <vcatafesta@gmail.com>"
)

// Constantes para cores ANSI
const (
	Reset   = "\x1b[0m"
	Red     = "\x1b[31m"
	Green   = "\x1b[32m"
	Yellow  = "\x1b[33m"
	Blue    = "\x1b[34m"
	Magenta = "\x1b[35m"
	Cyan    = "\x1b[36m"
	White   = "\x1b[37m"
)

type PackageInfoSearch struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Size        string `json:"size"`
	Status      string `json:"status"`
	Repo        string `json:"Repo"`
	Description string `json:"description"`
}

type PackageInfo struct {
	Repository    string   `json:"Repository"`
	Name          string   `json:"Name"`
	Version       string   `json:"Version"`
	Description   string   `json:"Description"`
	Architecture  string   `json:"Architecture"`
	URL           string   `json:"URL"`
	Licenses      []string `json:"Licenses"`
	Groups        string   `json:"Groups"`
	Provides      string   `json:"Provides"`
	DependsOn     []string `json:"DependsOn"`
	OptionalDeps  []string `json:"OptionalDeps"`
	RequiredBy    []string `json:"RequiredBy"`
	ConflictsWith string   `json:"ConflictsWith"`
	Replaces      string   `json:"Replaces"`
	DownloadSize  string   `json:"DownloadSize"`
	InstalledSize string   `json:"InstalledSize"`
	Packager      string   `json:"Packager"`
	BuildDate     string   `json:"BuildDate"`
	MD5Sum        string   `json:"MD5Sum"`
	SHA256Sum     string   `json:"SHA256Sum"`
	Signatures    string   `json:"Signatures"`
}

type PackageData struct {
	Name    string      `json:"Name"`
	Package PackageInfo `json:"Package"`
}

// Estrutura para armazenar os campos encontrados em uma linha
type LineFields struct {
	Name        string
	Version     string
	Size        string
	Status      string
	Repo        string
	Description string
}

var (
	Advanced bool = false
)

func main() {
	var input string

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input += scanner.Text() + "\n"
		}
		xcmd := "paru"
		// Processa a entrada
		ProcessOutputSearch(input, xcmd)
	} else if len(os.Args) > 1 {
		// Percorre os argumentos usando um loop for
		for _, arg := range os.Args {
			// Testa se ten argumento igual a "-Sii" ou "-Si"
			if arg == "-Sii" {
				Advanced = true
			} else if arg == "-Si" {
				Advanced = true
			} else if arg == "-V" || arg == "--version" {
				fmt.Printf("%s v%s\n", _APP_, _VERSION_)
				fmt.Printf("%s\n", _COPY_)
				os.Exit(0)
			} else if arg == "-h" || arg == "--help" {
				usage(true)
			}
		}

		if len(os.Args) >= 3 {
			// Os argumentos a partir do segundo são as entradas, processa-os
			args := os.Args[1:]
			xcmd := os.Args[1]
			cmd := exec.Command(args[0], args[1:]...) // Executa o comando com os argumentos
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("%sErro ao executar o comando: %s'%s' - %s%v%s\n", Red, Cyan, os.Args[1:], Yellow, err, Reset)
				return
			}
			input = string(output)
			if Advanced {
				// Chame a função ProcessOutput com a saída do comando como argumento
				ProcessOutput(input)
			} else {
				// Chame a função ProcessOutputSearch com a saída do comando como argumento
				ProcessOutputSearch(input, xcmd)
			}
		} else {
			usage(false)
		}
	} else {
		log.Printf("%sErro: nenhuma operação/entrada especificada (use -h para obter ajuda)%s", Red, Reset)
		os.Exit(1)
	}
}

func usage(IsValidParameter bool) {
	boolToInt := func(value bool) int {
		if value {
			return 0
		}
		return 1
	}

	if IsValidParameter == false {
		log.Printf("%sErro: Parâmetro(s) inválido(s) %s'%s'%s\n", Red, Cyan, os.Args[1:], Reset)
	}
	fmt.Printf("Uso:%s %s%s <comandos>%s\n", Red, _APP_, Cyan, Reset)
	fmt.Printf("%scomandos:%s\n", Cyan, Reset)
	fmt.Printf("     %s -h|--help\n", _APP_)
	fmt.Printf("     %s -v|--version\n", _APP_)
	fmt.Printf("%s     %s pacman %s-Ss [<pacote>] [<regex>]%s\n", Yellow, _APP_, Cyan, Reset)
	fmt.Printf("%s     %s pacman %s-Qm%s\n", Yellow, _APP_, Cyan, Reset)
	fmt.Printf("%s     %s pacman %s-Qn%s\n", Yellow, _APP_, Cyan, Reset)
	fmt.Printf("%s     %s pacman %s-Si [<pacote> [<...>]%s\n", Yellow, _APP_, Cyan, Reset)
	fmt.Printf("%s     %s pacman %s-Sii [<pacote> [<...>]%s\n", Yellow, _APP_, Cyan, Reset)
	fmt.Printf("%s     %s paru %s-Ss [<pacote> [<...>]%s\n", Yellow, _APP_, Cyan, Reset)
	fmt.Printf("%s     %s paru %s-Ssa [<pacote> [<...>]%s\n", Yellow, _APP_, Cyan, Reset)
	fmt.Printf("%s     %s yay %s-Ss [<pacote> [<...>]%s\n", Yellow, _APP_, Cyan, Reset)
	fmt.Printf("%s     %s yay %s-Sii [<pacote> [<...>]%s\n", Yellow, _APP_, Cyan, Reset)
	fmt.Printf("%s     %s pamac %s search [<pacote> [<...>]%s\n", Yellow, _APP_, Cyan, Reset)
	os.Exit(boolToInt(IsValidParameter))
}

func ProcessOutput(output string) {
	var packageInfos = make(map[string]PackageInfo) // Inicialize o mapa

	lines := strings.Split(output, "\n")
	var currentPackage PackageInfo

	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "Repository":
				currentPackage.Repository = value
			case "Name":
				currentPackage.Name = value
			case "Version":
				currentPackage.Version = value
			case "Description":
				currentPackage.Description = value
			case "Architecture":
				currentPackage.Architecture = value
			case "URL":
				currentPackage.URL = value
			case "Licenses":
				licenses := strings.Split(value, " ")
				currentPackage.Licenses = append(currentPackage.Licenses, licenses...)
			case "Groups":
				currentPackage.Groups = value
			case "Provides":
				currentPackage.Provides = value
			case "Depends On":
				dependencies := strings.Fields(value)
				currentPackage.DependsOn = append(currentPackage.DependsOn, dependencies...)
			case "Optional Deps":
				optionalDeps := strings.Fields(value)
				currentPackage.OptionalDeps = append(currentPackage.OptionalDeps, optionalDeps...)
			case "Required By":
				requiredBy := strings.Fields(value)
				currentPackage.RequiredBy = append(currentPackage.RequiredBy, requiredBy...)
			case "Conflicts With":
				currentPackage.ConflictsWith = value
			case "Replaces":
				currentPackage.Replaces = value
			case "Download Size":
				currentPackage.DownloadSize = value
			case "Installed Size":
				currentPackage.InstalledSize = value
			case "Packager":
				currentPackage.Packager = value
			case "Build Date":
				currentPackage.BuildDate = value
			case "MD5 Sum":
				currentPackage.MD5Sum = value
			case "SHA-256 Sum":
				currentPackage.SHA256Sum = value
			case "Signatures":
				currentPackage.Signatures = value
			}
		} else if len(parts) == 1 && len(parts[0]) == 0 && currentPackage.Name != "" {
			packageInfos[currentPackage.Name] = currentPackage
			currentPackage = PackageInfo{}
		}
	}

	// Salva no arquivo
	outputFilename := "/tmp/" + _APP_ + ".json"
	file, err := os.Create(outputFilename)
	if err != nil {
		log.Printf("%sErro ao criar arquivo JSON: %v%s\n", Red, err, Reset)
		os.Exit(1)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(packageInfos); err != nil {
		log.Printf("%sErro ao escrever no arquivo JSON: %v%s\n", Red, err, Reset)
		os.Exit(1)
	}

	// Converte a lista de pacotes em formato JSON
	jsonData, err := json.Marshal(packageInfos)
	if err != nil {
		log.Printf("%sErro ao serializar para JSON: %v%s\n", Red, err, Reset)
		os.Exit(1)
	}

	// Imprime os dados JSON na saída padrão
	fmt.Println(string(jsonData))
	os.Exit(0)
}

// xdebug exibe uma mensagem e/ou valores e aguarda a continuação
func xdebug(message string, values ...interface{}) {
	// Se a mensagem for vazia, define um valor padrão
	if message == "" {
		message = "Tecle algo..."
	}

	// Exibe os valores, se fornecidos
	if len(values) > 0 {
		for _, value := range values {
			fmt.Printf("%v ", value)
		}
	}
	//		fmt.Println("\n",message)
	var input string
	fmt.Scanln(&input)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Função para verificar se a linha começa com dois ou mais espaços em branco
func startsWithTwoOrMoreSpaces(line string) bool {
	if len(line) < 2 {
		return false
	}
	return strings.HasPrefix(line[:2], "  ")
}

// ProcessOutputSearch processa a saída da busca e a converte em uma lista de pacotes.
func ProcessOutputSearch(input, xcmd string) error {
	var packages []PackageInfoSearch
	var currentPackage PackageInfoSearch
	lines := strings.Split(input, "\n")
	isDescription := false
	isName := false

	for _, line := range lines {
		// Verifica se a linha tem mais de dois espaços iniciais e se estamos processando uma descrição.
		if startsWithTwoOrMoreSpaces(line) && isName {
			line = strings.TrimSpace(line)
			currentPackage.Description += line
			isDescription = true
		} else if isName {
			isDescription = true
		} else {
			// Processa a linha atual para preencher o pacote.
			fields := processLine(line, xcmd, &currentPackage)
			if len(fields) > 0 {
				isName = true
			}
		}

		// Se ambos nome e descrição foram processados, adiciona o pacote à lista.
		if isDescription && isName {
			packages = append(packages, currentPackage)
			currentPackage = PackageInfoSearch{}
			isDescription = false
			isName = false
		}
	}

	// Adiciona o último pacote se ele tiver um nome.
	if currentPackage.Name != "" {
		packages = append(packages, currentPackage)
	}

	//  return outputString(packages)  // Converte a lista de pacotes para uma string e imprime.
	return outputJSON(packages) // Converte a lista de pacotes para JSON e imprime.
}

// outputJSON converte a lista de pacotes para JSON e imprime.
func outputJSON(packages []PackageInfoSearch) error {
	jsonData, err := json.Marshal(packages)
	if err != nil {
		return fmt.Errorf("erro ao serializar para JSON: %v", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// outputString converte a lista de pacotes para uma string e imprime.
func outputString(packages []PackageInfoSearch) error {
	for _, pkg := range packages {
		fmt.Printf("Name: %s\nVersion: %s\nSize: %s\nStatus: %s\nRepo: %s\nDescription: %s\n\n",
			pkg.Name, pkg.Version, pkg.Size, pkg.Status, pkg.Repo, pkg.Description)
	}
	return nil
}

func processLine(line, xcmd string, currentPackage *PackageInfoSearch) []string {
	fields := strings.Fields(line)
	nlen := len(fields)
	f := LineFields{}

	// Verifica se a primeira palavra contém uma barra e a processa.
	if nlen > 0 && strings.Contains(fields[0], "/") {
		parts := strings.SplitN(fields[0], "/", 2)
		f.Repo = parts[0]
		f.Name = parts[1]
	} else if nlen > 0 {
		f.Repo = ""
		f.Name = fields[0]
	}

	// Atribui Name e Version diretamente dos campos fields.
	if nlen > 1 {
		currentPackage.Name = f.Name
		currentPackage.Version = fields[1]
		currentPackage.Repo = f.Repo // Atribui o repositório do pacote.
	}

	// Processamento específico para cada comando.
	switch xcmd {
	case "pacman":
		if nlen > 2 {
			currentPackage.Status = fields[nlen-1]
		}

	case "paru":
		// Identifica o campo de tamanho (entre colchetes).
		if nlen > 2 && strings.HasPrefix(fields[2], "[") && strings.HasSuffix(fields[2], "]") {
			currentPackage.Size = fields[2]
		}
		if nlen > 3 {
			currentPackage.Status = fields[3]
		}

	case "pamac":
		if nlen > 2 {
			currentPackage.Status = fields[2]
		}
		if nlen > 3 {
			currentPackage.Repo = fields[3]
		}

	case "yay":
		// O campo Size está entre parênteses.
		for i := 2; i < nlen; i++ {
			if strings.HasPrefix(fields[i], "(") && strings.HasSuffix(fields[i], ")") {
				currentPackage.Size = fields[i]
				break
			}
		}
		if nlen > 2 {
			currentPackage.Status = fields[nlen-1]
		}

	default:
		if nlen > 2 {
			currentPackage.Status = fields[nlen-1]
		}
	}

	return fields
}
