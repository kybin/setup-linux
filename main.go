package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

// userConfig는 .userrc에 들어갈 내용이다.
var userConfig = `
# ps1
parse_git_branch() {
    git branch 2> /dev/null | sed -e '/^[^*]/d' -e 's/* \(.*\)/ (\1)/'
}
export PS1="\u@\h \[\033[32m\]\w\[\033[33m\]\$(parse_git_branch)\[\033[00m\] $ "

# editor
export EDITOR=tor

# jd
function jd {
    cd -P "$HOME/.jd/$1"
}
function _jd {
    COMPREPLY=()
    if [ -d $HOME/.jd ]; then
        local dirs=$(ls $HOME/.jd)
        local cur="${COMP_WORDS[COMP_CWORD]}"
        COMPREPLY=($(compgen -W "$dirs" -- $cur))
        return 0
    fi
}
complete -F _jd jd

# go
export GOPATH=$HOME
export PATH=$GOPATH/bin:$PATH

# keep
export KEEP_GITHUB_USER="kybin"
export KEEP_GITHUB_AUTH=
`

// appendIfNotExist는 해당 파일에 필요한 줄이 없을때 추가해준다.
func appendIfNotExist(fname string, s string) error {
	if s == "" {
		return fmt.Errorf("empty string is invalid")
	}
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	// 파일의 마지막 줄이 비어있는지 검사하기 위해
	// line을 밖에서 선언함.
	line := ""
	find := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line = sc.Text()
		if line == s {
			find = true
			break
		}
	}
	if sc.Err() != nil {
		return err
	}
	if find {
		return nil
	}
	w := bufio.NewWriter(f)
	defer w.Flush()
	if line != "" {
		s = "\n" + s
	}
	_, err = w.WriteString(s + "\n")
	if err != nil {
		return err
	}
	return nil
}

// Runner는 Run을 실행하고 그 에러를 반환한다.
type Runner interface {
	Run() error
}

// download는 특정 URI에 있는 파일을 내 디스크에 다운로드 받아주는 Runner이다.
type download struct {
	from string
	to   string
	perm os.FileMode
}

func Download(from string, to string, perm os.FileMode) download {
	return download{from: from, to: to, perm: perm}
}

