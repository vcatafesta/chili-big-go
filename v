#!/usr/bin/env bash
# -*- coding: utf-8 -*-
# shellcheck shell=bash disable=SC1091,SC2039,SC2166
#
#  r
#  Created: 2024/08/15 - 13:55
#  Altered: 2024/08/15 - 13:55
#
#  Copyright (c) 2024-2024, Vilmar Catafesta <vcatafesta@gmail.com>
#  All rights reserved.
#
#  Redistribution and use in source and binary forms, with or without
#  modification, are permitted provided that the following conditions
#  are met:
#  1. Redistributions of source code must retain the above copyright
#     notice, this list of conditions and the following disclaimer.
#  2. Redistributions in binary form must reproduce the above copyright
#     notice, this list of conditions and the following disclaimer in the
#     documentation and/or other materials provided with the distribution.
#
#  THIS SOFTWARE IS PROVIDED BY THE AUTHOR AS IS'' AND ANY EXPRESS OR
#  IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
#  OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
#  IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT,
#  INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
#  NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
#  DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
#  THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
#  (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
#  THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
##############################################################################
#export LANGUAGE=pt_BR
export TEXTDOMAINDIR=/usr/share/locale
export TEXTDOMAIN=r

# Definir a variável de controle para restaurar a formatação original
reset=$(tput sgr0)

# Definir os estilos de texto como variáveis
bold=$(tput bold)
underline=$(tput smul)   # Início do sublinhado
nounderline=$(tput rmul) # Fim do sublinhado
reverse=$(tput rev)      # Inverte as cores de fundo e texto

# Definir as cores ANSI como variáveis
black=$(tput bold)$(tput setaf 0)
red=$(tput bold)$(tput setaf 196)
green=$(tput bold)$(tput setaf 2)
yellow=$(tput bold)$(tput setaf 3)
blue=$(tput setaf 4)
pink=$(tput setaf 5)
magenta=$(tput setaf 5)
cyan=$(tput setaf 6)
white=$(tput setaf 7)
gray=$(tput setaf 8)
orange=$(tput setaf 202)
purple=$(tput setaf 125)
violet=$(tput setaf 61)
light_red=$(tput setaf 9)
light_green=$(tput setaf 10)
light_yellow=$(tput setaf 11)
light_blue=$(tput setaf 12)
light_magenta=$(tput setaf 13)
light_cyan=$(tput setaf 14)
bright_white=$(tput setaf 15)

#debug
export PS4='${red}${0##*/}${green}[$FUNCNAME]${pink}[$LINENO]${reset}'
#set -x
#set -e
shopt -s extglob

#system
declare APP="${0##*/}"
declare _VERSION_="1.0.0-20240815"
declare distro="$(uname -n)"
declare DEPENDENCIES=(tput)
source /usr/share/fetch/core.sh

function MostraErro {
	echo "erro: ${red}$1${reset} => comando: ${cyan}'$2'${reset} => result=${yellow}$3${reset}"
}
trap 'MostraErro "$APP[$FUNCNAME][$LINENO]" "$BASH_COMMAND" "$?"; exit 1' ERR

function by_raw() {
	# Define o separador
	separator="|"
	output=$(go run big-search-aur.go -Ss elisa-git brave-bin falkon-git --raw)
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

function by_pairs() {
	separator="="
	mapfile -t data <<<$(go run big-search-aur.go -Si elisa-git brave-bin --pairs --sep =)

	# Itera sobre o array e processa cada linha
	for entry in "${data[@]}"; do
		# Usa a expansão de parâmetros para dividir a linha pelo caractere '='
		key="${entry%%=*}"
		value="${entry#*=}"

		# Exibe as informações formatadas
		case "$key" in
		Name) echo "Name: $value" ;;
		Version) echo "Version: $value" ;;
		Description) echo "Description: $value" ;;
		Maintainer) echo "Maintainer: $value" ;;
		NumVotes) echo "NumVotes: $value" ;;
		Popularity) echo "Popularity: $value" ;;
		URL)
			echo "URL: $value"
			echo
			;;
		esac
	done
}

function by_pairs_with_array_associativo() {
	separator="="
	# Declarar o array associativo
	declare -A info

	# Usar mapfile para armazenar a saída em um array
	command_output=$(go run big-search-aur.go -Si elisa-git brave-bin --pairs --sep =)

	# Função para processar a saída e armazenar em um array associativo
	process_package() {
		local package_info=("$@")
		local package_name=""
		declare -A package_data

		for line in "${package_info[@]}"; do
			# Divida a linha em chave e valor
			key="${line%%=*}"
			value="${line#*=}"

			# Verifique se a chave e o valor são não vazios e válidos
			if [[ -n "$key" && -n "$value" ]]; then
				if [[ "$key" == "Name" ]]; then
					package_name="$value"
				fi
				package_data["$key"]="$value"
			else
				echo "Aviso: Linha inválida ou chave/valor vazio: '$line'"
			fi
		done

		# Exibir informações do pacote
		for key in "${!package_data[@]}"; do
			echo "$key: ${package_data[$key]}"
		done
	}

	# Divida a saída em blocos por pacote
	IFS=$'\n\n' read -r -d '' -a packages <<<"$command_output"

	# Processar cada pacote
	for package in "${packages[@]}"; do
		mapfile -t package_info <<<"$package"
		process_package "${package_info[@]}"
	done
}

function by_pairs_with_read() {
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
	done < <(go run big-search-aur.go -Si elisa-git brave-bin --pairs --sep =)
}

#by_raw
#by_pairs
#by_pairs_with_array_associativo
by_pairs_with_read
