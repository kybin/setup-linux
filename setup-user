#!/bin/bash

# setup은 마지막에 실행됨
function setup {
    setupGo
    setupKeep
}

function appendIfNotExist {
	grep -q "$1" $2 || echo "$1" >> $2
}

function setupGo {
    echo "setting up go"
    appendIfNotExist 'export GOPATH=$HOME' $HOME/.bashrc
    appendIfNotExist 'export PATH=$GOPATH/bin:$PATH' $HOME/.bashrc
}

function setupKeep {
    echo "setting up keep"
    appendIfNotExist 'export KEEP_GITHUB_USER=' $HOME/.bashrc
    appendIfNotExist 'export KEEP_GITHUB_AUTH=' $HOME/.bashrc
}

setup
