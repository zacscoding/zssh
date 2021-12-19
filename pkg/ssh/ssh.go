package ssh

import (
	"fmt"
	"github.com/shiena/ansicolor"
	"github.com/zacscoding/zssh/pkg/host"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"os"
)

type ClientParams struct {
	ServerInfo *host.ServerInfo
	StdIn      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
}

type Client struct {
	ServerInfo *host.ServerInfo

	conn   *ssh.Client
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// NewClient create a new Client for ssh from given params ClientParams.
func NewClient(params *ClientParams) (*Client, error) {
	sshCli, err := dial(params.ServerInfo)
	if err != nil {
		return nil, err
	}

	cli := &Client{
		conn:       sshCli,
		ServerInfo: params.ServerInfo,
		stdin:      os.Stdin,
		stdout:     os.Stdout,
		stderr:     os.Stderr,
	}

	if params.StdIn != nil {
		cli.stdin = params.StdIn
	}
	if params.Stdout != nil {
		cli.stdout = params.Stdout
	}
	if params.Stderr != nil {
		cli.stderr = params.Stderr
	}
	return cli, nil
}

// OpenShell starts the shell on the remote host in Client.
func (c *Client) OpenShell() error {
	session, err := c.conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = c.stdin
	session.Stdout = ansicolor.NewAnsiColorWriter(c.stdout)
	session.Stderr = ansicolor.NewAnsiColorWriter(c.stderr)

	// copy from http://talks.rodaine.com/gosf-ssh/present.slide#9
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,      // please print what I type
		ssh.ECHOCTL:       0,      // please don't print control chars
		ssh.TTY_OP_ISPEED: 115200, // baud in
		ssh.TTY_OP_OSPEED: 115200, // baud out
	}

	termFD := int(os.Stdin.Fd())

	width, height, err := terminal.GetSize(termFD)
	if err != nil {
		return err
	}

	termState, _ := terminal.MakeRaw(termFD)
	defer terminal.Restore(termFD, termState)

	err = session.RequestPty("xterm-256color", height, width, modes)
	if err != nil {
		return err
	}
	err = session.Shell()
	if err != nil {
		return err
	}
	err = session.Wait()
	if err != nil {
		switch err.(type) {
		case *ssh.ExitError:
			return nil
		}
		return err
	}
	return nil
}

// Run runs the given cmd on the remote host in Client.
func (c *Client) Run(cmd string) error {
	session, err := c.conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = ansicolor.NewAnsiColorWriter(c.stdout)
	session.Stderr = ansicolor.NewAnsiColorWriter(c.stderr)

	return session.Run(cmd)
}

func dial(info *host.ServerInfo) (*ssh.Client, error) {
	auth, err := newAuthMethod(info)
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: info.User,
		Auth: []ssh.AuthMethod{
			auth,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	addr := fmt.Sprintf("%s:%d", info.Address, info.Port)
	return ssh.Dial("tcp", addr, config)
}

func newAuthMethod(info *host.ServerInfo) (ssh.AuthMethod, error) {
	if info.KeyPath == "" {
		return ssh.Password(info.Password), nil
	}
	pemBytes, err := ioutil.ReadFile(info.KeyPath)
	if err != nil {
		return nil, err
	}
	var signer ssh.Signer
	if info.Password == "" {
		signer, err = ssh.ParsePrivateKey(pemBytes)
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(info.Password))
	}
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}
