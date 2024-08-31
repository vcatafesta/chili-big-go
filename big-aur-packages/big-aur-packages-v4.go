package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Defina a estrutura para um item no array JSON, ajustando conforme necessário
type Package struct {
	ID             int         `json:"ID"`
	Name           string      `json:"Name"`
	PackageBaseID  int         `json:"PackageBaseID"`
	PackageBase    string      `json:"PackageBase"`
	Version        string      `json:"Version"`
	Description    string      `json:"Description"`
	URL            string      `json:"URL"`
	NumVotes       int         `json:"NumVotes"`
	Popularity     float64     `json:"Popularity"`
	OutOfDate      interface{} `json:"OutOfDate"` // Usar interface{} para lidar com múltiplos tipos
	Maintainer     string      `json:"Maintainer"`
	Submitter      string      `json:"Submitter"`
	FirstSubmitted int64       `json:"FirstSubmitted"`
	LastModified   int64       `json:"LastModified"`
	URLPath        string      `json:"URLPath"`
}

func main() {
	// URL do arquivo JSON
	url := "https://chililinux.com/packages-meta-v1.json.gz"

	// Faz o download do arquivo JSON
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Erro ao fazer a requisição:", err)
		return
	}
	defer resp.Body.Close()

	// Lê o conteúdo do arquivo JSON
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao ler o arquivo:", err)
		return
	}

	// Decodifica o JSON
	var packages []Package
	if err := json.Unmarshal(data, &packages); err != nil {
		fmt.Println("Erro ao decodificar o JSON:", err)
		return
	}

	// Prepara a lista para os primeiros 10 pacotes
	var output []Package
	for i, pkg := range packages {
		if i >= 10 {
			break
		}
		// Convertendo OutOfDate para string se necessário
		outOfDateStr := ""
		if pkg.OutOfDate != nil {
			switch v := pkg.OutOfDate.(type) {
			case float64:
				outOfDateStr = fmt.Sprintf("%v", v)
			case string:
				outOfDateStr = v
			}
		}
		// Adicionando pacote à lista de saída
		output = append(output, Package{
			ID:             pkg.ID,
			Name:           pkg.Name,
			PackageBaseID:  pkg.PackageBaseID,
			PackageBase:    pkg.PackageBase,
			Version:        pkg.Version,
			Description:    pkg.Description,
			URL:            pkg.URL,
			NumVotes:       pkg.NumVotes,
			Popularity:     pkg.Popularity,
			OutOfDate:      outOfDateStr,
			Maintainer:     pkg.Maintainer,
			Submitter:      pkg.Submitter,
			FirstSubmitted: pkg.FirstSubmitted,
			LastModified:   pkg.LastModified,
			URLPath:        pkg.URLPath,
		})
	}

	// Começa o array JSON
	fmt.Print("[\n")

	// Exibe a saída JSON formatada com um pacote por linha
	for i, pkg := range output {
		if i > 0 {
			fmt.Print(",\n")
		}
		pkgJSON, err := json.Marshal(pkg)
		if err != nil {
			fmt.Println("Erro ao formatar o JSON:", err)
			continue
		}
//		fmt.Print("  ")
		fmt.Print(string(pkgJSON))
	}

	// Fecha o array JSON
	fmt.Print("\n]\n")
}
