#!/bin/bash

TOKEN_FILE="bws-token"

# Check if BWS_ACCESS_TOKEN is already set
if [[ -z "${BWS_ACCESS_TOKEN}" ]]; then
    # If BWS_ACCESS_TOKEN is not set, check for a token file
    if [[ -f "${TOKEN_FILE}" ]] && [[ -r "${TOKEN_FILE}" ]]; then
        # The file exists and is readable, read the token from the file
        export BWS_ACCESS_TOKEN=$(<"${TOKEN_FILE}")
    else
        # The file does not exist or is not readable, prompt user input
        read -rp "Enter your token: " TOKEN
        export BWS_ACCESS_TOKEN="$TOKEN"
    fi
fi

# Retrieve additional secrets and set corresponding environment variables
export TELEGRAM_TOKEN=$(bws secret get 7f42da1c-25ae-497a-89de-b041013fa10a | jq -r .value)
export TELEGRAM_GROUP=$(bws secret get 0a25642b-b877-4415-962b-b041013fb276 | jq -r .value)
export TELEGRAM_ALLOWEDUSERNAMES=$(bws secret get bc6f96c0-c7a9-4273-bac0-b041013fc1fa | jq -r .value)
export WEATHERAPI_KEY=$(bws secret get 627efa5d-6fdb-48bd-8956-b041013fcdab | jq -r .value)
export CURRENCYAPI_KEY=$(bws secret get 3cdeba53-2ee5-49c6-9956-b041013fe498 | jq -r .value)
export OPENAIAPI_KEY=$(bws secret get 645021f0-b0eb-41af-9aa9-b041013ff1a5 | jq -r .value)
export MINIFLUXAPI_KEY=$(bws secret get 457ec119-e932-4cbd-8ad2-b0ef00e47798 | jq -r .value)
export DEEPLAPI_KEY=$(bws secret get 0d11a328-87b9-4ade-837b-b0ef00e4baff | jq -r .value)

echo "Done"