func (d download) Run() error {
	_, err := os.Stat(d.to)
	if err == nil {
		return fmt.Errorf("download destination '" + d.to + "' exists. remove it first.")
	} else if !os.IsNotExist(err) {
		return err
	}
	resp, err := http.Get(d.from)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.OpenFile(d.to, os.O_CREATE|os.O_WRONLY, d.perm)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

// command는 특정 디렉토리에서 명령을 실행하는 Runner이다.
type command struct {
	dir string
	cmd *exec.Cmd
}

func Command(dir string, cmd *exec.Cmd) command {
	return command{dir: dir, cmd: cmd}
}

func (c command) Run() error {
	c.cmd.Stdout = os.Stdout
	c.cmd.Stderr = os.Stderr
	c.cmd.Dir = c.dir
	return c.cmd.Run()
}

// installGo는 go를 설치한다.
func installGo() error {
	fmt.Println("setting up go")
	_, err := os.Stat("/usr/local/go")
	if err == nil {
		fmt.Println("'/usr/local/go' exists. skip.")
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	runners := []Runner{
		Download("https://dl.google.com/go/go1.13.7.linux-amd64.tar.gz", "go.tar.gz", 0644),
		Command("", exec.Command("tar", "-C", "/usr/local", "-zxf", "go.tar.gz")),
		Command("", exec.Command("rm", "go.tar.gz")),
	}
	for _, r := range runners {
		err := r.Run()
		if err != nil {
			return err
		}
	}
	appendIfNotExist("/etc/profile.d/go.sh", "export PATH=/usr/local/go/bin:$PATH")
	return nil
}

// installGoimports는 go 코드에 사용된 모듈이 빠져있을때 자동으로 불러주는 goimport를 설치한다.
func installGoimports() error {
	fmt.Println("setting up goimports")
	_, err := exec.LookPath("goimports")
	if err == nil {
		fmt.Println("'goimports' exist. skip.")
		return nil
	}
	if !errors.Is(err, exec.ErrNotFound) {
		return err
	}
	runners := []Runner{
		Download("https://github.com/kybin/goimports/releases/download/tip/goimports", "/usr/local/bin/goimports", 0755),
	}
	for _, r := range runners {
		err := r.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

// installTor는 내 에디터 tor를 설치한다.
func installTor() error {
	fmt.Println("setting up tor")
	_, err := exec.LookPath("tor")
	if err == nil {
		fmt.Println("'tor' exist. skip.")
		return nil
	}
	if !errors.Is(err, exec.ErrNotFound) {
		return err
	}
	runners := []Runner{
		Download("https://github.com/kybin/tor/releases/download/tip/tor", "/usr/local/bin/tor", 0755),
	}
	for _, r := range runners {
		err := r.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

// installKeep은 레포지터리를 KEEPPATH 아래에 다운로드 받는 keep을 설치한다.
func installKeep() error {
	fmt.Println("setting up keep")
	_, err := exec.LookPath("keep")
	if err == nil {
		fmt.Println("'keep' exist. skip.")
		return nil
	}
	if !errors.Is(err, exec.ErrNotFound) {
		return err
	}
	runners := []Runner{
		Download("https://github.com/lazypic/keep/releases/download/tip/keep", "/usr/local/bin/keep", 0755),
	}
	for _, r := range runners {
		err := r.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

// installRipgrep은 grep과 비슷하지만 더 나은 rg를 설치한다.
func installRipgrep() error {
	fmt.Println("setting up rg")
	_, err := exec.LookPath("rg")
	if err == nil {
		fmt.Println("'rg' exist. skip.")
		return nil
	}
	if !errors.Is(err, exec.ErrNotFound) {
		return err
	}
	runners := []Runner{
		Download("https://github.com/BurntSushi/ripgrep/releases/download/11.0.2/ripgrep-11.0.2-i686-unknown-linux-musl.tar.gz", "rg.tar.gz", 0644),
		Command("", exec.Command("tar", "-zxf", "rg.tar.gz", "ripgrep-11.0.2-i686-unknown-linux-musl/rg", "--strip-components", "1")),
		Command("", exec.Command("mv", "rg", "/usr/local/bin/rg")),
		Command("", exec.Command("rm", "rg.tar.gz")),
	}
	for _, r := range runners {
		err := r.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

// setupGit은 항상 사용하는 git 설정을 지정한다.
func setupGit() error {
	fmt.Println("setting up git")
	_, err := exec.LookPath("git")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Println("'git' not exist. skip.")
			return nil
		}
		return err
	}
	runners := []Runner{
		Command("", exec.Command("git", "config", "--global", "user.email", "kybinz@gmail.com")),
		Command("", exec.Command("git", "config", "--global", "user.name", "kim yongbin")),
	}
	for _, r := range runners {
		err := r.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

// setupUserrc는 홈 디렉토리에 .userrc를 설치한다.
// .userrc는 .bashrc에서 부르게 된다.
func setupUserrc() error {
	fmt.Println("setting up .userrc")
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	userrc := home + "/.userrc"
	_, err = os.Stat(userrc)
	if err == nil {
		fmt.Println("'" + userrc + "' exists. skip.")
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	err = os.MkdirAll(home+"/.jd", 0755)
	if err != nil {
		return err
	}
	err = appendIfNotExist(userrc, userConfig)
	if err != nil {
		return err
	}
	err = appendIfNotExist(home+"/.bashrc", "source $HOME/.userrc")
	if err != nil {
		return err
	}
	return nil
}

// die는 에러를 출력하고 프로그램을 종료한다.
func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// setupRoot는 루트 계정에서 필요한 셋업이다.
func setupRoot() error {
	funcs := []func() error{
		installGo,
		installGoimports,
		installTor,
		installKeep,
		installRipgrep,
	}
	for _, fn := range funcs {
		err := fn()
		if err != nil {
			return err
		}
	}
	return nil
}

// setupUser는 루트를 포함한 모든 계정에서 필요한 셋업이다.
func setupUser() error {
	funcs := []func() error{
		setupGit,
		setupUserrc,
	}
	for _, fn := range funcs {
		err := fn()
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	uid := os.Geteuid()
	if uid == 0 {
		// user is root
		err := setupRoot()
		if err != nil {
			die(err)
		}
	}
	err := setupUser()
	if err != nil {
		die(err)
	}
	fmt.Println("setup for linux ended successfully. please logout and login again.")
}
