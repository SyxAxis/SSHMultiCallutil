package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

type SSHScriptConfig struct {
	SSHScrCfgName           string   `json:"scriptName"`
	SSHScrCfgHost           string   `json:"hostname"`
	SSHScrCfgUserID         string   `json:"userid"`
	SSHScrCfgPrivateKeyFile string   `json:"privatekeyfile"`
	SSHScrCfgScriptContent  []string `json:"sshscriptcontent"`
}

var sshExecutionConfig SSHScriptConfig

func main() {

	sshCmdCfgFile := flag.String("scriptfile", "configTest01.json", "blah blah")
	flag.Parse()

	InitProcess(*sshCmdCfgFile)

}

func InitProcess(sshCmdCfgFile string) {

	ReadJSONConfigFile(sshCmdCfgFile, &sshExecutionConfig)

	log.Printf("%q\n", sshExecutionConfig)

	sshConn, err := OpenSSHConnection(&sshExecutionConfig)
	if err != nil {
		log.Fatalln("Connection failed.")
	}

	defer func() {
		sshConn.Close()
		log.Println("Disconnected.")
	}()

	RunCommands(&sshExecutionConfig, sshConn)
}

func CheckError(err error, msg string) {
	if err != nil {
		log.Fatal("FAILED: ", msg)
		log.Fatal("ERROR : ", err)
	}
}

func OpenSSHConnection(sshExecutionConfig *SSHScriptConfig) (*ssh.Client, error) {
	pemBytes, err := ioutil.ReadFile(sshExecutionConfig.SSHScrCfgPrivateKeyFile)
	if err != nil {
		log.Println("Failed to read PPK file!")
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		log.Println("Failed to parse PPK data!")
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: sshExecutionConfig.SSHScrCfgUserID,
		// Auth: []ssh.AuthMethod{ ssh.Password("password"),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshConn, err := ssh.Dial("tcp", sshExecutionConfig.SSHScrCfgHost, config)
	if err != nil {
		log.Println("Failed to connect!")
		return nil, err
	}
	log.Println("Connected...")

	return sshConn, nil

}

func ReadJSONConfigFile(cfgFilename string, sshExecutionConfig *SSHScriptConfig) {
	// Error: "invalid character 'ÿ' looking for beginning of value"
	// Issue: The text file has not been encoded with UTF8
	//        Often happens with raw Windows text files
	// Fix  : Use Powershell cmd [  cat sourcefile.json | Out-File -FilePath "targetfile.json" -Encoding "UTF8"  ]
	log.Printf("JSON cfgfile : %v\n", cfgFilename)
	rawJson, err := ioutil.ReadFile(cfgFilename)

	// JSON specs state you can simply ignore the BOM ( Byte Order Marker )
	rawJSONByte := bytes.TrimPrefix(rawJson, []byte("\xef\xbb\xbf")) // Or []byte{239, 187, 191}
	CheckError(err, "Unable to convert JSON file content to struct.")

	err = json.Unmarshal(rawJSONByte, &sshExecutionConfig)
	CheckError(err, "Unable to exec json.Unmarshal")

}

func RunCommands(sshExecutionConfig *SSHScriptConfig, conn *ssh.Client) {
	// func RunCommands(sshExecutionConfig *SSHScriptConfig) {

	log.Printf("============== OUTPUT BEGIN - [%v] ==================\n", sshExecutionConfig.SSHScrCfgName)
	for _, tmpCmd := range sshExecutionConfig.SSHScrCfgScriptContent {
		log.Printf("============== COMMAND - [%v] ==================\n", tmpCmd)

		sess, err := conn.NewSession()
		if err != nil {
			panic(err)
		}
		defer sess.Close()

		sessStdOut, err := sess.StdoutPipe()
		if err != nil {
			panic(err)
		}
		go io.Copy(os.Stdout, sessStdOut)

		sessStderr, err := sess.StderrPipe()
		if err != nil {
			panic(err)
		}
		go io.Copy(os.Stderr, sessStderr)

		err = sess.Run(tmpCmd) // eg., /usr/bin/whoami
		if err != nil {
			log.Println(err)
		}

	}
	log.Printf("============== OUTPUT END   - [%v] ==================\n", sshExecutionConfig.SSHScrCfgName)
}
