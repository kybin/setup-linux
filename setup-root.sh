#!/bin/bash

# setup은 마지막에 실행됨
function setup {
    if [ ! "$(id -u)" -eq 0 ]; then
        echo "please run as root"
        return
    fi

    setupGo
    setupTor
    setupKeep
    setupRipgrep
}

function appendIfNotExist {
	grep -q '$1' $2 || echo '$1' >> $2
}

function setupGo {
    echo "setting up go"
    if [ -d /usr/local/go ]; then
        wget -nc -q https://dl.google.com/go/go1.13.5.linux-amd64.tar.gz
        tar -C /usr/local -zxf go1.13.5.linux-amd64.tar.gz
        appendIfNotExist 'export PATH=/usr/local/go/bin:$PATH' /etc/profile
        appendIfNotExist 'export GOPATH=$HOME' $HOME/.bashrc
        appendIfNotExist 'export PATH=$GOPATH/bin:$PATH' $HOME/.bashrc
        rm go1.13.5.linux-amd64.tar.gz
    fi
}


function setupTor {
    echo "setting up tor"
    git clone https://github.com/kybin/tor
    cd tor
    go build -o /usr/local/bin
    cd $OLDPWD
    rm -rf tor
}

function setupKeep {
    echo "setting up keep"
    git clone https://github.com/lazypic/keep
    cd keep
    go build -o /usr/local/bin
    cd $OLDPWD
    rm -rf keep
}

function setupRipgrep {
    echo "setting up rg"
    wget -nc -q https://github.com/BurntSushi/ripgrep/releases/download/11.0.2/ripgrep-11.0.2-i686-unknown-linux-musl.tar.gz
    tar -zxf ripgrep-11.0.2-i686-unknown-linux-musl.tar.gz ripgrep-11.0.2-i686-unknown-linux-musl/rg
    mv ripgrep-11.0.2-i686-unknown-linux-musl/rg /usr/local/bin
    rm -r ripgrep-11.0.2-i686-unknown-linux-musl
    rm ripgrep-11.0.2-i686-unknown-linux-musl.tar.gz
}

setup
